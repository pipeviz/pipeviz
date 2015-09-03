package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	gjs "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
)

const (
	FileReadFail = 1 << iota
	ValidationFail
	ValidationError
)

// TODO need a whole build toolchain around this
const schemaRaw = `
{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "description": "Complete schema for pipeviz messages.",
    "type": "object",
    "properties": {
        "environments": {
            "type": "array",
            "minItems": 1,
            "items": { "$ref": "#/definitions/environment" }
        },
        "logic-states": {
            "type": "array",
            "minItems": 1,
            "items": { "$ref": "#/definitions/logic-state" }
        },
        "datasets": {
            "type": "array",
            "minItems": 1,
            "items": { "$ref": "#/definitions/dataset" }
        },
        "processes": {
            "type": "array",
            "minItems": 1,
            "items": { "$ref": "#/definitions/process" }
        },
        "commits": {
            "type": "array",
            "minItems": 1,
            "items": { "$ref": "#/definitions/commit" }
        },
        "commit-meta": {
            "type": "array",
            "minItems": 1,
            "items": { "$ref": "#/definitions/commit-meta" }
        },
        "yum-pkg": {
            "type": "array",
            "minItems": 1,
            "items": { "$ref": "#definitions/yum-pkg" }
        }
    },
    "additionalProperties": false,
    "definitions": {
        "environment": {
            "title": "Environment",
            "description": "A representation of an environment - physical, virtual, or container.",
            "type": "object",
            "properties": {
                "type": {
                    "enum": [ "physical", "virtual", "container" ],
                    "default": "virtual"
                },
                "os": {
                    "enum": [ "windows", "linux", "darwin", "freebsd", "unix" ],
                    "default": "unix"
                },
                "address": {
                    "$ref": "#/definitions/address"
                },
                "nick": {
                    "type": "string",
                    "description": "A nickname identifying this environment. Nicknames can be used as referents for defining the hierarchical relationship between an environment and its contents, but not for real network/addressable relationships. Need not correspond to any real state. Nicknames co-exist in a global namespace for all environments known to any pipeviz instance, so pick them wisely."
                },
                "provider": {
                    "type": "string"
                },
                "logic-states": {
                    "type": "array",
                    "minItems": 1,
                    "items": { "$ref": "#/definitions/logic-state" }
                },
                "processes": {
                    "type": "array",
                    "minItems": 1,
                    "items": { "$ref": "#/definitions/process" }
                },
                "datasets": {
                    "type": "array",
                    "minItems": 1,
                    "items": { "$ref": "#/definitions/dataset" }
                }
            },
            "additionalProperties": false,
            "anyOf": [
                { "required": [ "nick" ] },
                { "required": [ "address" ] }
            ]
        },
        "address": {
            "type": "object",
            "properties": {
                "hostname": {
                    "type": "string",
                    "format": "host-name"
                },
                "ipv4": {
                    "type": "string",
                    "format": "ipv4"
                },
                "ipv6": {
                    "type": "string",
                    "format": "ipv6"
                }
            },
            "anyOf": [
                { "required": [ "hostname" ] },
                { "required": [ "ipv4" ] },
                { "required": [ "ipv6" ] }
            ],
            "additionalProperties": false
        },
        "env-link": {
            "type": "object",
            "description": "Describes a link back from a thing to the environment that contains it. 'Contains' in as physical a sense as possible. This approach needs a lot of thought.",
            "properties": {
                "address": { "$ref": "#/definitions/address" }
            },
            "required": [ "address" ],
            "additionalProperties": false
        },
        "logic-state": {
            "type": "object",
            "properties": {
                "type": {
                    "enum": [ "binary", "code", "library" ]
                },
                "path": { "type": "string" },
                "nick": { "type": "string" },
                "lgroup": { "type": "string" },
                "environment": { "$ref": "#/definitions/env-link" },
                "libraries": {
                    "type": "array",
                    "minItems": 1,
                    "items": { "type": "string" }
                },
                "id": {
                    "type": "object",
                    "oneOf": [
                        {
                            "properties": {
                                "commit": { "type" : "string" }
                            },
                            "required": [ "commit" ],
                            "additionalProperties": false
                        },
                        {
                            "properties": {
                                "version": { "type": "string" }
                            },
                            "required": [ "version" ],
                            "additionalProperties": false
                        },
                        {
                            "properties": {
                                "semver": { "type": "string" }
                            },
                            "required": [ "semver" ],
                            "additionalProperties": false
                        }
                    ]
                },
                "datasets": {
                    "type": "array",
                    "minItems": 1,
                    "items": { "$ref": "#/definitions/conn-data" }
                }
            },
            "required": [ "path", "id" ],
            "additionalProperties": false
        },
        "conn-data": {
            "type": "object",
            "oneOf": [
                {
                    "properties": {
                        "name": { "type": "string" },
                        "type": { "enum": [ "mediated" ] },
                        "subset": { "type": "string" },
                        "interaction": {
                            "enum": [ "rw", "ro" ]
                        },
                        "connNet": { "$ref": "#/definitions/conn-net" }
                    },
                    "required": [ "interaction", "type", "connNet" ],
                    "additionalProperties": false
                },
                {
                    "properties": {
                        "name": { "type": "string" },
                        "type": { "enum": [ "mediated" ] },
                        "subset": { "type": "string" },
                        "interaction": {
                            "enum": [ "rw", "ro" ]
                        },
                        "connUnix": { "$ref": "#/definitions/conn-unix" }
                    },
                    "required": [ "interaction", "type", "connUnix" ],
                    "additionalProperties": false
                },
                {
                    "properties": {
                        "name": { "type": "string" },
                        "type": { "enum": [ "file" ] },
                        "subset": { "type": "string" },
                        "interaction": {
                            "enum": [ "rw", "ro" ]
                        },
                        "connUnix": { "$ref": "#/definitions/conn-unix" }
                    },
                    "required": [ "interaction", "type", "connUnix" ],
                    "additionalProperties": false
                }
            ]
        },
        "conn-net": {
            "type": "object",
            "description": "Descriptor of an outgoing network connection; a target/address",
            "oneOf": [
                {
                    "properties": {
                        "hostname": {
                            "type": "string",
                            "format": "host-name"
                        },
                        "port": { "type": "integer" },
                        "proto": { "type": "string" }
                    },
                    "required": [ "hostname", "port", "proto" ],
                    "additionalProperties": false
                },
                {
                    "properties": {
                        "ipv4": {
                            "type": "string",
                            "format": "ipv4"
                        },
                        "port": { "type": "integer" },
                        "proto": { "type": "string" }
                    },
                    "required": [ "ipv4", "port", "proto" ],
                    "additionalProperties": false
                },
                {
                    "properties": {
                        "ipv6": {
                            "type": "string",
                            "format": "ipv6"
                        },
                        "port": { "type": "integer" },
                        "proto": { "type": "string" }
                    },
                    "required": [ "ipv6", "port", "proto" ],
                    "additionalProperties": false
                }
            ]
        },
        "conn-unix": {
            "type": "object",
            "description": "Descriptor of an outgoing local unix connection; a target/path",
            "properties": {
                "path": { "type": "string" }
            }
        },
        "process": {
            "type": "object",
            "properties": {
                "logic-states": {
                    "type": "array",
                    "minItems": 1,
                    "items": { "type": "string" }
                },
                "user": { "type": "string" },
                "environment": { "$ref": "#/definitions/env-link" },
                "group": { "type": "string" },
                "cwd": { "type": "string" },
                "dataset": { "type": "string" },
                "pid": { "type": "integer" },
                "listen": {
                    "type": "array",
                    "minItems": 1,
                    "items": { "$ref": "#/definitions/addr-listen" }
                }
            },
            "required": [ "logic-states", "pid" ],
            "additionalProperties": false
        },
        "addr-listen": {
            "type": "object",
            "description": "Describes the listener side of an inter-process connection. Currently just ports and sockets, lots of love needed.",
            "oneOf": [
                {
                    "properties": {
                        "type": { "enum": [ "port" ] },
                        "port": { "type": "integer" },
                        "proto": {
                            "type": "array",
                            "minItems": 1,
                            "items": { "type": "string" }
                        }
                    },
                    "required": [ "type", "port", "proto" ],
                    "additionalProperties": false
                },
                {
                    "properties": {
                        "type": { "enum": [ "unix" ] },
                        "path": { "type": "string" }
                    },
                    "required": [ "type", "path" ],
                    "additionalProperties": false
                }
            ]
        },
        "dataset": {
            "type": "object",
            "description": "A dataset. Think of the term really, really broadly. variant 1 is a self-sourced dataset; variant 2 is a functioning dataset; variant 3 is a dataset container",
            "oneOf": [
                {
                    "properties": {
                        "name": { "type": "string" },
                        "create-time": { "format": "date-time" },
                        "genesis": { "enum": [ "α" ] }
                    },
                    "required": [ "name", "genesis" ],
                    "additionalProperties": false
                },
                {
                    "properties": {
                        "name": { "type": "string" },
                        "create-time": { "format": "date-time" },
                        "genesis": {
                            "type": "object",
                            "properties": {
                                "address": { "$ref": "#/definitions/address" },
                                "dataset": {
                                    "type": "array",
                                    "minItems": 1,
                                    "items": { "type": "string" }
                                },
                                "snap-time": { "format": "date-time" }
                            },
                            "required": [ "address", "dataset", "snap-time" ],
                            "additionalProperties": false
                        }
                    },
                    "required": [ "name", "genesis" ],
                    "additionalProperties": false
                },
                {
                    "properties": {
                        "name": { "type": "string" },
                        "environment": { "$ref": "#/definitions/env-link" },
                        "path": { "type": "string" },
                        "subsets": {
                            "type": "array",
                            "minItems": 1,
                            "items": { "$ref": "#/definitions/dataset" }
                        }
                    },
                    "required": [ "name" ],
                    "additionalProperties": false
                }
            ]
        },
        "commit": {
            "type": "object",
            "description": "Describes a source control commit object. For now, just git. Probably for always, only SCMs that have atomic commits/a DAG.",
            "properties": {
                "sha1": { "type": "string" },
                "repository": { "type": "string" },
                "date": { "type": "string" },
                "author": { "type": "string" },
                "subject": { "type": "string" },
                "parents": {
                    "type": "array",
                    "minItems": 0,
                    "items": { "type": "string" }
                }
            },
            "required": [ "sha1", "date", "author", "subject", "parents", "repository" ],
            "additionalProperties": false
        },
        "commit-meta": {
            "type": "object",
            "description": "Describes metadata about a commit - branches or tags associated with it, testing states, etc.",
            "properties": {
                "sha1": { "type": "string" },
                "testState": {
                    "enum": [ "passed", "pending", "failed" ]
                },
                "tags": {
                    "type": "array",
                    "minItems": 1,
                    "items": { "type": "string" }
                },
                "branches": {
                    "type": "array",
                    "minItems": 1,
                    "items": { "type": "string" }
                }
            },
            "required": [ "sha1" ],
            "additionalProperties": false
        },
        "yum-pkg": {
            "type": "object",
            "description": "A package, as understood by the yum package manager used in rpm-based Linux distributions.",
            "properties": {
                "name": { "type": "string" },
                "version": { "type": "string" },
                "epoch": { "type": "integer" },
                "release": { "type": "string" },
                "arch": { "type": "string" }
            },
            "required": [ "name", "version", "release", "epoch", "arch" ],
            "additionalProperties": false
        }
    }
}
`

func validateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <dir>...",
		Short: "Reads JSON message fixtures from a directory and validates them against the master schema.",
		Long:  `Given one or more directories containing a set of JSON pipeviz message fixtures, validates each of those messages against the master schema and reports any failures.`,
		Run:   runValidate,
	}

	return cmd
}

func runValidate(cmd *cobra.Command, args []string) {
	schema, err := gjs.NewSchema(gjs.NewStringLoader(schemaRaw))
	if err != nil {
		panic("bad schema...?")
	}

	var errors int
	for _, dir := range args {
		fl, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Printf("Failed to read directory '%v' with error %v\n", dir, err)
		}

		for _, f := range fl {
			if match, _ := regexp.MatchString("\\.json$", f.Name()); match && !f.IsDir() {
				src, err := ioutil.ReadFile(dir + "/" + f.Name())
				if err != nil {
					errors |= FileReadFail
					fmt.Printf("Failed to read fixture file %v/%v\n", dir, f.Name())
					continue
				}

				result, err := schema.Validate(gjs.NewStringLoader(string(src)))
				if err != nil {
					errors |= ValidationError
					fmt.Printf("Validation process terminated with errors for %v/%v. Error: \n%v\n", dir, f.Name(), err.Error())
					continue
				}

				if !result.Valid() {
					errors |= ValidationFail
					fmt.Printf("Errors in %v/%v:\n", dir, f.Name())
					for _, desc := range result.Errors() {
						fmt.Printf("\t%s\n", desc)
					}
				} else {
					fmt.Printf("%v/%v successfully validated\n", dir, f.Name())
				}
			}
		}
	}
	os.Exit(errors)
}

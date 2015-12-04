//go:generate go-bindata -nometadata -pkg $GOPACKAGE -o bindata_schema.go schema.json
package schema

import (
	"github.com/Sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

var schema *gojsonschema.Schema

func init() {
	raw, err := Asset("schema.json")
	if err != nil {
		// It is correct to fatal out here because there is no use case for importing
		// the schema package that does not require the master schema to be present
		logrus.WithFields(logrus.Fields{
			"system": "schema",
			"err":    err,
		}).Fatal("Failed to locate raw bytes/file for master schema")
	}

	schema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(string(raw)))
	if err != nil {
		// Correct to fatal, again for the same reasons
		logrus.WithFields(logrus.Fields{
			"system": "schema",
			"err":    err,
		}).Fatal("Failed to create master schema object")
	}
}

// Master returns the master pipeviz JSON schema as a gojsonschema *Schema.
func Master() *gojsonschema.Schema {
	return schema
}

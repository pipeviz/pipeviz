package interpret

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	log "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/represent/types"
)

// Not actually used right now, but this interface must be satisfied by all types
type unifier interface {
	UnificationForm(uint64) []types.UnifyInstructionForm
}

type Message struct {
	m *message
}

type message struct {
	Env []Environment   `json:"environments"`
	Ls  []LogicState    `json:"logic-states"`
	Pds []ParentDataset `json:"datasets"`
	Ds  []Dataset
	P   []Process    `json:"processes"`
	C   []Commit     `json:"commits"`
	Cm  []CommitMeta `json:"commit-meta"`
	Yp  []PkgYum     `json:"yum-pkg"`
}

// UnmarshalJSON implements the json.Unmarshaler interface. It translates a
// JSON message into a series of discrete objects that can then be merged
// into the graph.
func (m *Message) UnmarshalJSON(data []byte) error {
	logEntry := log.WithFields(log.Fields{
		"system": "interpet",
	})

	m.m = new(message)
	err := json.Unmarshal(data, m.m)
	if err != nil {
		logEntry.WithField("err", err).Info("Error while unmarshaling message JSON")
	}

	// TODO separate all of this into pluggable/generated structures
	// first, dump all top-level objects into the graph.
	for _, e := range m.m.Env {
		envlink := EnvLink{Address: Address{}}
		// Create an envlink for any nested items, preferring nick, then hostname, ipv4, ipv6.
		if e.Nick != "" {
			envlink.Nick = e.Nick
		} else if e.Address.Hostname != "" {
			envlink.Address.Hostname = e.Address.Hostname
		} else if e.Address.Ipv4 != "" {
			envlink.Address.Ipv4 = e.Address.Ipv4
		} else if e.Address.Ipv6 != "" {
			envlink.Address.Ipv6 = e.Address.Ipv6
		}

		// manage the little environment hierarchy
		for _, ls := range e.LogicStates {
			ls.Environment = envlink
			m.m.Ls = append(m.m.Ls, ls)
		}
		for _, p := range e.Processes {
			p.Environment = envlink
			m.m.P = append(m.m.P, p)
		}
		for _, pds := range e.Datasets {
			pds.Environment = envlink
			m.m.Pds = append(m.m.Pds, pds)
		}
	}

	return nil
}

// UnificationForm translates all data in the message into the standard
// UnifyInstructionForm, suitable for merging into the dataset.
func (m Message) UnificationForm(id uint64) []types.UnifyInstructionForm {
	logEntry := log.WithFields(log.Fields{
		"system": "interpet",
		"msgid":  id,
	})

	ret := make([]types.UnifyInstructionForm, 0)

	for _, e := range m.m.Env {
		logEntry.WithField("vtype", "environment").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}
	for _, e := range m.m.Ls {
		if e.ID.CommitStr != "" {
			byts, err := hex.DecodeString(e.ID.CommitStr)
			if err != nil {
				log.WithFields(log.Fields{
					"system":    "interpet",
					"semantics": true, // TODO have some constants/errtypes for this
				}).Warn("Invalid input: logic state's referenced commit sha1 was not hex encoded")
			}
			copy(e.ID.Commit[:], byts[0:20])
			e.ID.CommitStr = ""
		}

		logEntry.WithField("vtype", "logic state").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}
	for _, e := range m.m.Pds {
		logEntry.WithField("vtype", "parent dataset").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}
	for _, e := range m.m.Ds {
		logEntry.WithField("vtype", "dataset").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}
	for _, e := range m.m.P {
		logEntry.WithField("vtype", "process").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}
	for _, e := range m.m.C {
		byts, err := hex.DecodeString(e.Sha1Str)
		if err != nil {
			log.WithFields(log.Fields{
				"system":    "interpet",
				"semantics": true, // TODO have some constants/errtypes for this
			}).Warn("Invalid input: commit sha1 was not hex encoded. Skipping item")
			continue
		}
		copy(e.Sha1[:], byts[0:20])
		e.Sha1Str = ""

		for _, pstr := range e.ParentsStr {
			byts, err := hex.DecodeString(pstr)
			if err != nil {
				log.WithFields(log.Fields{
					"system":    "interpet",
					"semantics": true, // TODO have some constants/errtypes for this
				}).Warn("Invalid input: commit parent sha1 was not hex encoded. Skipping parent")
				continue
			}
			var sha1 Sha1
			copy(sha1[:], byts[0:20])
			e.Parents = append(e.Parents, sha1)
		}
		e.ParentsStr = nil

		logEntry.WithField("vtype", "git commit").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}
	for _, e := range m.m.Cm {
		byts, err := hex.DecodeString(e.Sha1Str)
		if err != nil {
			log.WithFields(log.Fields{
				"system":    "interpet",
				"semantics": true, // TODO have some constants/errtypes for this
			}).Warn("Invalid input: commit meta's referenced commit sha1 was not hex encoded. Skipping item")
			continue
		}
		copy(e.Sha1[:], byts[0:20])
		e.Sha1Str = ""

		logEntry.WithField("vtype", "commit meta").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}

	for _, e := range m.m.Yp {
		logEntry.WithField("vtype", "yum-pkg").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}

	return ret
}

// Unmarshalers for variant subtypes

// Unmarshaling a Dataset involves resolving whether it has α genesis (string), or
// a provenancial one (struct). So we have to decode directly, here.
func (ds *Dataset) UnmarshalJSON(data []byte) (err error) {
	type αDataset struct {
		Name       string    `json:"name"`
		CreateTime string    `json:"create-time"`
		Genesis    DataAlpha `json:"genesis"`
	}
	type provDataset struct {
		Name       string         `json:"name"`
		CreateTime string         `json:"create-time"`
		Genesis    DataProvenance `json:"genesis"`
	}

	a, b := αDataset{}, provDataset{}
	// use α first, as that can match the case where it's not specified (though schema
	// currently does not allow that)
	if err = json.Unmarshal(data, &a); err == nil {
		ds.Name, ds.CreateTime, ds.Genesis = a.Name, a.CreateTime, a.Genesis
	} else if err = json.Unmarshal(data, &b); err == nil {
		ds.Name, ds.CreateTime, ds.Genesis = b.Name, b.CreateTime, b.Genesis
	} else {
		err = errors.New("JSON genesis did not match either alpha or provenancial forms.")
	}

	return err
}

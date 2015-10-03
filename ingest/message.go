package ingest

import (
	"encoding/json"

	log "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/types/semantic"
	"github.com/tag1consulting/pipeviz/types/system"
)

// Not actually used right now, but this interface must be satisfied by all types
type unifier interface {
	UnificationForm(uint64) []system.UnifyInstructionForm
}

type Message struct {
	m *message
}

type message struct {
	Env []semantic.Environment   `json:"environments"`
	Ls  []semantic.LogicState    `json:"logic-states"`
	Pds []semantic.ParentDataset `json:"datasets"`
	Ds  []semantic.Dataset
	P   []semantic.Process    `json:"processes"`
	C   []semantic.Commit     `json:"commits"`
	Cm  []semantic.CommitMeta `json:"commit-meta"`
	Yp  []semantic.PkgYum     `json:"yum-pkg"`
}

// UnmarshalJSON implements the json.Unmarshaler interface. It translates a
// JSON message into a series of discrete objects that can then be merged
// into the graph.
func (m *Message) UnmarshalJSON(data []byte) error {
	m.m = new(message)
	err := json.Unmarshal(data, m.m)
	if err != nil {
		log.WithFields(log.Fields{
			"system": "interpet",
			"err":    err,
		}).Info("Error while unmarshaling message JSON")
	}

	return nil
}

// UnificationForm translates all data in the message into the standard
// UnifyInstructionForm, suitable for merging into the dataset.
func (m Message) UnificationForm(id uint64) []system.UnifyInstructionForm {
	logEntry := log.WithFields(log.Fields{
		"system": "interpet",
		"msgid":  id,
	})

	ret := make([]system.UnifyInstructionForm, 0)

	for _, e := range m.m.Env {
		logEntry.WithField("vtype", "environment").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}
	for _, e := range m.m.Ls {
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
		logEntry.WithField("vtype", "git commit").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}
	for _, e := range m.m.Cm {
		logEntry.WithField("vtype", "commit meta").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}

	for _, e := range m.m.Yp {
		logEntry.WithField("vtype", "yum-pkg").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm(id)...)
	}

	return ret
}

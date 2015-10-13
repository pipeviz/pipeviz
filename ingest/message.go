package ingest

import (
	"errors"

	log "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/types/semantic"
	"github.com/tag1consulting/pipeviz/types/system"
)

// Not actually used right now, but this interface must be satisfied by all types
type Message struct {
	Env []semantic.Environment   `json:"environments"`
	Ls  []semantic.LogicState    `json:"logic-states"`
	Pds []semantic.ParentDataset `json:"datasets"`
	Ds  []semantic.Dataset
	P   []semantic.Process    `json:"processes"`
	C   []semantic.Commit     `json:"commits"`
	Cm  []semantic.CommitMeta `json:"commit-meta"`
	Yp  []semantic.PkgYum     `json:"yum-pkg"`
}

// UnificationForm translates all data in the message into the standard
// UnifyInstructionForm, suitable for merging into the dataset.
func (m Message) UnificationForm() []system.UnifyInstructionForm {
	logEntry := log.WithFields(log.Fields{
		"system": "interpet",
	})

	ret := make([]system.UnifyInstructionForm, 0)

	for _, e := range m.Env {
		logEntry.WithField("vtype", "environment").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm()...)
	}
	for _, e := range m.Ls {
		logEntry.WithField("vtype", "logic state").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm()...)
	}
	for _, e := range m.Pds {
		logEntry.WithField("vtype", "parent dataset").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm()...)
	}
	for _, e := range m.Ds {
		logEntry.WithField("vtype", "dataset").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm()...)
	}
	for _, e := range m.P {
		logEntry.WithField("vtype", "process").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm()...)
	}
	for _, e := range m.C {
		logEntry.WithField("vtype", "git commit").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm()...)
	}
	for _, e := range m.Cm {
		logEntry.WithField("vtype", "commit meta").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm()...)
	}

	for _, e := range m.Yp {
		logEntry.WithField("vtype", "yum-pkg").Debug("Preparing to translate into UnifyInstructionForm")
		ret = append(ret, e.UnificationForm()...)
	}

	return ret
}

// Add adds an object to the data in the message.
//
// The semantics here are generally additive - if it makes sense for one object to be merged into
// another rather than directly be added, that may happen. Otherwise, duplication may happen,
// though duplication is, by contract, not an issue because of pipeviz' guarantee of idempotent
// interpretation.
//
// This is only intended for use when assembling a message to be sent (from a message producer),
// not when reading a real one.
func (m *Message) Add(d system.Unifier) error {
	switch obj := d.(type) {
	case semantic.Environment:
		if m.Env == nil {
			m.Env = make([]semantic.Environment, 0)
		}
		m.Env = append(m.Env, obj)
	case semantic.LogicState:
		if m.Ls == nil {
			m.Ls = make([]semantic.LogicState, 0)
		}
		m.Ls = append(m.Ls, obj)
	case semantic.ParentDataset:
		if m.Pds == nil {
			m.Pds = make([]semantic.ParentDataset, 0)
		}
		m.Pds = append(m.Pds, obj)
	case semantic.Dataset:
		return errors.New("not supported yet bc datasets are a mess")
	case semantic.Process:
		if m.P == nil {
			m.P = make([]semantic.Process, 0)
		}
		m.P = append(m.P, obj)
	case semantic.Commit:
		if m.C == nil {
			m.C = make([]semantic.Commit, 0)
		}
		m.C = append(m.C, obj)
	case semantic.CommitMeta:
		if m.Cm == nil {
			m.Cm = make([]semantic.CommitMeta, 0)
		}
		m.Cm = append(m.Cm, obj)
	case semantic.PkgYum:
		if m.Yp == nil {
			m.Yp = make([]semantic.PkgYum, 0)
		}
		m.Yp = append(m.Yp, obj)
	}

	return errors.New("type not supported")
}

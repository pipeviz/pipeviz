package ingest

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/pipeviz/pipeviz/types/semantic"
	"github.com/pipeviz/pipeviz/types/system"
)

// Message represents a valid pipeviz message. It is, currently, intended for use both
// in encoding messages to be sent to a pipeviz server by clients, and decoding messages
// by the server once they have arrived.
//
// This symmetry will not necessarily always be the case. As message definitions become
// more compositional over time, the encode/decode cases will likely diverge.
//
// The decode/serverside case, however, will always respond to the UnificationForm()
// method, as that is how the message contents are translated to their next interpreted form.
type Message struct {
	Env []semantic.Environment   `json:"environments,omitempty"`
	Ls  []semantic.LogicState    `json:"logic-states,omitempty"`
	Pds []semantic.ParentDataset `json:"datasets,omitempty"`
	Ds  []semantic.Dataset       `json:"-"`
	P   []semantic.Process       `json:"processes,omitempty"`
	C   []semantic.Commit        `json:"commits,omitempty"`
	Cm  []semantic.CommitMeta    `json:"commit-meta,omitempty"`
	Yp  []semantic.PkgYum        `json:"yum-pkg,omitempty"`
}

// UnificationForm translates all data in the message into the standard
// UnifyInstructionForm, suitable for merging into the dataset.
func (m Message) UnificationForm() []system.UnifyInstructionForm {
	logEntry := log.WithFields(log.Fields{
		"system": "interpet",
	})

	var ret []system.UnifyInstructionForm

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
		// These share an ident - dedupe them if possible
		if m.Cm == nil {
			m.Cm = make([]semantic.CommitMeta, 0)
		}

		// O(n), but hundreds at most so who cares
		for k, cm := range m.Cm {
			if cm.Sha1Str == obj.Sha1Str {
				if len(obj.Tags) > 0 {
					cm.Tags = append(cm.Tags, obj.Tags...)
				}
				if len(obj.Branches) > 0 {
					cm.Branches = append(cm.Branches, obj.Branches...)
				}
				if obj.TestState != "" {
					cm.TestState = obj.TestState
				}
				m.Cm[k] = cm
				return nil
			}
		}

		m.Cm = append(m.Cm, obj)
	case semantic.PkgYum:
		if m.Yp == nil {
			m.Yp = make([]semantic.PkgYum, 0)
		}
		m.Yp = append(m.Yp, obj)
	default:
		return errors.New("type not supported")
	}

	return nil
}

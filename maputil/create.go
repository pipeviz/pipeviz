package maputil

import (
	"errors"

	"github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/types/system"
)

// FillPropMap fills a ps.Map with the provided value pairs, wrapping values in a
// types.Property struct using the provided msgid.
//
// If allowEmpty is false, only non-empty values will be included in the created map.
func FillPropMap(msgid uint64, allowEmpty bool, p ...system.PropPair) ps.Map {
	m := ps.NewMap()
	var empty bool
	var err error

	if allowEmpty {
		for _, pair := range p {
			m = m.Set(pair.K, system.Property{MsgSrc: msgid, Value: pair.V})
		}
	} else {
		for _, pair := range p {
			if empty, err = isEmptyValue(pair.V); !empty && err == nil {
				m = m.Set(pair.K, system.Property{MsgSrc: msgid, Value: pair.V})
			}
		}
	}

	return m
}

// RawMapToPropPMap fills a ps.Map with k/v pairs from the provided native Go map,
// wrapping values in a types.Property struct using the provided msgid.
//
// If allowEmpty is false, only non-empty values will be included in the created map.
func RawMapToPropPMap(msgid uint64, allowEmpty bool, in system.RawProps) ps.Map {
	m := ps.NewMap()
	var empty bool
	var err error

	if allowEmpty {
		for k, v := range in {
			m = m.Set(k, system.Property{MsgSrc: msgid, Value: v})
		}
	} else {
		for k, v := range in {
			if empty, err = isEmptyValue(v); !empty && err == nil {
				m = m.Set(k, system.Property{MsgSrc: msgid, Value: v})
			}
		}
	}

	return m
}

func isEmptyValue(v interface{}) (bool, error) {
	switch v.(type) {
	case bool:
		return v == false, nil
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return v == 0, nil
	case int, int8, int16, int32, int64:
		return v == 0, nil
	case float32, float64:
		return v == 0.0, nil
	case string:
		return v == "", nil
	case [20]byte:
		return v == [20]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, nil
	default:
		if ec, ok := v.(emptyChecker); ok {
			return ec.IsEmpty(), nil
		}
	}
	return false, errors.New("No static zero value defined for provided type")
}

// emptyChecker is an interface that raw property values can implement to
// indicate if they are empty.
type emptyChecker interface {
	IsEmpty() bool
}

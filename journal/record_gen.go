package journal

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Record) DecodeMsg(dc *msgp.Reader) (err error) {
	var ssz uint32
	ssz, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if ssz != 4 {
		err = msgp.ArrayError{Wanted: 4, Got: ssz}
		return
	}
	z.Index, err = dc.ReadUint64()
	if err != nil {
		return
	}
	z.Time, err = dc.ReadTime()
	if err != nil {
		return
	}
	z.RemoteAddr, err = dc.ReadBytes(z.RemoteAddr)
	if err != nil {
		return
	}
	z.Message, err = dc.ReadBytes(z.Message)
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Record) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 4
	err = en.Append(0x94)
	if err != nil {
		return err
	}
	err = en.WriteUint64(z.Index)
	if err != nil {
		return
	}
	err = en.WriteTime(z.Time)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.RemoteAddr)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Message)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Record) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 4
	o = append(o, 0x94)
	o = msgp.AppendUint64(o, z.Index)
	o = msgp.AppendTime(o, z.Time)
	o = msgp.AppendBytes(o, z.RemoteAddr)
	o = msgp.AppendBytes(o, z.Message)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Record) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var ssz uint32
		ssz, bts, err = msgp.ReadArrayHeaderBytes(bts)
		if err != nil {
			return
		}
		if ssz != 4 {
			err = msgp.ArrayError{Wanted: 4, Got: ssz}
			return
		}
	}
	z.Index, bts, err = msgp.ReadUint64Bytes(bts)
	if err != nil {
		return
	}
	z.Time, bts, err = msgp.ReadTimeBytes(bts)
	if err != nil {
		return
	}
	z.RemoteAddr, bts, err = msgp.ReadBytesBytes(bts, z.RemoteAddr)
	if err != nil {
		return
	}
	z.Message, bts, err = msgp.ReadBytesBytes(bts, z.Message)
	if err != nil {
		return
	}
	o = bts
	return
}

func (z *Record) Msgsize() (s int) {
	s = 1 + msgp.Uint64Size + msgp.TimeSize + msgp.BytesPrefixSize + len(z.RemoteAddr) + msgp.BytesPrefixSize + len(z.Message)
	return
}

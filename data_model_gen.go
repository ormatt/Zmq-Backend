package main

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Msg) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Ts":
			z.Ts, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "MsgID":
			z.MsgID, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Data":
			z.Data, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Msg) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "Ts"
	err = en.Append(0x83, 0xa2, 0x54, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.Ts)
	if err != nil {
		return
	}
	// write "MsgID"
	err = en.Append(0xa5, 0x4d, 0x73, 0x67, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(z.MsgID)
	if err != nil {
		return
	}
	// write "Data"
	err = en.Append(0xa4, 0x44, 0x61, 0x74, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Data)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Msg) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Ts"
	o = append(o, 0x83, 0xa2, 0x54, 0x73)
	o = msgp.AppendInt64(o, z.Ts)
	// string "MsgID"
	o = append(o, 0xa5, 0x4d, 0x73, 0x67, 0x49, 0x44)
	o = msgp.AppendString(o, z.MsgID)
	// string "Data"
	o = append(o, 0xa4, 0x44, 0x61, 0x74, 0x61)
	o = msgp.AppendString(o, z.Data)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Msg) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Ts":
			z.Ts, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "MsgID":
			z.MsgID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Data":
			z.Data, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Msg) Msgsize() (s int) {
	s = 1 + 3 + msgp.Int64Size + 6 + msgp.StringPrefixSize + len(z.MsgID) + 5 + msgp.StringPrefixSize + len(z.Data)
	return
}

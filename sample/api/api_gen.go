package api

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *AuthRequest) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Username":
			z.Username, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Pw":
			z.Pw, err = dc.ReadBytes(z.Pw)
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
func (z *AuthRequest) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Username"
	err = en.Append(0x82, 0xa8, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65)
	if err != nil {
		return
	}
	err = en.WriteString(z.Username)
	if err != nil {
		return
	}
	// write "Pw"
	err = en.Append(0xa2, 0x50, 0x77)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Pw)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *AuthRequest) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Username"
	o = append(o, 0x82, 0xa8, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.Username)
	// string "Pw"
	o = append(o, 0xa2, 0x50, 0x77)
	o = msgp.AppendBytes(o, z.Pw)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AuthRequest) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Username":
			z.Username, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Pw":
			z.Pw, bts, err = msgp.ReadBytesBytes(bts, z.Pw)
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
func (z *AuthRequest) Msgsize() (s int) {
	s = 1 + 9 + msgp.StringPrefixSize + len(z.Username) + 3 + msgp.BytesPrefixSize + len(z.Pw)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *AuthResponse) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Ok":
			z.Ok, err = dc.ReadBool()
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
func (z AuthResponse) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Ok"
	err = en.Append(0x81, 0xa2, 0x4f, 0x6b)
	if err != nil {
		return
	}
	err = en.WriteBool(z.Ok)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z AuthResponse) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Ok"
	o = append(o, 0x81, 0xa2, 0x4f, 0x6b)
	o = msgp.AppendBool(o, z.Ok)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AuthResponse) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Ok":
			z.Ok, bts, err = msgp.ReadBoolBytes(bts)
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
func (z AuthResponse) Msgsize() (s int) {
	s = 1 + 3 + msgp.BoolSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ClientInfoRet) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "RemoteAddr":
			z.RemoteAddr, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Uptime":
			z.Uptime, err = dc.ReadTime()
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
func (z ClientInfoRet) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "RemoteAddr"
	err = en.Append(0x82, 0xaa, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return
	}
	err = en.WriteString(z.RemoteAddr)
	if err != nil {
		return
	}
	// write "Uptime"
	err = en.Append(0xa6, 0x55, 0x70, 0x74, 0x69, 0x6d, 0x65)
	if err != nil {
		return
	}
	err = en.WriteTime(z.Uptime)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ClientInfoRet) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "RemoteAddr"
	o = append(o, 0x82, 0xaa, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.RemoteAddr)
	// string "Uptime"
	o = append(o, 0xa6, 0x55, 0x70, 0x74, 0x69, 0x6d, 0x65)
	o = msgp.AppendTime(o, z.Uptime)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ClientInfoRet) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "RemoteAddr":
			z.RemoteAddr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Uptime":
			z.Uptime, bts, err = msgp.ReadTimeBytes(bts)
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
func (z ClientInfoRet) Msgsize() (s int) {
	s = 1 + 11 + msgp.StringPrefixSize + len(z.RemoteAddr) + 7 + msgp.TimeSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *NewsletterFilterData) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Subscribe":
			z.Subscribe, err = dc.ReadBool()
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
func (z NewsletterFilterData) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Subscribe"
	err = en.Append(0x81, 0xa9, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65)
	if err != nil {
		return
	}
	err = en.WriteBool(z.Subscribe)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z NewsletterFilterData) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Subscribe"
	o = append(o, 0x81, 0xa9, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65)
	o = msgp.AppendBool(o, z.Subscribe)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *NewsletterFilterData) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Subscribe":
			z.Subscribe, bts, err = msgp.ReadBoolBytes(bts)
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
func (z NewsletterFilterData) Msgsize() (s int) {
	s = 1 + 10 + msgp.BoolSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *NewsletterSignalData) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Subject":
			z.Subject, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Msg":
			z.Msg, err = dc.ReadString()
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
func (z NewsletterSignalData) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Subject"
	err = en.Append(0x82, 0xa7, 0x53, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74)
	if err != nil {
		return
	}
	err = en.WriteString(z.Subject)
	if err != nil {
		return
	}
	// write "Msg"
	err = en.Append(0xa3, 0x4d, 0x73, 0x67)
	if err != nil {
		return
	}
	err = en.WriteString(z.Msg)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z NewsletterSignalData) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Subject"
	o = append(o, 0x82, 0xa7, 0x53, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74)
	o = msgp.AppendString(o, z.Subject)
	// string "Msg"
	o = append(o, 0xa3, 0x4d, 0x73, 0x67)
	o = msgp.AppendString(o, z.Msg)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *NewsletterSignalData) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Subject":
			z.Subject, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Msg":
			z.Msg, bts, err = msgp.ReadStringBytes(bts)
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
func (z NewsletterSignalData) Msgsize() (s int) {
	s = 1 + 8 + msgp.StringPrefixSize + len(z.Subject) + 4 + msgp.StringPrefixSize + len(z.Msg)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ServerInfoRet) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "RemoteAddr":
			z.RemoteAddr, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Uptime":
			z.Uptime, err = dc.ReadTime()
			if err != nil {
				return
			}
		case "ClientsCount":
			z.ClientsCount, err = dc.ReadInt()
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
func (z ServerInfoRet) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "RemoteAddr"
	err = en.Append(0x83, 0xaa, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return
	}
	err = en.WriteString(z.RemoteAddr)
	if err != nil {
		return
	}
	// write "Uptime"
	err = en.Append(0xa6, 0x55, 0x70, 0x74, 0x69, 0x6d, 0x65)
	if err != nil {
		return
	}
	err = en.WriteTime(z.Uptime)
	if err != nil {
		return
	}
	// write "ClientsCount"
	err = en.Append(0xac, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74)
	if err != nil {
		return
	}
	err = en.WriteInt(z.ClientsCount)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ServerInfoRet) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "RemoteAddr"
	o = append(o, 0x83, 0xaa, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.RemoteAddr)
	// string "Uptime"
	o = append(o, 0xa6, 0x55, 0x70, 0x74, 0x69, 0x6d, 0x65)
	o = msgp.AppendTime(o, z.Uptime)
	// string "ClientsCount"
	o = append(o, 0xac, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74)
	o = msgp.AppendInt(o, z.ClientsCount)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ServerInfoRet) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "RemoteAddr":
			z.RemoteAddr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Uptime":
			z.Uptime, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				return
			}
		case "ClientsCount":
			z.ClientsCount, bts, err = msgp.ReadIntBytes(bts)
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
func (z ServerInfoRet) Msgsize() (s int) {
	s = 1 + 11 + msgp.StringPrefixSize + len(z.RemoteAddr) + 7 + msgp.TimeSize + 13 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TimeBombData) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Countdown":
			z.Countdown, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "HasDetonated":
			z.HasDetonated, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "Gift":
			z.Gift, err = dc.ReadString()
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
func (z TimeBombData) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "Countdown"
	err = en.Append(0x83, 0xa9, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x64, 0x6f, 0x77, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteInt(z.Countdown)
	if err != nil {
		return
	}
	// write "HasDetonated"
	err = en.Append(0xac, 0x48, 0x61, 0x73, 0x44, 0x65, 0x74, 0x6f, 0x6e, 0x61, 0x74, 0x65, 0x64)
	if err != nil {
		return
	}
	err = en.WriteBool(z.HasDetonated)
	if err != nil {
		return
	}
	// write "Gift"
	err = en.Append(0xa4, 0x47, 0x69, 0x66, 0x74)
	if err != nil {
		return
	}
	err = en.WriteString(z.Gift)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z TimeBombData) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Countdown"
	o = append(o, 0x83, 0xa9, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x64, 0x6f, 0x77, 0x6e)
	o = msgp.AppendInt(o, z.Countdown)
	// string "HasDetonated"
	o = append(o, 0xac, 0x48, 0x61, 0x73, 0x44, 0x65, 0x74, 0x6f, 0x6e, 0x61, 0x74, 0x65, 0x64)
	o = msgp.AppendBool(o, z.HasDetonated)
	// string "Gift"
	o = append(o, 0xa4, 0x47, 0x69, 0x66, 0x74)
	o = msgp.AppendString(o, z.Gift)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TimeBombData) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Countdown":
			z.Countdown, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "HasDetonated":
			z.HasDetonated, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "Gift":
			z.Gift, bts, err = msgp.ReadStringBytes(bts)
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
func (z TimeBombData) Msgsize() (s int) {
	s = 1 + 10 + msgp.IntSize + 13 + msgp.BoolSize + 5 + msgp.StringPrefixSize + len(z.Gift)
	return
}

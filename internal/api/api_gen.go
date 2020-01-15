package api

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Call) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "ID":
			z.ID, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "ID")
				return
			}
		case "Key":
			z.Key, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "Key")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Call) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "ID"
	err = en.Append(0x82, 0xa2, 0x49, 0x44)
	if err != nil {
		return
	}
	err = en.WriteString(z.ID)
	if err != nil {
		err = msgp.WrapError(err, "ID")
		return
	}
	// write "Key"
	err = en.Append(0xa3, 0x4b, 0x65, 0x79)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Key)
	if err != nil {
		err = msgp.WrapError(err, "Key")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Call) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "ID"
	o = append(o, 0x82, 0xa2, 0x49, 0x44)
	o = msgp.AppendString(o, z.ID)
	// string "Key"
	o = append(o, 0xa3, 0x4b, 0x65, 0x79)
	o = msgp.AppendUint32(o, z.Key)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Call) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "ID":
			z.ID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "ID")
				return
			}
		case "Key":
			z.Key, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Key")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Call) Msgsize() (s int) {
	s = 1 + 3 + msgp.StringPrefixSize + len(z.ID) + 4 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *CallCancel) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Key":
			z.Key, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "Key")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z CallCancel) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Key"
	err = en.Append(0x81, 0xa3, 0x4b, 0x65, 0x79)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Key)
	if err != nil {
		err = msgp.WrapError(err, "Key")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z CallCancel) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Key"
	o = append(o, 0x81, 0xa3, 0x4b, 0x65, 0x79)
	o = msgp.AppendUint32(o, z.Key)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CallCancel) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Key":
			z.Key, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Key")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z CallCancel) Msgsize() (s int) {
	s = 1 + 4 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *CallReturn) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Key":
			z.Key, err = dc.ReadUint32()
			if err != nil {
				err = msgp.WrapError(err, "Key")
				return
			}
		case "Msg":
			z.Msg, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Msg")
				return
			}
		case "Code":
			z.Code, err = dc.ReadInt()
			if err != nil {
				err = msgp.WrapError(err, "Code")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z CallReturn) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "Key"
	err = en.Append(0x83, 0xa3, 0x4b, 0x65, 0x79)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Key)
	if err != nil {
		err = msgp.WrapError(err, "Key")
		return
	}
	// write "Msg"
	err = en.Append(0xa3, 0x4d, 0x73, 0x67)
	if err != nil {
		return
	}
	err = en.WriteString(z.Msg)
	if err != nil {
		err = msgp.WrapError(err, "Msg")
		return
	}
	// write "Code"
	err = en.Append(0xa4, 0x43, 0x6f, 0x64, 0x65)
	if err != nil {
		return
	}
	err = en.WriteInt(z.Code)
	if err != nil {
		err = msgp.WrapError(err, "Code")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z CallReturn) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Key"
	o = append(o, 0x83, 0xa3, 0x4b, 0x65, 0x79)
	o = msgp.AppendUint32(o, z.Key)
	// string "Msg"
	o = append(o, 0xa3, 0x4d, 0x73, 0x67)
	o = msgp.AppendString(o, z.Msg)
	// string "Code"
	o = append(o, 0xa4, 0x43, 0x6f, 0x64, 0x65)
	o = msgp.AppendInt(o, z.Code)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CallReturn) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Key":
			z.Key, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Key")
				return
			}
		case "Msg":
			z.Msg, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Msg")
				return
			}
		case "Code":
			z.Code, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Code")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z CallReturn) Msgsize() (s int) {
	s = 1 + 4 + msgp.Uint32Size + 4 + msgp.StringPrefixSize + len(z.Msg) + 5 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *InitStream) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "ID":
			z.ID, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "ID")
				return
			}
		case "Type":
			{
				var zb0002 int
				zb0002, err = dc.ReadInt()
				if err != nil {
					err = msgp.WrapError(err, "Type")
					return
				}
				z.Type = StreamType(zb0002)
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z InitStream) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "ID"
	err = en.Append(0x82, 0xa2, 0x49, 0x44)
	if err != nil {
		return
	}
	err = en.WriteString(z.ID)
	if err != nil {
		err = msgp.WrapError(err, "ID")
		return
	}
	// write "Type"
	err = en.Append(0xa4, 0x54, 0x79, 0x70, 0x65)
	if err != nil {
		return
	}
	err = en.WriteInt(int(z.Type))
	if err != nil {
		err = msgp.WrapError(err, "Type")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z InitStream) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "ID"
	o = append(o, 0x82, 0xa2, 0x49, 0x44)
	o = msgp.AppendString(o, z.ID)
	// string "Type"
	o = append(o, 0xa4, 0x54, 0x79, 0x70, 0x65)
	o = msgp.AppendInt(o, int(z.Type))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *InitStream) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "ID":
			z.ID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "ID")
				return
			}
		case "Type":
			{
				var zb0002 int
				zb0002, bts, err = msgp.ReadIntBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Type")
					return
				}
				z.Type = StreamType(zb0002)
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z InitStream) Msgsize() (s int) {
	s = 1 + 3 + msgp.StringPrefixSize + len(z.ID) + 5 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *StreamType) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 int
		zb0001, err = dc.ReadInt()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = StreamType(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z StreamType) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteInt(int(z))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z StreamType) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendInt(o, int(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *StreamType) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 int
		zb0001, bts, err = msgp.ReadIntBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = StreamType(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z StreamType) Msgsize() (s int) {
	s = msgp.IntSize
	return
}

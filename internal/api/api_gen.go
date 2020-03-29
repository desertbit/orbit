/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2020 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2020 Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package api

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *HandshakeArgs) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Version":
			z.Version, err = dc.ReadByte()
			if err != nil {
				err = msgp.WrapError(err, "Version")
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
func (z HandshakeArgs) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Version"
	err = en.Append(0x81, 0xa7, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteByte(z.Version)
	if err != nil {
		err = msgp.WrapError(err, "Version")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z HandshakeArgs) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Version"
	o = append(o, 0x81, 0xa7, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	o = msgp.AppendByte(o, z.Version)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *HandshakeArgs) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Version":
			z.Version, bts, err = msgp.ReadByteBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Version")
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
func (z HandshakeArgs) Msgsize() (s int) {
	s = 1 + 8 + msgp.ByteSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *HandshakeCode) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 int
		zb0001, err = dc.ReadInt()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = HandshakeCode(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z HandshakeCode) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteInt(int(z))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z HandshakeCode) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendInt(o, int(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *HandshakeCode) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 int
		zb0001, bts, err = msgp.ReadIntBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = HandshakeCode(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z HandshakeCode) Msgsize() (s int) {
	s = msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *HandshakeRet) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Code":
			{
				var zb0002 int
				zb0002, err = dc.ReadInt()
				if err != nil {
					err = msgp.WrapError(err, "Code")
					return
				}
				z.Code = HandshakeCode(zb0002)
			}
		case "SessionID":
			z.SessionID, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "SessionID")
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
func (z HandshakeRet) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Code"
	err = en.Append(0x82, 0xa4, 0x43, 0x6f, 0x64, 0x65)
	if err != nil {
		return
	}
	err = en.WriteInt(int(z.Code))
	if err != nil {
		err = msgp.WrapError(err, "Code")
		return
	}
	// write "SessionID"
	err = en.Append(0xa9, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x44)
	if err != nil {
		return
	}
	err = en.WriteString(z.SessionID)
	if err != nil {
		err = msgp.WrapError(err, "SessionID")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z HandshakeRet) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Code"
	o = append(o, 0x82, 0xa4, 0x43, 0x6f, 0x64, 0x65)
	o = msgp.AppendInt(o, int(z.Code))
	// string "SessionID"
	o = append(o, 0xa9, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x44)
	o = msgp.AppendString(o, z.SessionID)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *HandshakeRet) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Code":
			{
				var zb0002 int
				zb0002, bts, err = msgp.ReadIntBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Code")
					return
				}
				z.Code = HandshakeCode(zb0002)
			}
		case "SessionID":
			z.SessionID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "SessionID")
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
func (z HandshakeRet) Msgsize() (s int) {
	s = 1 + 5 + msgp.IntSize + 10 + msgp.StringPrefixSize + len(z.SessionID)
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
		case "Data":
			var zb0003 uint32
			zb0003, err = dc.ReadMapHeader()
			if err != nil {
				err = msgp.WrapError(err, "Data")
				return
			}
			if z.Data == nil {
				z.Data = make(map[string][]byte, zb0003)
			} else if len(z.Data) > 0 {
				for key := range z.Data {
					delete(z.Data, key)
				}
			}
			for zb0003 > 0 {
				zb0003--
				var za0001 string
				var za0002 []byte
				za0001, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Data")
					return
				}
				za0002, err = dc.ReadBytes(za0002)
				if err != nil {
					err = msgp.WrapError(err, "Data", za0001)
					return
				}
				z.Data[za0001] = za0002
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
func (z *InitStream) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "ID"
	err = en.Append(0x83, 0xa2, 0x49, 0x44)
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
	// write "Data"
	err = en.Append(0xa4, 0x44, 0x61, 0x74, 0x61)
	if err != nil {
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Data)))
	if err != nil {
		err = msgp.WrapError(err, "Data")
		return
	}
	for za0001, za0002 := range z.Data {
		err = en.WriteString(za0001)
		if err != nil {
			err = msgp.WrapError(err, "Data")
			return
		}
		err = en.WriteBytes(za0002)
		if err != nil {
			err = msgp.WrapError(err, "Data", za0001)
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *InitStream) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "ID"
	o = append(o, 0x83, 0xa2, 0x49, 0x44)
	o = msgp.AppendString(o, z.ID)
	// string "Type"
	o = append(o, 0xa4, 0x54, 0x79, 0x70, 0x65)
	o = msgp.AppendInt(o, int(z.Type))
	// string "Data"
	o = append(o, 0xa4, 0x44, 0x61, 0x74, 0x61)
	o = msgp.AppendMapHeader(o, uint32(len(z.Data)))
	for za0001, za0002 := range z.Data {
		o = msgp.AppendString(o, za0001)
		o = msgp.AppendBytes(o, za0002)
	}
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
		case "Data":
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Data")
				return
			}
			if z.Data == nil {
				z.Data = make(map[string][]byte, zb0003)
			} else if len(z.Data) > 0 {
				for key := range z.Data {
					delete(z.Data, key)
				}
			}
			for zb0003 > 0 {
				var za0001 string
				var za0002 []byte
				zb0003--
				za0001, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Data")
					return
				}
				za0002, bts, err = msgp.ReadBytesBytes(bts, za0002)
				if err != nil {
					err = msgp.WrapError(err, "Data", za0001)
					return
				}
				z.Data[za0001] = za0002
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
func (z *InitStream) Msgsize() (s int) {
	s = 1 + 3 + msgp.StringPrefixSize + len(z.ID) + 5 + msgp.IntSize + 5 + msgp.MapHeaderSize
	if z.Data != nil {
		for za0001, za0002 := range z.Data {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + msgp.BytesPrefixSize + len(za0002)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RPCCall) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Data":
			var zb0002 uint32
			zb0002, err = dc.ReadMapHeader()
			if err != nil {
				err = msgp.WrapError(err, "Data")
				return
			}
			if z.Data == nil {
				z.Data = make(map[string][]byte, zb0002)
			} else if len(z.Data) > 0 {
				for key := range z.Data {
					delete(z.Data, key)
				}
			}
			for zb0002 > 0 {
				zb0002--
				var za0001 string
				var za0002 []byte
				za0001, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Data")
					return
				}
				za0002, err = dc.ReadBytes(za0002)
				if err != nil {
					err = msgp.WrapError(err, "Data", za0001)
					return
				}
				z.Data[za0001] = za0002
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
func (z *RPCCall) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "ID"
	err = en.Append(0x83, 0xa2, 0x49, 0x44)
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
	// write "Data"
	err = en.Append(0xa4, 0x44, 0x61, 0x74, 0x61)
	if err != nil {
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Data)))
	if err != nil {
		err = msgp.WrapError(err, "Data")
		return
	}
	for za0001, za0002 := range z.Data {
		err = en.WriteString(za0001)
		if err != nil {
			err = msgp.WrapError(err, "Data")
			return
		}
		err = en.WriteBytes(za0002)
		if err != nil {
			err = msgp.WrapError(err, "Data", za0001)
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *RPCCall) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "ID"
	o = append(o, 0x83, 0xa2, 0x49, 0x44)
	o = msgp.AppendString(o, z.ID)
	// string "Key"
	o = append(o, 0xa3, 0x4b, 0x65, 0x79)
	o = msgp.AppendUint32(o, z.Key)
	// string "Data"
	o = append(o, 0xa4, 0x44, 0x61, 0x74, 0x61)
	o = msgp.AppendMapHeader(o, uint32(len(z.Data)))
	for za0001, za0002 := range z.Data {
		o = msgp.AppendString(o, za0001)
		o = msgp.AppendBytes(o, za0002)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RPCCall) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Data":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Data")
				return
			}
			if z.Data == nil {
				z.Data = make(map[string][]byte, zb0002)
			} else if len(z.Data) > 0 {
				for key := range z.Data {
					delete(z.Data, key)
				}
			}
			for zb0002 > 0 {
				var za0001 string
				var za0002 []byte
				zb0002--
				za0001, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Data")
					return
				}
				za0002, bts, err = msgp.ReadBytesBytes(bts, za0002)
				if err != nil {
					err = msgp.WrapError(err, "Data", za0001)
					return
				}
				z.Data[za0001] = za0002
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
func (z *RPCCall) Msgsize() (s int) {
	s = 1 + 3 + msgp.StringPrefixSize + len(z.ID) + 4 + msgp.Uint32Size + 5 + msgp.MapHeaderSize
	if z.Data != nil {
		for za0001, za0002 := range z.Data {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + msgp.BytesPrefixSize + len(za0002)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RPCCancel) DecodeMsg(dc *msgp.Reader) (err error) {
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
func (z RPCCancel) EncodeMsg(en *msgp.Writer) (err error) {
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
func (z RPCCancel) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Key"
	o = append(o, 0x81, 0xa3, 0x4b, 0x65, 0x79)
	o = msgp.AppendUint32(o, z.Key)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RPCCancel) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
func (z RPCCancel) Msgsize() (s int) {
	s = 1 + 4 + msgp.Uint32Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RPCReturn) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Err":
			z.Err, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Err")
				return
			}
		case "ErrCode":
			z.ErrCode, err = dc.ReadInt()
			if err != nil {
				err = msgp.WrapError(err, "ErrCode")
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
func (z RPCReturn) EncodeMsg(en *msgp.Writer) (err error) {
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
	// write "Err"
	err = en.Append(0xa3, 0x45, 0x72, 0x72)
	if err != nil {
		return
	}
	err = en.WriteString(z.Err)
	if err != nil {
		err = msgp.WrapError(err, "Err")
		return
	}
	// write "ErrCode"
	err = en.Append(0xa7, 0x45, 0x72, 0x72, 0x43, 0x6f, 0x64, 0x65)
	if err != nil {
		return
	}
	err = en.WriteInt(z.ErrCode)
	if err != nil {
		err = msgp.WrapError(err, "ErrCode")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z RPCReturn) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Key"
	o = append(o, 0x83, 0xa3, 0x4b, 0x65, 0x79)
	o = msgp.AppendUint32(o, z.Key)
	// string "Err"
	o = append(o, 0xa3, 0x45, 0x72, 0x72)
	o = msgp.AppendString(o, z.Err)
	// string "ErrCode"
	o = append(o, 0xa7, 0x45, 0x72, 0x72, 0x43, 0x6f, 0x64, 0x65)
	o = msgp.AppendInt(o, z.ErrCode)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RPCReturn) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Err":
			z.Err, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Err")
				return
			}
		case "ErrCode":
			z.ErrCode, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "ErrCode")
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
func (z RPCReturn) Msgsize() (s int) {
	s = 1 + 4 + msgp.Uint32Size + 4 + msgp.StringPrefixSize + len(z.Err) + 8 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RPCType) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 byte
		zb0001, err = dc.ReadByte()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = RPCType(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RPCType) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteByte(byte(z))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z RPCType) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendByte(o, byte(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RPCType) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 byte
		zb0001, bts, err = msgp.ReadByteBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = RPCType(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z RPCType) Msgsize() (s int) {
	s = msgp.ByteSize
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

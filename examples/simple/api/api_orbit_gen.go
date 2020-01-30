/* code generated by orbit */
package api

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/codec"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/desertbit/orbit/pkg/packet"
)

var (
	_ context.Context
	_ = errors.New("")
	_ net.Conn
	_ time.Time
	_ sync.Locker
	_ orbit.Conn
	_ = packet.MaxSize
	_ closer.Closer
)

//##############//
//### Errors ###//
//##############//

var (
	ErrClosed = errors.New("closed")
)

const (
	ErrCodeTheFirstError  = 1
	ErrCodeTheSecondError = 2
	ErrCodeTheThirdError  = 3
)

var (
	ErrTheFirstError       = errors.New("the first error")
	orbitErrTheFirstError  = orbit.Err(ErrTheFirstError, ErrTheFirstError.Error(), ErrCodeTheFirstError)
	ErrTheSecondError      = errors.New("the second error")
	orbitErrTheSecondError = orbit.Err(ErrTheSecondError, ErrTheSecondError.Error(), ErrCodeTheSecondError)
	ErrTheThirdError       = errors.New("the third error")
	orbitErrTheThirdError  = orbit.Err(ErrTheThirdError, ErrTheThirdError.Error(), ErrCodeTheThirdError)
)

//#############//
//### Types ###//
//#############//

type Args struct {
	Crazy map[string][][]map[string]En1
	I     int
	M     map[string]int
	S     string
	Sl    []time.Time
	St    *Ret
}

type Rc1Ret struct {
	Crazy map[string][][]map[string]En1
	I     int
	M     map[string]int
	S     string
	Sl    []time.Time
	St    *Ret
}

type Rc2Args struct {
	B   byte
	F   float64
	U16 uint16
	U32 uint32
	U64 uint64
	U8  uint8
}

type Ret struct {
	B   byte
	F   float64
	U16 uint16
	U32 uint32
	U64 uint64
	U8  uint8
}

//msgp:ignore StringReadChan
type StringReadChan struct {
	closer.Closer
	C   <-chan string
	c   chan string
	mx  sync.Mutex
	err error
}

func newStringReadChan(cl closer.Closer, size uint) *StringReadChan {
	c := &StringReadChan{Closer: cl, c: make(chan string, size)}
	c.C = c.c
	return c
}

func (c *StringReadChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *StringReadChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

//msgp:ignore StringWriteChan
type StringWriteChan struct {
	closer.Closer
	C   chan<- string
	c   chan string
	mx  sync.Mutex
	err error
}

func newStringWriteChan(cl closer.Closer, size uint) *StringWriteChan {
	c := &StringWriteChan{Closer: cl, c: make(chan string, size)}
	c.C = c.c
	return c
}

func (c *StringWriteChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *StringWriteChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

//msgp:ignore En1ReadChan
type En1ReadChan struct {
	closer.Closer
	C   <-chan En1
	c   chan En1
	mx  sync.Mutex
	err error
}

func newEn1ReadChan(cl closer.Closer, size uint) *En1ReadChan {
	c := &En1ReadChan{Closer: cl, c: make(chan En1, size)}
	c.C = c.c
	return c
}

func (c *En1ReadChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *En1ReadChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

//msgp:ignore En1WriteChan
type En1WriteChan struct {
	closer.Closer
	C   chan<- En1
	c   chan En1
	mx  sync.Mutex
	err error
}

func newEn1WriteChan(cl closer.Closer, size uint) *En1WriteChan {
	c := &En1WriteChan{Closer: cl, c: make(chan En1, size)}
	c.C = c.c
	return c
}

func (c *En1WriteChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *En1WriteChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

//msgp:ignore ArgsReadChan
type ArgsReadChan struct {
	closer.Closer
	stream net.Conn
	codec  codec.Codec
}

func newArgsReadChan(cl closer.Closer, stream net.Conn, cc codec.Codec) *ArgsReadChan {
	return &ArgsReadChan{Closer: cl, stream: stream, codec: cc}
}

func (c *ArgsReadChan) Read() (arg *Args, err error) {
	if c.IsClosing() {
		err = ErrClosed
		return
	}
	err = packet.ReadDecode(c.stream, &arg, c.codec)
	if err != nil {
		if errors.Is(err, packet.ErrZeroData) || errors.Is(err, io.EOF) {
			err = ErrClosed
		}
		c.Close_()
		return
	}
	return
}

//msgp:ignore ArgsWriteChan
type ArgsWriteChan struct {
	closer.Closer
	C   chan<- *Args
	c   chan *Args
	mx  sync.Mutex
	err error
}

func newArgsWriteChan(cl closer.Closer, size uint) *ArgsWriteChan {
	c := &ArgsWriteChan{Closer: cl, c: make(chan *Args, size)}
	c.C = c.c
	return c
}

func (c *ArgsWriteChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *ArgsWriteChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

//msgp:ignore RetReadChan
type RetReadChan struct {
	closer.Closer
	stream net.Conn
	codec  codec.Codec
}

func newRetReadChan(cl closer.Closer, stream net.Conn, cc codec.Codec) *RetReadChan {
	return &RetReadChan{Closer: cl, stream: stream, codec: cc}
}

//msgp:ignore RetWriteChan
type RetWriteChan struct {
	closer.Closer
	stream net.Conn
	codec  codec.Codec
}

func newRetWriteChan(cl closer.Closer, stream net.Conn, cc codec.Codec) *RetWriteChan {
	cl.OnClosing(func() error { return packet.Write(stream, nil) })
	return &RetWriteChan{Closer: cl, stream: stream, codec: cc}
}

func (c *RetWriteChan) Write(data *Ret) (err error) {
	if c.IsClosing() {
		return ErrClosed
	}
	err = packet.WriteEncode(c.stream, data, c.codec)
	if errors.Is(err, io.EOF) {
		c.Close_()
		return ErrClosed
	}
	return
}

//msgp:ignore MapStringIntReadChan
type MapStringIntReadChan struct {
	closer.Closer
	C   <-chan map[string]int
	c   chan map[string]int
	mx  sync.Mutex
	err error
}

func newMapStringIntReadChan(cl closer.Closer, size uint) *MapStringIntReadChan {
	c := &MapStringIntReadChan{Closer: cl, c: make(chan map[string]int, size)}
	c.C = c.c
	return c
}

func (c *MapStringIntReadChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *MapStringIntReadChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

//msgp:ignore MapStringIntWriteChan
type MapStringIntWriteChan struct {
	closer.Closer
	C   chan<- map[string]int
	c   chan map[string]int
	mx  sync.Mutex
	err error
}

func newMapStringIntWriteChan(cl closer.Closer, size uint) *MapStringIntWriteChan {
	c := &MapStringIntWriteChan{Closer: cl, c: make(chan map[string]int, size)}
	c.C = c.c
	return c
}

func (c *MapStringIntWriteChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *MapStringIntWriteChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

//#############//
//### Enums ###//
//#############//

type En1 int

const (
	Val1 En1 = 1
	Val2 En1 = 2
	Val3 En1 = 3
)

//################//
//### Services ###//
//################//

// S1  ---------------------
// Service
const (
	ServiceS1 = "S1"
	S1C1      = "C1"
	S1C2      = "C2"
	S1C3      = "C3"
	S1Rc1     = "Rc1"
	S1Rc2     = "Rc2"
	S1Rc3     = "Rc3"
	S1S1      = "S1"
	S1S2      = "S2"
	S1S3      = "S3"
	S1Rs1     = "Rs1"
	S1Rs2     = "Rs2"
	S1Rs3     = "Rs3"
)

type S1ConsumerCaller interface {
	// Calls
	C1(ctx context.Context, args int) (ret float32, err error)
	C2(ctx context.Context, args time.Time) (ret []map[string][]*Ret, err error)
	C3(ctx context.Context) (err error)
	// Streams
	S1(ctx context.Context) (stream net.Conn, err error)
	S2(ctx context.Context) (args *StringWriteChan, err error)
	S3(ctx context.Context) (ret *En1ReadChan, err error)
}

type S1ConsumerHandler interface {
	// Calls
	Rc1(ctx context.Context, s *orbit.Session, args *Args) (ret *Rc1Ret, err error)
	Rc2(ctx context.Context, s *orbit.Session, args *Rc2Args) (err error)
	Rc3(ctx context.Context, s *orbit.Session) (err error)
	// Streams
	Rs1(s *orbit.Session, args *ArgsReadChan, ret *RetWriteChan)
	Rs2(s *orbit.Session, args *MapStringIntReadChan) (err error)
	Rs3(s *orbit.Session, stream net.Conn) (err error)
}

type S1ProviderCaller interface {
	// Calls
	Rc1(ctx context.Context, args *Args) (ret *Rc1Ret, err error)
	Rc2(ctx context.Context, args *Rc2Args) (err error)
	Rc3(ctx context.Context) (err error)
	// Streams
	Rs1(ctx context.Context) (args *ArgsWriteChan, ret *RetReadChan, err error)
	Rs2(ctx context.Context) (args *MapStringIntWriteChan, err error)
	Rs3(ctx context.Context) (stream net.Conn, err error)
}

type S1ProviderHandler interface {
	// Calls
	C1(ctx context.Context, s *orbit.Session, args int) (ret float32, err error)
	C2(ctx context.Context, s *orbit.Session, args time.Time) (ret []map[string][]*Ret, err error)
	C3(ctx context.Context, s *orbit.Session) (err error)
	// Streams
	S1(s *orbit.Session, stream net.Conn) (err error)
	S2(s *orbit.Session, args *StringReadChan) (err error)
	S3(s *orbit.Session, ret *En1WriteChan) (err error)
}

type s1Consumer struct {
	h S1ConsumerHandler
	s *orbit.Session
}

func RegisterS1Consumer(s *orbit.Session, h S1ConsumerHandler) S1ConsumerCaller {
	cc := &s1Consumer{h: h, s: s}
	s.RegisterCall(ServiceS1, S1Rc1, cc.rc1)
	s.RegisterCall(ServiceS1, S1Rc2, cc.rc2)
	s.RegisterCall(ServiceS1, S1Rc3, cc.rc3)
	s.RegisterStream(ServiceS1, S1Rs1, cc.rs1)
	s.RegisterStream(ServiceS1, S1Rs2, cc.rs2)
	s.RegisterStream(ServiceS1, S1Rs3, cc.rs3)
	return cc
}

func (v1 *s1Consumer) C1(ctx context.Context, args int) (ret float32, err error) {
	retData, err := v1.s.Call(ctx, ServiceS1, S1C1, args)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeTheFirstError:
				err = ErrTheFirstError
			case ErrCodeTheSecondError:
				err = ErrTheSecondError
			case ErrCodeTheThirdError:
				err = ErrTheThirdError
			}
		}
		return
	}
	err = retData.Decode(&ret)
	if err != nil {
		return
	}
	return
}

func (v1 *s1Consumer) C2(ctx context.Context, args time.Time) (ret []map[string][]*Ret, err error) {
	retData, err := v1.s.CallAsync(ctx, ServiceS1, S1C2, args)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeTheFirstError:
				err = ErrTheFirstError
			case ErrCodeTheSecondError:
				err = ErrTheSecondError
			case ErrCodeTheThirdError:
				err = ErrTheThirdError
			}
		}
		return
	}
	err = retData.Decode(&ret)
	if err != nil {
		return
	}
	return
}

func (v1 *s1Consumer) C3(ctx context.Context) (err error) {
	_, err = v1.s.Call(ctx, ServiceS1, S1C3, nil)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeTheFirstError:
				err = ErrTheFirstError
			case ErrCodeTheSecondError:
				err = ErrTheSecondError
			case ErrCodeTheThirdError:
				err = ErrTheThirdError
			}
		}
		return
	}
	return
}

func (v1 *s1Consumer) rc1(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	var args *Args
	err = ad.Decode(&args)
	if err != nil {
		return
	}
	ret, err := v1.h.Rc1(ctx, s, args)
	if err != nil {
		if errors.Is(err, ErrTheFirstError) {
			err = orbitErrTheFirstError
		} else if errors.Is(err, ErrTheSecondError) {
			err = orbitErrTheSecondError
		} else if errors.Is(err, ErrTheThirdError) {
			err = orbitErrTheThirdError
		}
		return
	}
	r = ret
	return
}

func (v1 *s1Consumer) rc2(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	var args *Rc2Args
	err = ad.Decode(&args)
	if err != nil {
		return
	}
	err = v1.h.Rc2(ctx, s, args)
	if err != nil {
		if errors.Is(err, ErrTheFirstError) {
			err = orbitErrTheFirstError
		} else if errors.Is(err, ErrTheSecondError) {
			err = orbitErrTheSecondError
		} else if errors.Is(err, ErrTheThirdError) {
			err = orbitErrTheThirdError
		}
		return
	}
	return
}

func (v1 *s1Consumer) rc3(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	err = v1.h.Rc3(ctx, s)
	if err != nil {
		if errors.Is(err, ErrTheFirstError) {
			err = orbitErrTheFirstError
		} else if errors.Is(err, ErrTheSecondError) {
			err = orbitErrTheSecondError
		} else if errors.Is(err, ErrTheThirdError) {
			err = orbitErrTheThirdError
		}
		return
	}
	return
}

func (v1 *s1Consumer) S1(ctx context.Context) (stream net.Conn, err error) {
	return v1.s.OpenStream(ctx, ServiceS1, S1S1)
}

func (v1 *s1Consumer) S2(ctx context.Context) (args *StringWriteChan, err error) {
	stream, err := v1.s.OpenStream(ctx, ServiceS1, S1S2)
	if err != nil {
		return
	}
	args = newStringWriteChan(v1.s.CloserOneWay(), v1.s.StreamChanSize())
	args.OnClosing(func() error { return stream.Close() })
	go func() {
		closingChan := args.ClosingChan()
		codec := v1.s.Codec()
		for {
			select {
			case <-closingChan:
				return
			case arg := <-args.c:
				err := packet.WriteEncode(stream, arg, codec)
				if err != nil {
					if args.IsClosing() {
						err = nil
					}
					args.setError(err)
					return
				}
			}
		}
	}()
	return
}

func (v1 *s1Consumer) S3(ctx context.Context) (ret *En1ReadChan, err error) {
	stream, err := v1.s.OpenStream(ctx, ServiceS1, S1S3)
	if err != nil {
		return
	}
	ret = newEn1ReadChan(v1.s.CloserOneWay(), v1.s.StreamChanSize())
	ret.OnClosing(func() error { return stream.Close() })
	go func() {
		closingChan := ret.ClosingChan()
		codec := v1.s.Codec()
		for {
			var data En1
			err := packet.ReadDecode(stream, &data, codec)
			if err != nil {
				if ret.IsClosing() {
					err = nil
				}
				ret.setError(err)
				return
			}
			select {
			case <-closingChan:
				return
			case ret.c <- data:
			}
		}
	}()
	return
}

func (v1 *s1Consumer) rs1(s *orbit.Session, stream net.Conn) {
	args := newArgsReadChan(v1.s.CloserOneWay(), stream, v1.s.Codec())
	ret := newRetWriteChan(v1.s.CloserOneWay(), stream, v1.s.Codec())

	go func() {
		<-args.ClosedChan()
		<-ret.ClosedChan()
		_ = stream.Close()
	}()

	v1.h.Rs1(s, args, ret)
}

func (v1 *s1Consumer) rs2(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	args := newMapStringIntReadChan(v1.s.CloserOneWay(), v1.s.StreamChanSize())
	go func() {
		closingChan := args.ClosingChan()
		codec := v1.s.Codec()
		for {
			var arg map[string]int
			err := packet.ReadDecode(stream, &arg, codec)
			if err != nil {
				if args.IsClosing() {
					err = nil
				}
				args.setError(err)
				return
			}
			select {
			case <-closingChan:
				return
			case args.c <- arg:
			}
		}
	}()

	err = v1.h.Rs2(s, args)
	if err != nil {
		return
	}
	return
}

func (v1 *s1Consumer) rs3(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	err = v1.h.Rs3(s, stream)
	if err != nil {
		return
	}
	return
}

type s1Provider struct {
	h S1ProviderHandler
	s *orbit.Session
}

func RegisterS1Provider(s *orbit.Session, h S1ProviderHandler) S1ProviderCaller {
	cc := &s1Provider{h: h, s: s}
	s.RegisterCall(ServiceS1, S1C1, cc.c1)
	s.RegisterCall(ServiceS1, S1C2, cc.c2)
	s.RegisterCall(ServiceS1, S1C3, cc.c3)
	s.RegisterStream(ServiceS1, S1S1, cc.s1)
	s.RegisterStream(ServiceS1, S1S2, cc.s2)
	s.RegisterStream(ServiceS1, S1S3, cc.s3)
	return cc
}

func (v1 *s1Provider) Rc1(ctx context.Context, args *Args) (ret *Rc1Ret, err error) {
	retData, err := v1.s.Call(ctx, ServiceS1, S1Rc1, args)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeTheFirstError:
				err = ErrTheFirstError
			case ErrCodeTheSecondError:
				err = ErrTheSecondError
			case ErrCodeTheThirdError:
				err = ErrTheThirdError
			}
		}
		return
	}
	err = retData.Decode(&ret)
	if err != nil {
		return
	}
	return
}

func (v1 *s1Provider) Rc2(ctx context.Context, args *Rc2Args) (err error) {
	_, err = v1.s.CallAsync(ctx, ServiceS1, S1Rc2, args)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeTheFirstError:
				err = ErrTheFirstError
			case ErrCodeTheSecondError:
				err = ErrTheSecondError
			case ErrCodeTheThirdError:
				err = ErrTheThirdError
			}
		}
		return
	}
	return
}

func (v1 *s1Provider) Rc3(ctx context.Context) (err error) {
	_, err = v1.s.Call(ctx, ServiceS1, S1Rc3, nil)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeTheFirstError:
				err = ErrTheFirstError
			case ErrCodeTheSecondError:
				err = ErrTheSecondError
			case ErrCodeTheThirdError:
				err = ErrTheThirdError
			}
		}
		return
	}
	return
}

func (v1 *s1Provider) c1(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	var args int
	err = ad.Decode(&args)
	if err != nil {
		return
	}
	ret, err := v1.h.C1(ctx, s, args)
	if err != nil {
		if errors.Is(err, ErrTheFirstError) {
			err = orbitErrTheFirstError
		} else if errors.Is(err, ErrTheSecondError) {
			err = orbitErrTheSecondError
		} else if errors.Is(err, ErrTheThirdError) {
			err = orbitErrTheThirdError
		}
		return
	}
	r = ret
	return
}

func (v1 *s1Provider) c2(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	var args time.Time
	err = ad.Decode(&args)
	if err != nil {
		return
	}
	ret, err := v1.h.C2(ctx, s, args)
	if err != nil {
		if errors.Is(err, ErrTheFirstError) {
			err = orbitErrTheFirstError
		} else if errors.Is(err, ErrTheSecondError) {
			err = orbitErrTheSecondError
		} else if errors.Is(err, ErrTheThirdError) {
			err = orbitErrTheThirdError
		}
		return
	}
	r = ret
	return
}

func (v1 *s1Provider) c3(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	err = v1.h.C3(ctx, s)
	if err != nil {
		if errors.Is(err, ErrTheFirstError) {
			err = orbitErrTheFirstError
		} else if errors.Is(err, ErrTheSecondError) {
			err = orbitErrTheSecondError
		} else if errors.Is(err, ErrTheThirdError) {
			err = orbitErrTheThirdError
		}
		return
	}
	return
}

func (v1 *s1Provider) Rs1(ctx context.Context) (args *ArgsWriteChan, ret *RetReadChan, err error) {
	stream, err := v1.s.OpenStream(ctx, ServiceS1, S1Rs1)
	if err != nil {
		return
	}
	args = newArgsWriteChan(v1.s.CloserOneWay(), v1.s.StreamChanSize())
	args.OnClosing(func() error { return stream.Close() })
	go func() {
		closingChan := args.ClosingChan()
		codec := v1.s.Codec()
		for {
			select {
			case <-closingChan:
				return
			case arg := <-args.c:
				err := packet.WriteEncode(stream, arg, codec)
				if err != nil {
					if args.IsClosing() {
						err = nil
					}
					args.setError(err)
					return
				}
			}
		}
	}()
	ret = newRetReadChan(v1.s.CloserOneWay(), v1.s.StreamChanSize())
	ret.OnClosing(func() error { return stream.Close() })
	go func() {
		closingChan := ret.ClosingChan()
		codec := v1.s.Codec()
		for {
			var data *Ret
			err := packet.ReadDecode(stream, &data, codec)
			if err != nil {
				if ret.IsClosing() {
					err = nil
				}
				ret.setError(err)
				return
			}
			select {
			case <-closingChan:
				return
			case ret.c <- data:
			}
		}
	}()
	return
}

func (v1 *s1Provider) Rs2(ctx context.Context) (args *MapStringIntWriteChan, err error) {
	stream, err := v1.s.OpenStream(ctx, ServiceS1, S1Rs2)
	if err != nil {
		return
	}
	args = newMapStringIntWriteChan(v1.s.CloserOneWay(), v1.s.StreamChanSize())
	args.OnClosing(func() error { return stream.Close() })
	go func() {
		closingChan := args.ClosingChan()
		codec := v1.s.Codec()
		for {
			select {
			case <-closingChan:
				return
			case arg := <-args.c:
				err := packet.WriteEncode(stream, arg, codec)
				if err != nil {
					if args.IsClosing() {
						err = nil
					}
					args.setError(err)
					return
				}
			}
		}
	}()
	return
}

func (v1 *s1Provider) Rs3(ctx context.Context) (stream net.Conn, err error) {
	return v1.s.OpenStream(ctx, ServiceS1, S1Rs3)
}

func (v1 *s1Provider) s1(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	err = v1.h.S1(s, stream)
	if err != nil {
		return
	}
	return
}

func (v1 *s1Provider) s2(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	args := newStringReadChan(v1.s.CloserOneWay(), v1.s.StreamChanSize())
	go func() {
		closingChan := args.ClosingChan()
		codec := v1.s.Codec()
		for {
			var arg string
			err := packet.ReadDecode(stream, &arg, codec)
			if err != nil {
				if args.IsClosing() {
					err = nil
				}
				args.setError(err)
				return
			}
			select {
			case <-closingChan:
				return
			case args.c <- arg:
			}
		}
	}()

	err = v1.h.S2(s, args)
	if err != nil {
		return
	}
	return
}

func (v1 *s1Provider) s3(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	ret := newEn1WriteChan(v1.s.CloserOneWay(), v1.s.StreamChanSize())
	go func() {
		closingChan := ret.ClosingChan()
		codec := v1.s.Codec()
		for {
			select {
			case <-closingChan:
				return
			case data := <-ret.c:
				err := packet.WriteEncode(stream, data, codec)
				if err != nil {
					if ret.IsClosing() {
						err = nil
					}
					ret.setError(err)
					return
				}
			}
		}
	}()
	err = v1.h.S3(s, ret)
	if err != nil {
		return
	}
	return
}

// ---------------------

/* code generated by orbit */
package api

import (
	"context"
	"errors"
	"github.com/desertbit/closer/v3"
	"github.com/desertbit/orbit/pkg/orbit"
	"github.com/desertbit/orbit/pkg/packet"
	"net"
	"sync"
	"time"
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

//#####################//
//### Global Errors ###//
//#####################//

const (
	ErrCodeDatasetDoesNotExist = 101
	ErrCodeNotFound            = 100
)

var (
	ErrDatasetDoesNotExist      = errors.New("dataset does not exist")
	orbitErrDatasetDoesNotExist = orbit.Err(ErrDatasetDoesNotExist, ErrDatasetDoesNotExist.Error(), ErrCodeDatasetDoesNotExist)
	ErrNotFound                 = errors.New("not found")
	orbitErrNotFound            = orbit.Err(ErrNotFound, ErrNotFound.Error(), ErrCodeNotFound)
)

//####################//
//### Global Types ###//
//####################//

type Char struct {
	Lol string
}

//msgp:ignore CharWriteChan
type CharWriteChan struct {
	closer.Closer
	C   chan<- *Char
	c   chan *Char
	mx  sync.Mutex
	err error
}

func newCharWriteChan(cl closer.Closer) *CharWriteChan {
	c := &CharWriteChan{Closer: cl, c: make(chan *Char, 3)}
	c.C = c.c
	return c
}

func (c *CharWriteChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *CharWriteChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

//msgp:ignore CharReadChan
type CharReadChan struct {
	closer.Closer
	C   <-chan *Char
	c   chan *Char
	mx  sync.Mutex
	err error
}

func newCharReadChan(cl closer.Closer) *CharReadChan {
	c := &CharReadChan{Closer: cl, c: make(chan *Char, 3)}
	c.C = c.c
	return c
}

func (c *CharReadChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *CharReadChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

type Plate struct {
	Name    string
	Rect    *Rect
	Test    map[int]*Rect
	Test2   []*Rect
	Test3   []float32
	Test4   map[string]map[int][]*Rect
	Ts      time.Time
	Version int
}

//msgp:ignore PlateWriteChan
type PlateWriteChan struct {
	closer.Closer
	C   chan<- *Plate
	c   chan *Plate
	mx  sync.Mutex
	err error
}

func newPlateWriteChan(cl closer.Closer) *PlateWriteChan {
	c := &PlateWriteChan{Closer: cl, c: make(chan *Plate, 3)}
	c.C = c.c
	return c
}

func (c *PlateWriteChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *PlateWriteChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

//msgp:ignore PlateReadChan
type PlateReadChan struct {
	closer.Closer
	C   <-chan *Plate
	c   chan *Plate
	mx  sync.Mutex
	err error
}

func newPlateReadChan(cl closer.Closer) *PlateReadChan {
	c := &PlateReadChan{Closer: cl, c: make(chan *Plate, 3)}
	c.C = c.c
	return c
}

func (c *PlateReadChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *PlateReadChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

type Rect struct {
	C  *Char
	X1 float32
	X2 float32
	Y1 float32
	Y2 float32
}

//################//
//### Services ###//
//################//

// Example  ---------------------
// Errors
const (
	ErrCodeExampleAborted = 1
)

var (
	ErrExampleAborted      = errors.New("example aborted")
	orbitErrExampleAborted = orbit.Err(ErrExampleAborted, ErrExampleAborted.Error(), ErrCodeExampleAborted)
)

// Types
type ExampleChar struct {
	Lol string
}

//msgp:ignore ExampleCharWriteChan
type ExampleCharWriteChan struct {
	closer.Closer
	C   chan<- *ExampleChar
	c   chan *ExampleChar
	mx  sync.Mutex
	err error
}

func newExampleCharWriteChan(cl closer.Closer) *ExampleCharWriteChan {
	c := &ExampleCharWriteChan{Closer: cl, c: make(chan *ExampleChar, 3)}
	c.C = c.c
	return c
}

func (c *ExampleCharWriteChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *ExampleCharWriteChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

//msgp:ignore ExampleCharReadChan
type ExampleCharReadChan struct {
	closer.Closer
	C   <-chan *ExampleChar
	c   chan *ExampleChar
	mx  sync.Mutex
	err error
}

func newExampleCharReadChan(cl closer.Closer) *ExampleCharReadChan {
	c := &ExampleCharReadChan{Closer: cl, c: make(chan *ExampleChar, 3)}
	c.C = c.c
	return c
}

func (c *ExampleCharReadChan) setError(err error) {
	c.mx.Lock()
	c.err = err
	c.mx.Unlock()
	c.Close_()
}

func (c *ExampleCharReadChan) Err() (err error) {
	c.mx.Lock()
	err = c.err
	c.mx.Unlock()
	return
}

type ExampleRect struct {
	C  *ExampleChar
	X1 float32
	X2 float32
	Y1 float32
	Y2 float32
}

type ExampleTest3Args struct {
	C map[int][]*ExampleRect
	I int
	V float64
}

type ExampleTest3Ret struct {
	Lol string
}

// Service
const (
	ServiceExample = "Example"
	ExampleTest    = "Test"
	ExampleTest2   = "Test2"
	ExampleTest3   = "Test3"
	ExampleTest4   = "Test4"
	ExampleHello   = "Hello"
	ExampleHello2  = "Hello2"
	ExampleHello3  = "Hello3"
	ExampleHello4  = "Hello4"
)

type ExampleConsumerCaller interface {
	// Calls
	ExampleTest(ctx context.Context, args *Plate) (ret *ExampleRect, err error)
	ExampleTest2(ctx context.Context, args *ExampleRect) (err error)
	// Streams
	ExampleHello(ctx context.Context) (stream net.Conn, err error)
	ExampleHello2(ctx context.Context) (args *ExampleCharWriteChan, err error)
}

type ExampleConsumerHandler interface {
	// Calls
	ExampleTest3(ctx context.Context, s *orbit.Session, args *ExampleTest3Args) (ret *ExampleTest3Ret, err error)
	ExampleTest4(ctx context.Context, s *orbit.Session) (ret *ExampleRect, err error)
	// Streams
	ExampleHello3(s *orbit.Session, ret *PlateWriteChan) (err error)
	ExampleHello4(s *orbit.Session, args *ExampleCharReadChan, ret *PlateWriteChan) (err error)
}

type ExampleProviderCaller interface {
	// Calls
	ExampleTest3(ctx context.Context, args *ExampleTest3Args) (ret *ExampleTest3Ret, err error)
	ExampleTest4(ctx context.Context) (ret *ExampleRect, err error)
	// Streams
	ExampleHello3(ctx context.Context) (ret *PlateReadChan, err error)
	ExampleHello4(ctx context.Context) (args *ExampleCharWriteChan, ret *PlateReadChan, err error)
}

type ExampleProviderHandler interface {
	// Calls
	ExampleTest(ctx context.Context, s *orbit.Session, args *Plate) (ret *ExampleRect, err error)
	ExampleTest2(ctx context.Context, s *orbit.Session, args *ExampleRect) (err error)
	// Streams
	ExampleHello(s *orbit.Session, stream net.Conn) (err error)
	ExampleHello2(s *orbit.Session, args *ExampleCharReadChan) (err error)
}

type exampleConsumer struct {
	h ExampleConsumerHandler
	s *orbit.Session
}

func RegisterExampleConsumer(s *orbit.Session, h ExampleConsumerHandler) ExampleConsumerCaller {
	cc := &exampleConsumer{h: h, s: s}
	s.RegisterCall(ServiceExample, ExampleTest3, cc.exampleTest3)
	s.RegisterCall(ServiceExample, ExampleTest4, cc.exampleTest4)
	s.RegisterStream(ServiceExample, ExampleHello3, cc.exampleHello3)
	s.RegisterStream(ServiceExample, ExampleHello4, cc.exampleHello4)
	return cc
}

func (v1 *exampleConsumer) ExampleTest(ctx context.Context, args *Plate) (ret *ExampleRect, err error) {
	retData, err := v1.s.Call(ctx, ServiceExample, ExampleTest, args)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeDatasetDoesNotExist:
				err = ErrDatasetDoesNotExist
			case ErrCodeNotFound:
				err = ErrNotFound
			case ErrCodeExampleAborted:
				err = ErrExampleAborted
			}
		}
		return
	}
	ret = &ExampleRect{}
	err = retData.Decode(ret)
	if err != nil {
		return
	}
	return
}

func (v1 *exampleConsumer) ExampleTest2(ctx context.Context, args *ExampleRect) (err error) {
	_, err = v1.s.Call(ctx, ServiceExample, ExampleTest2, args)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeDatasetDoesNotExist:
				err = ErrDatasetDoesNotExist
			case ErrCodeNotFound:
				err = ErrNotFound
			case ErrCodeExampleAborted:
				err = ErrExampleAborted
			}
		}
		return
	}
	return
}

func (v1 *exampleConsumer) exampleTest3(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	args := &ExampleTest3Args{}
	err = ad.Decode(args)
	if err != nil {
		return
	}
	ret, err := v1.h.ExampleTest3(ctx, s, args)
	if err != nil {
		if errors.Is(err, ErrDatasetDoesNotExist) {
			err = orbitErrDatasetDoesNotExist
		} else if errors.Is(err, ErrNotFound) {
			err = orbitErrNotFound
		} else if errors.Is(err, ErrExampleAborted) {
			err = orbitErrExampleAborted
		}
		return
	}
	r = ret
	return
}

func (v1 *exampleConsumer) exampleTest4(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	ret, err := v1.h.ExampleTest4(ctx, s)
	if err != nil {
		if errors.Is(err, ErrDatasetDoesNotExist) {
			err = orbitErrDatasetDoesNotExist
		} else if errors.Is(err, ErrNotFound) {
			err = orbitErrNotFound
		} else if errors.Is(err, ErrExampleAborted) {
			err = orbitErrExampleAborted
		}
		return
	}
	r = ret
	return
}

func (v1 *exampleConsumer) ExampleHello(ctx context.Context) (stream net.Conn, err error) {
	return v1.s.OpenStream(ctx, ServiceExample, ExampleHello)
}

func (v1 *exampleConsumer) ExampleHello2(ctx context.Context) (args *ExampleCharWriteChan, err error) {
	stream, err := v1.s.OpenStream(ctx, ServiceExample, ExampleHello2)
	if err != nil {
		return
	}
	args = newExampleCharWriteChan(v1.s.CloserOneWay())
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

func (v1 *exampleConsumer) exampleHello3(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	ret := newPlateWriteChan(v1.s.CloserOneWay())
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
	err = v1.h.ExampleHello3(s, ret)
	if err != nil {
		return
	}
	return
}

func (v1 *exampleConsumer) exampleHello4(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	args := newExampleCharReadChan(v1.s.CloserOneWay())
	go func() {
		closingChan := args.ClosingChan()
		codec := v1.s.Codec()
		for {
			arg := &ExampleChar{}
			err := packet.ReadDecode(stream, arg, codec)
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

	ret := newPlateWriteChan(v1.s.CloserOneWay())
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
	err = v1.h.ExampleHello4(s, args, ret)
	if err != nil {
		return
	}
	return
}

type exampleProvider struct {
	h ExampleProviderHandler
	s *orbit.Session
}

func RegisterExampleProvider(s *orbit.Session, h ExampleProviderHandler) ExampleProviderCaller {
	cc := &exampleProvider{h: h, s: s}
	s.RegisterCall(ServiceExample, ExampleTest, cc.exampleTest)
	s.RegisterCall(ServiceExample, ExampleTest2, cc.exampleTest2)
	s.RegisterStream(ServiceExample, ExampleHello, cc.exampleHello)
	s.RegisterStream(ServiceExample, ExampleHello2, cc.exampleHello2)
	return cc
}

func (v1 *exampleProvider) ExampleTest3(ctx context.Context, args *ExampleTest3Args) (ret *ExampleTest3Ret, err error) {
	retData, err := v1.s.Call(ctx, ServiceExample, ExampleTest3, args)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeDatasetDoesNotExist:
				err = ErrDatasetDoesNotExist
			case ErrCodeNotFound:
				err = ErrNotFound
			case ErrCodeExampleAborted:
				err = ErrExampleAborted
			}
		}
		return
	}
	ret = &ExampleTest3Ret{}
	err = retData.Decode(ret)
	if err != nil {
		return
	}
	return
}

func (v1 *exampleProvider) ExampleTest4(ctx context.Context) (ret *ExampleRect, err error) {
	retData, err := v1.s.Call(ctx, ServiceExample, ExampleTest4, nil)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeDatasetDoesNotExist:
				err = ErrDatasetDoesNotExist
			case ErrCodeNotFound:
				err = ErrNotFound
			case ErrCodeExampleAborted:
				err = ErrExampleAborted
			}
		}
		return
	}
	ret = &ExampleRect{}
	err = retData.Decode(ret)
	if err != nil {
		return
	}
	return
}

func (v1 *exampleProvider) exampleTest(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	args := &Plate{}
	err = ad.Decode(args)
	if err != nil {
		return
	}
	ret, err := v1.h.ExampleTest(ctx, s, args)
	if err != nil {
		if errors.Is(err, ErrDatasetDoesNotExist) {
			err = orbitErrDatasetDoesNotExist
		} else if errors.Is(err, ErrNotFound) {
			err = orbitErrNotFound
		} else if errors.Is(err, ErrExampleAborted) {
			err = orbitErrExampleAborted
		}
		return
	}
	r = ret
	return
}

func (v1 *exampleProvider) exampleTest2(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	args := &ExampleRect{}
	err = ad.Decode(args)
	if err != nil {
		return
	}
	err = v1.h.ExampleTest2(ctx, s, args)
	if err != nil {
		if errors.Is(err, ErrDatasetDoesNotExist) {
			err = orbitErrDatasetDoesNotExist
		} else if errors.Is(err, ErrNotFound) {
			err = orbitErrNotFound
		} else if errors.Is(err, ErrExampleAborted) {
			err = orbitErrExampleAborted
		}
		return
	}
	return
}

func (v1 *exampleProvider) ExampleHello3(ctx context.Context) (ret *PlateReadChan, err error) {
	stream, err := v1.s.OpenStream(ctx, ServiceExample, ExampleHello3)
	if err != nil {
		return
	}
	ret = newPlateReadChan(v1.s.CloserOneWay())
	ret.OnClosing(func() error { return stream.Close() })
	go func() {
		closingChan := ret.ClosingChan()
		codec := v1.s.Codec()
		for {
			data := &Plate{}
			err := packet.ReadDecode(stream, data, codec)
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

func (v1 *exampleProvider) ExampleHello4(ctx context.Context) (args *ExampleCharWriteChan, ret *PlateReadChan, err error) {
	stream, err := v1.s.OpenStream(ctx, ServiceExample, ExampleHello4)
	if err != nil {
		return
	}
	args = newExampleCharWriteChan(v1.s.CloserOneWay())
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
	ret = newPlateReadChan(v1.s.CloserOneWay())
	ret.OnClosing(func() error { return stream.Close() })
	go func() {
		closingChan := ret.ClosingChan()
		codec := v1.s.Codec()
		for {
			data := &Plate{}
			err := packet.ReadDecode(stream, data, codec)
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

func (v1 *exampleProvider) exampleHello(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	err = v1.h.ExampleHello(s, stream)
	if err != nil {
		return
	}
	return
}

func (v1 *exampleProvider) exampleHello2(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	args := newExampleCharReadChan(v1.s.CloserOneWay())
	go func() {
		closingChan := args.ClosingChan()
		codec := v1.s.Codec()
		for {
			arg := &ExampleChar{}
			err := packet.ReadDecode(stream, arg, codec)
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

	err = v1.h.ExampleHello2(s, args)
	if err != nil {
		return
	}
	return
}

// ---------------------

// Trainer  ---------------------
// Errors
const (
	ErrCodeTrainerAborted = 1
)

var (
	ErrTrainerAborted      = errors.New("trainer aborted")
	orbitErrTrainerAborted = orbit.Err(ErrTrainerAborted, ErrTrainerAborted.Error(), ErrCodeTrainerAborted)
)

// Types
// Service
const (
	ServiceTrainer  = "Trainer"
	TrainerStart    = "Start"
	TrainerUpdate   = "Update"
	TrainerUpload   = "Upload"
	TrainerDownload = "Download"
	TrainerSend     = "Send"
	TrainerReceive  = "Receive"
	TrainerLink     = "Link"
)

type TrainerConsumerCaller interface {
	// Calls
	TrainerStart(ctx context.Context, args *Plate) (err error)
	TrainerUpdate(ctx context.Context, args *Char) (ret *Char, err error)
	// Streams
	TrainerUpload(ctx context.Context) (stream net.Conn, err error)
	TrainerDownload(ctx context.Context) (args *CharWriteChan, err error)
}

type TrainerConsumerHandler interface {
	// Streams
	TrainerSend(s *orbit.Session, args *PlateReadChan) (err error)
	TrainerReceive(s *orbit.Session, ret *CharWriteChan) (err error)
	TrainerLink(s *orbit.Session, args *PlateReadChan, ret *CharWriteChan) (err error)
}

type TrainerProviderCaller interface {
	// Streams
	TrainerSend(ctx context.Context) (args *PlateWriteChan, err error)
	TrainerReceive(ctx context.Context) (ret *CharReadChan, err error)
	TrainerLink(ctx context.Context) (args *PlateWriteChan, ret *CharReadChan, err error)
}

type TrainerProviderHandler interface {
	// Calls
	TrainerStart(ctx context.Context, s *orbit.Session, args *Plate) (err error)
	TrainerUpdate(ctx context.Context, s *orbit.Session, args *Char) (ret *Char, err error)
	// Streams
	TrainerUpload(s *orbit.Session, stream net.Conn) (err error)
	TrainerDownload(s *orbit.Session, args *CharReadChan) (err error)
}

type trainerConsumer struct {
	h TrainerConsumerHandler
	s *orbit.Session
}

func RegisterTrainerConsumer(s *orbit.Session, h TrainerConsumerHandler) TrainerConsumerCaller {
	cc := &trainerConsumer{h: h, s: s}
	s.RegisterStream(ServiceTrainer, TrainerSend, cc.trainerSend)
	s.RegisterStream(ServiceTrainer, TrainerReceive, cc.trainerReceive)
	s.RegisterStream(ServiceTrainer, TrainerLink, cc.trainerLink)
	return cc
}

func (v1 *trainerConsumer) TrainerStart(ctx context.Context, args *Plate) (err error) {
	_, err = v1.s.Call(ctx, ServiceTrainer, TrainerStart, args)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeDatasetDoesNotExist:
				err = ErrDatasetDoesNotExist
			case ErrCodeNotFound:
				err = ErrNotFound
			case ErrCodeTrainerAborted:
				err = ErrTrainerAborted
			}
		}
		return
	}
	return
}

func (v1 *trainerConsumer) TrainerUpdate(ctx context.Context, args *Char) (ret *Char, err error) {
	retData, err := v1.s.Call(ctx, ServiceTrainer, TrainerUpdate, args)
	if err != nil {
		var cErr *orbit.ErrorCode
		if errors.As(err, &cErr) {
			switch cErr.Code {
			case ErrCodeDatasetDoesNotExist:
				err = ErrDatasetDoesNotExist
			case ErrCodeNotFound:
				err = ErrNotFound
			case ErrCodeTrainerAborted:
				err = ErrTrainerAborted
			}
		}
		return
	}
	ret = &Char{}
	err = retData.Decode(ret)
	if err != nil {
		return
	}
	return
}

func (v1 *trainerConsumer) TrainerUpload(ctx context.Context) (stream net.Conn, err error) {
	return v1.s.OpenStream(ctx, ServiceTrainer, TrainerUpload)
}

func (v1 *trainerConsumer) TrainerDownload(ctx context.Context) (args *CharWriteChan, err error) {
	stream, err := v1.s.OpenStream(ctx, ServiceTrainer, TrainerDownload)
	if err != nil {
		return
	}
	args = newCharWriteChan(v1.s.CloserOneWay())
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

func (v1 *trainerConsumer) trainerSend(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	args := newPlateReadChan(v1.s.CloserOneWay())
	go func() {
		closingChan := args.ClosingChan()
		codec := v1.s.Codec()
		for {
			arg := &Plate{}
			err := packet.ReadDecode(stream, arg, codec)
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

	err = v1.h.TrainerSend(s, args)
	if err != nil {
		return
	}
	return
}

func (v1 *trainerConsumer) trainerReceive(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	ret := newCharWriteChan(v1.s.CloserOneWay())
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
	err = v1.h.TrainerReceive(s, ret)
	if err != nil {
		return
	}
	return
}

func (v1 *trainerConsumer) trainerLink(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	args := newPlateReadChan(v1.s.CloserOneWay())
	go func() {
		closingChan := args.ClosingChan()
		codec := v1.s.Codec()
		for {
			arg := &Plate{}
			err := packet.ReadDecode(stream, arg, codec)
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

	ret := newCharWriteChan(v1.s.CloserOneWay())
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
	err = v1.h.TrainerLink(s, args, ret)
	if err != nil {
		return
	}
	return
}

type trainerProvider struct {
	h TrainerProviderHandler
	s *orbit.Session
}

func RegisterTrainerProvider(s *orbit.Session, h TrainerProviderHandler) TrainerProviderCaller {
	cc := &trainerProvider{h: h, s: s}
	s.RegisterCall(ServiceTrainer, TrainerStart, cc.trainerStart)
	s.RegisterCall(ServiceTrainer, TrainerUpdate, cc.trainerUpdate)
	s.RegisterStream(ServiceTrainer, TrainerUpload, cc.trainerUpload)
	s.RegisterStream(ServiceTrainer, TrainerDownload, cc.trainerDownload)
	return cc
}

func (v1 *trainerProvider) trainerStart(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	args := &Plate{}
	err = ad.Decode(args)
	if err != nil {
		return
	}
	err = v1.h.TrainerStart(ctx, s, args)
	if err != nil {
		if errors.Is(err, ErrDatasetDoesNotExist) {
			err = orbitErrDatasetDoesNotExist
		} else if errors.Is(err, ErrNotFound) {
			err = orbitErrNotFound
		} else if errors.Is(err, ErrTrainerAborted) {
			err = orbitErrTrainerAborted
		}
		return
	}
	return
}

func (v1 *trainerProvider) trainerUpdate(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
	args := &Char{}
	err = ad.Decode(args)
	if err != nil {
		return
	}
	ret, err := v1.h.TrainerUpdate(ctx, s, args)
	if err != nil {
		if errors.Is(err, ErrDatasetDoesNotExist) {
			err = orbitErrDatasetDoesNotExist
		} else if errors.Is(err, ErrNotFound) {
			err = orbitErrNotFound
		} else if errors.Is(err, ErrTrainerAborted) {
			err = orbitErrTrainerAborted
		}
		return
	}
	r = ret
	return
}

func (v1 *trainerProvider) TrainerSend(ctx context.Context) (args *PlateWriteChan, err error) {
	stream, err := v1.s.OpenStream(ctx, ServiceTrainer, TrainerSend)
	if err != nil {
		return
	}
	args = newPlateWriteChan(v1.s.CloserOneWay())
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

func (v1 *trainerProvider) TrainerReceive(ctx context.Context) (ret *CharReadChan, err error) {
	stream, err := v1.s.OpenStream(ctx, ServiceTrainer, TrainerReceive)
	if err != nil {
		return
	}
	ret = newCharReadChan(v1.s.CloserOneWay())
	ret.OnClosing(func() error { return stream.Close() })
	go func() {
		closingChan := ret.ClosingChan()
		codec := v1.s.Codec()
		for {
			data := &Char{}
			err := packet.ReadDecode(stream, data, codec)
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

func (v1 *trainerProvider) TrainerLink(ctx context.Context) (args *PlateWriteChan, ret *CharReadChan, err error) {
	stream, err := v1.s.OpenStream(ctx, ServiceTrainer, TrainerLink)
	if err != nil {
		return
	}
	args = newPlateWriteChan(v1.s.CloserOneWay())
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
	ret = newCharReadChan(v1.s.CloserOneWay())
	ret.OnClosing(func() error { return stream.Close() })
	go func() {
		closingChan := ret.ClosingChan()
		codec := v1.s.Codec()
		for {
			data := &Char{}
			err := packet.ReadDecode(stream, data, codec)
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

func (v1 *trainerProvider) trainerUpload(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	err = v1.h.TrainerUpload(s, stream)
	if err != nil {
		return
	}
	return
}

func (v1 *trainerProvider) trainerDownload(s *orbit.Session, stream net.Conn) (err error) {
	defer stream.Close()
	args := newCharReadChan(v1.s.CloserOneWay())
	go func() {
		closingChan := args.ClosingChan()
		codec := v1.s.Codec()
		for {
			arg := &Char{}
			err := packet.ReadDecode(stream, arg, codec)
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

	err = v1.h.TrainerDownload(s, args)
	if err != nil {
		return
	}
	return
}

// ---------------------

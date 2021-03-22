/* code generated by orbit */
package hello

import (
	context "context"
	errors "errors"
	fmt "fmt"
	closer "github.com/desertbit/closer/v3"
	oclient "github.com/desertbit/orbit/pkg/client"
	codec "github.com/desertbit/orbit/pkg/codec"
	packet "github.com/desertbit/orbit/pkg/packet"
	oservice "github.com/desertbit/orbit/pkg/service"
	transport "github.com/desertbit/orbit/pkg/transport"
	validator "github.com/go-playground/validator/v10"
	io "io"
	net "net"
	strings "strings"
	sync "sync"
	time "time"
)

// Ensure that all imports are used.
var (
	_ context.Context
	_ = errors.New("")
	_ = fmt.Sprint()
	_ io.Closer
	_ net.Conn
	_ time.Time
	_ strings.Builder
	_ sync.Locker
	_ oclient.Client
	_ closer.Closer
	_ codec.Codec
	_ = packet.MaxSize
	_ oservice.Service
	_ transport.Transport
	_ validator.StructLevel
)

//##############//
//### Errors ###//
//##############//

var ErrClosed = errors.New("closed")

const (
	ErrCodeIAmAnError  = 2
	ErrCodeThisIsATest = 1
)

var (
	ErrIAmAnError  = errors.New("i am an error")
	ErrThisIsATest = errors.New("this is a test")
)

func _clientErrorCheck(err error) error {
	var cErr oclient.Error
	if errors.As(err, &cErr) {
		switch cErr.Code() {
		case ErrCodeIAmAnError:
			return ErrIAmAnError
		case ErrCodeThisIsATest:
			return ErrThisIsATest
		}
	}
	return err
}

func _serviceErrorCheck(err error) error {
	if errors.Is(err, ErrIAmAnError) {
		return oservice.NewError(err, ErrIAmAnError.Error(), ErrCodeIAmAnError)
	} else if errors.Is(err, ErrThisIsATest) {
		return oservice.NewError(err, ErrThisIsATest.Error(), ErrCodeThisIsATest)
	}
	return err
}

func _valErrCheck(err error) error {
	if vErrs, ok := err.(validator.ValidationErrors); ok {
		var errMsg strings.Builder
		for _, err := range vErrs {
			errMsg.WriteString(fmt.Sprintf("[name: '%s', value: '%s', tag: '%s']", err.StructNamespace(), err.Value(), err.Tag()))
		}
		return errors.New(errMsg.String())
	}
	return err
}

var validate = validator.New()

//#############//
//### Types ###//
//#############//

type BidirectionalArg struct {
	Question string
}

type BidirectionalRet struct {
	Answer string
}

type ClockTimeRet struct {
	Ts time.Time `validate:"required"`
}

type Info struct {
	Name    string `validate:"required,min=1"`
	Age     int    `validate:"required,min=1,max=155"`
	Locale  string `validate:"required,len=5"`
	Address string `validate:"omitempty"`
}

type SayHiArg struct {
	Name string `validate:"required,min=1"`
	Ts   time.Time
}

type SayHiRet struct {
	Res []int `validate:"required,min=1"`
}

type TestArg struct {
	S string
}

type TestRet struct {
	Name string `validate:"required,min=1"`
	Ts   time.Time
}

//msgp:ignore TimeStreamClientStream
type TimeStreamClientStream struct {
	oclient.TypedStreamCloser
	stream oclient.TypedWStream
}

func newTimeStreamClientStream(s oclient.TypedWStream) *TimeStreamClientStream {
	return &TimeStreamClientStream{TypedStreamCloser: s, stream: s}
}

func (v1 *TimeStreamClientStream) Write(arg Info) (err error) {
	err = v1.stream.Write(arg)
	if err != nil {
		err = _clientErrorCheck(err)
		if errors.Is(err, oclient.ErrClosed) {
			err = ErrClosed
		}
		return
	}
	return
}

//msgp:ignore TimeStreamServiceStream
type TimeStreamServiceStream struct {
	oservice.TypedStreamCloser
	stream oservice.TypedRStream
}

func newTimeStreamServiceStream(s oservice.TypedRStream) *TimeStreamServiceStream {
	return &TimeStreamServiceStream{TypedStreamCloser: s, stream: s}
}

func (v1 *TimeStreamServiceStream) Read() (arg Info, err error) {
	err = v1.stream.Read(&arg)
	if err != nil {
		err = _serviceErrorCheck(err)
		if errors.Is(err, oservice.ErrClosed) {
			err = ErrClosed
		}
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	return
}

//msgp:ignore ClockTimeClientStream
type ClockTimeClientStream struct {
	oclient.TypedStreamCloser
	stream oclient.TypedRStream
}

func newClockTimeClientStream(s oclient.TypedRStream) *ClockTimeClientStream {
	return &ClockTimeClientStream{TypedStreamCloser: s, stream: s}
}

func (v1 *ClockTimeClientStream) Read() (ret ClockTimeRet, err error) {
	err = v1.stream.Read(&ret)
	if err != nil {
		err = _clientErrorCheck(err)
		if errors.Is(err, oclient.ErrClosed) {
			err = ErrClosed
		}
		return
	}
	err = validate.Struct(ret)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	return
}

//msgp:ignore ClockTimeServiceStream
type ClockTimeServiceStream struct {
	oservice.TypedStreamCloser
	stream oservice.TypedWStream
}

func newClockTimeServiceStream(s oservice.TypedWStream) *ClockTimeServiceStream {
	return &ClockTimeServiceStream{TypedStreamCloser: s, stream: s}
}

func (v1 *ClockTimeServiceStream) Write(ret ClockTimeRet) (err error) {
	err = v1.stream.Write(ret)
	if err != nil {
		err = _serviceErrorCheck(err)
		if errors.Is(err, oservice.ErrClosed) {
			err = ErrClosed
		}
		return
	}
	return
}

//msgp:ignore BidirectionalClientStream
type BidirectionalClientStream struct {
	oclient.TypedStreamCloser
	stream oclient.TypedRWStream
}

func newBidirectionalClientStream(s oclient.TypedRWStream) *BidirectionalClientStream {
	return &BidirectionalClientStream{TypedStreamCloser: s, stream: s}
}

func (v1 *BidirectionalClientStream) Read() (ret BidirectionalRet, err error) {
	err = v1.stream.Read(&ret)
	if err != nil {
		err = _clientErrorCheck(err)
		if errors.Is(err, oclient.ErrClosed) {
			err = ErrClosed
		}
		return
	}
	err = validate.Struct(ret)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	return
}

func (v1 *BidirectionalClientStream) Write(arg BidirectionalArg) (err error) {
	err = v1.stream.Write(arg)
	if err != nil {
		err = _clientErrorCheck(err)
		if errors.Is(err, oclient.ErrClosed) {
			err = ErrClosed
		}
		return
	}
	return
}

//msgp:ignore BidirectionalServiceStream
type BidirectionalServiceStream struct {
	oservice.TypedStreamCloser
	stream oservice.TypedRWStream
}

func newBidirectionalServiceStream(s oservice.TypedRWStream) *BidirectionalServiceStream {
	return &BidirectionalServiceStream{TypedStreamCloser: s, stream: s}
}

func (v1 *BidirectionalServiceStream) Read() (arg BidirectionalArg, err error) {
	err = v1.stream.Read(&arg)
	if err != nil {
		err = _serviceErrorCheck(err)
		if errors.Is(err, oservice.ErrClosed) {
			err = ErrClosed
		}
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	return
}

func (v1 *BidirectionalServiceStream) Write(ret BidirectionalRet) (err error) {
	err = v1.stream.Write(ret)
	if err != nil {
		err = _serviceErrorCheck(err)
		if errors.Is(err, oservice.ErrClosed) {
			err = ErrClosed
		}
		return
	}
	return
}

//#############//
//### Enums ###//
//#############//

type Vehicle int

const (
	Car    Vehicle = 1
	Pickup Vehicle = 2
)

//###############//
//### Service ###//
//###############//

const (
	CallIDSayHi           = "SayHi"
	CallIDTest            = "Test"
	StreamIDLul           = "Lul"
	StreamIDTimeStream    = "TimeStream"
	StreamIDClockTime     = "ClockTime"
	StreamIDBidirectional = "Bidirectional"
)

type Client interface {
	closer.Closer
	StateChan() <-chan oclient.State
	// Calls
	SayHi(ctx context.Context, arg SayHiArg) (ret SayHiRet, err error)
	Test(ctx context.Context, arg TestArg) (ret TestRet, err error)
	// Streams
	Lul(ctx context.Context) (stream transport.Stream, err error)
	TimeStream(ctx context.Context) (stream *TimeStreamClientStream, err error)
	ClockTime(ctx context.Context) (stream *ClockTimeClientStream, err error)
	Bidirectional(ctx context.Context) (stream *BidirectionalClientStream, err error)
}

type Service interface {
	closer.Closer
	Run() error
}

type ServiceHandler interface {
	// Calls
	SayHi(ctx oservice.Context, arg SayHiArg) (ret SayHiRet, err error)
	Test(ctx oservice.Context, arg TestArg) (ret TestRet, err error)
	// Streams
	Lul(ctx oservice.Context, stream transport.Stream)
	TimeStream(ctx oservice.Context, stream *TimeStreamServiceStream) error
	ClockTime(ctx oservice.Context, stream *ClockTimeServiceStream) error
	Bidirectional(ctx oservice.Context, stream *BidirectionalServiceStream) error
}

type client struct {
	oclient.Client
	codec             codec.Codec
	callTimeout       time.Duration
	streamInitTimeout time.Duration
	maxArgSize        int
	maxRetSize        int
}

func NewClient(opts *oclient.Options) (c Client, err error) {
	oc, err := oclient.New(opts)
	if err != nil {
		return
	}
	c = &client{Client: oc, codec: opts.Codec, callTimeout: opts.CallTimeout, streamInitTimeout: opts.StreamInitTimeout, maxArgSize: opts.MaxArgSize, maxRetSize: opts.MaxRetSize}
	return
}

func (v1 *client) StateChan() <-chan oclient.State {
	return v1.Client.StateChan()
}

func (v1 *client) SayHi(ctx context.Context, arg SayHiArg) (ret SayHiRet, err error) {
	if v1.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.callTimeout)
		defer cancel()
	}
	err = v1.Call(ctx, CallIDSayHi, arg, &ret)
	if err != nil {
		err = _clientErrorCheck(err)
		return
	}
	err = validate.Struct(ret)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	return
}

func (v1 *client) Test(ctx context.Context, arg TestArg) (ret TestRet, err error) {
	ctx, cancel := context.WithTimeout(ctx, 500000000*time.Nanosecond)
	defer cancel()
	err = v1.AsyncCall(ctx, CallIDTest, arg, &ret, oclient.DefaultMaxSize, 10240)
	if err != nil {
		err = _clientErrorCheck(err)
		return
	}
	err = validate.Struct(ret)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	return
}

func (v1 *client) Lul(ctx context.Context) (stream transport.Stream, err error) {
	if v1.streamInitTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.streamInitTimeout)
		defer cancel()
	}
	stream, err = v1.Stream(ctx, StreamIDLul)
	if err != nil {
		return
	}
	return
}

func (v1 *client) TimeStream(ctx context.Context) (stream *TimeStreamClientStream, err error) {
	if v1.streamInitTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.streamInitTimeout)
		defer cancel()
	}
	str, err := v1.TypedWStream(ctx, StreamIDTimeStream, oclient.DefaultMaxSize)
	if err != nil {
		return
	}
	stream = newTimeStreamClientStream(str)
	return
}

func (v1 *client) ClockTime(ctx context.Context) (stream *ClockTimeClientStream, err error) {
	if v1.streamInitTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.streamInitTimeout)
		defer cancel()
	}
	str, err := v1.TypedRStream(ctx, StreamIDClockTime, oclient.DefaultMaxSize)
	if err != nil {
		return
	}
	stream = newClockTimeClientStream(str)
	return
}

func (v1 *client) Bidirectional(ctx context.Context) (stream *BidirectionalClientStream, err error) {
	if v1.streamInitTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.streamInitTimeout)
		defer cancel()
	}
	str, err := v1.TypedRWStream(ctx, StreamIDBidirectional, 102400, oclient.DefaultMaxSize)
	if err != nil {
		return
	}
	stream = newBidirectionalClientStream(str)
	return
}

type service struct {
	oservice.Service
	h          ServiceHandler
	codec      codec.Codec
	maxArgSize int
	maxRetSize int
}

func NewService(h ServiceHandler, opts *oservice.Options) (s Service, err error) {
	os, err := oservice.New(opts)
	if err != nil {
		return
	}
	srvc := &service{Service: os, h: h, codec: opts.Codec, maxArgSize: opts.MaxArgSize, maxRetSize: opts.MaxRetSize}
	// Ensure usage.
	_ = srvc
	os.RegisterCall(CallIDSayHi, srvc.sayHi, oservice.DefaultTimeout)
	os.RegisterAsyncCall(CallIDTest, srvc.test, 500000000*time.Nanosecond, oservice.DefaultMaxSize, 10240)
	os.RegisterStream(StreamIDLul, srvc.lul)
	os.RegisterTypedRStream(StreamIDTimeStream, srvc.timeStream, oservice.DefaultMaxSize)
	os.RegisterTypedWStream(StreamIDClockTime, srvc.clockTime, oservice.DefaultMaxSize)
	os.RegisterTypedRWStream(StreamIDBidirectional, srvc.bidirectional, 102400, oservice.DefaultMaxSize)
	s = os
	return
}

func (v1 *service) sayHi(ctx oservice.Context, argData []byte) (retData interface{}, err error) {
	var arg SayHiArg
	err = v1.codec.Decode(argData, &arg)
	if err != nil {
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	ret, err := v1.h.SayHi(ctx, arg)
	if err != nil {
		err = _serviceErrorCheck(err)
		return
	}
	retData = &ret
	return
}

func (v1 *service) test(ctx oservice.Context, argData []byte) (retData interface{}, err error) {
	var arg TestArg
	err = v1.codec.Decode(argData, &arg)
	if err != nil {
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	ret, err := v1.h.Test(ctx, arg)
	if err != nil {
		err = _serviceErrorCheck(err)
		return
	}
	retData = &ret
	return
}

func (v1 *service) lul(ctx oservice.Context, stream transport.Stream) {
	v1.h.Lul(ctx, stream)
}

func (v1 *service) timeStream(ctx oservice.Context, stream oservice.TypedRStream) (err error) {
	err = v1.h.TimeStream(ctx, newTimeStreamServiceStream(stream))
	if err != nil {
		err = _serviceErrorCheck(err)
	}
	return
}

func (v1 *service) clockTime(ctx oservice.Context, stream oservice.TypedWStream) (err error) {
	err = v1.h.ClockTime(ctx, newClockTimeServiceStream(stream))
	if err != nil {
		err = _serviceErrorCheck(err)
	}
	return
}

func (v1 *service) bidirectional(ctx oservice.Context, stream oservice.TypedRWStream) (err error) {
	err = v1.h.Bidirectional(ctx, newBidirectionalServiceStream(stream))
	if err != nil {
		err = _serviceErrorCheck(err)
	}
	return
}

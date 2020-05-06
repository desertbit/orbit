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
		return oservice.Err(err, ErrIAmAnError.Error(), ErrCodeIAmAnError)
	} else if errors.Is(err, ErrThisIsATest) {
		return oservice.Err(err, ErrThisIsATest.Error(), ErrCodeThisIsATest)
	}
	return err
}

func _valErrCheck(err error) error {
	if vErrs, ok := err.(validator.ValidationErrors); ok {
		var errMsg strings.Builder
		for _, err := range vErrs {
			errMsg.WriteString(fmt.Sprintf("-> name: '%s', value: '%s', tag: '%s'", err.StructNamespace(), err.Value(), err.Tag()))
		}
		return errors.New(errMsg.String())
	}
	return err
}

var validate = validator.New()

//#############//
//### Types ###//
//#############//

type ClockTimeRet struct {
	Ts time.Time `validation:"required`
}

type Info struct {
	Name    string `validation:"required,min=1`
	Age     int    `validation:"required,min=1,max=155`
	Locale  string `validation:"required,len=5`
	Address string `validation:"omitempty`
}

type SayHiArg struct {
	Name string `validation:"required,min=1"`
	Ts   time.Time
}

type SayHiRet struct {
	Res []int `validation:"required,min=1"`
}

type TestArg struct {
	S string
}

type TestRet struct {
	Name string `validation:"required,min=1`
	Ts   time.Time
}

//msgp:ignore InfoReadStream
type InfoReadStream struct {
	closer.Closer
	stream  transport.Stream
	codec   codec.Codec
	maxSize int
}

func newInfoReadStream(cl closer.Closer, s transport.Stream, cc codec.Codec, ms int) *InfoReadStream {
	return &InfoReadStream{Closer: cl, stream: s, codec: cc, maxSize: ms}
}

func (v1 *InfoReadStream) Read() (arg Info, err error) {
	if v1.IsClosing() {
		err = ErrClosed
		return
	}
	arg = Info{}
	err = packet.ReadDecode(v1.stream, &arg, v1.codec, v1.maxSize)
	if err != nil {
		if errors.Is(err, packet.ErrZeroData) || errors.Is(err, io.EOF) || v1.stream.IsClosed() {
			err = ErrClosed
		}
		v1.Close_()
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	return
}

//msgp:ignore InfoWriteStream
type InfoWriteStream struct {
	closer.Closer
	stream  transport.Stream
	codec   codec.Codec
	maxSize int
}

func newInfoWriteStream(cl closer.Closer, s transport.Stream, cc codec.Codec, ms int) *InfoWriteStream {
	cl.OnClosing(func() error { return packet.Write(s, nil, 0) })
	return &InfoWriteStream{Closer: cl, stream: s, codec: cc, maxSize: ms}
}

func (v1 *InfoWriteStream) Write(ret Info) (err error) {
	if v1.IsClosing() {
		err = ErrClosed
		return
	}
	err = packet.WriteEncode(v1.stream, ret, v1.codec, v1.maxSize)
	if err != nil {
		if errors.Is(err, io.EOF) || v1.stream.IsClosed() {
			v1.Close_()
			return ErrClosed
		}
	}
	return
}

//msgp:ignore ClockTimeRetReadStream
type ClockTimeRetReadStream struct {
	closer.Closer
	stream  transport.Stream
	codec   codec.Codec
	maxSize int
}

func newClockTimeRetReadStream(cl closer.Closer, s transport.Stream, cc codec.Codec, ms int) *ClockTimeRetReadStream {
	return &ClockTimeRetReadStream{Closer: cl, stream: s, codec: cc, maxSize: ms}
}

func (v1 *ClockTimeRetReadStream) Read() (arg ClockTimeRet, err error) {
	if v1.IsClosing() {
		err = ErrClosed
		return
	}
	arg = ClockTimeRet{}
	err = packet.ReadDecode(v1.stream, &arg, v1.codec, v1.maxSize)
	if err != nil {
		if errors.Is(err, packet.ErrZeroData) || errors.Is(err, io.EOF) || v1.stream.IsClosed() {
			err = ErrClosed
		}
		v1.Close_()
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	return
}

//msgp:ignore ClockTimeRetWriteStream
type ClockTimeRetWriteStream struct {
	closer.Closer
	stream  transport.Stream
	codec   codec.Codec
	maxSize int
}

func newClockTimeRetWriteStream(cl closer.Closer, s transport.Stream, cc codec.Codec, ms int) *ClockTimeRetWriteStream {
	cl.OnClosing(func() error { return packet.Write(s, nil, 0) })
	return &ClockTimeRetWriteStream{Closer: cl, stream: s, codec: cc, maxSize: ms}
}

func (v1 *ClockTimeRetWriteStream) Write(ret ClockTimeRet) (err error) {
	if v1.IsClosing() {
		err = ErrClosed
		return
	}
	err = packet.WriteEncode(v1.stream, ret, v1.codec, v1.maxSize)
	if err != nil {
		if errors.Is(err, io.EOF) || v1.stream.IsClosed() {
			v1.Close_()
			return ErrClosed
		}
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

// CallIDs
const (
	SayHi = "SayHi"
	Test  = "Test"
	// StreamIDs
	Lul        = "Lul"
	TimeStream = "TimeStream"
	ClockTime  = "ClockTime"
)

type Client interface {
	closer.Closer
	// Calls
	SayHi(ctx context.Context, arg SayHiArg) (ret SayHiRet, err error)
	Test(ctx context.Context, arg TestArg) (ret TestRet, err error)
	// Streams
	Lul(ctx context.Context) (stream transport.Stream, err error)
	TimeStream(ctx context.Context) (arg *InfoWriteStream, err error)
	ClockTime(ctx context.Context) (ret *ClockTimeRetReadStream, err error)
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
	TimeStream(ctx oservice.Context, arg *InfoReadStream)
	ClockTime(ctx oservice.Context, ret *ClockTimeRetWriteStream)
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

func (v1 *client) SayHi(ctx context.Context, arg SayHiArg) (ret SayHiRet, err error) {
	if v1.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.callTimeout)
		defer cancel()
	}
	ret = SayHiRet{}
	err = v1.Call(ctx, SayHi, arg, &ret)
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
	ret = TestRet{}
	err = v1.AsyncCall(ctx, Test, arg, &ret, oclient.DefaultMaxSize, 10240)
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
	stream, err = v1.Stream(ctx, Lul)
	if err != nil {
		return
	}
	return
}

func (v1 *client) TimeStream(ctx context.Context) (arg *InfoWriteStream, err error) {
	if v1.streamInitTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.streamInitTimeout)
		defer cancel()
	}
	stream, err := v1.Stream(ctx, TimeStream)
	if err != nil {
		return
	}
	arg = newInfoWriteStream(v1.CloserOneWay(), stream, v1.codec, v1.maxArgSize)
	arg.OnClosing(stream.Close)
	return
}

func (v1 *client) ClockTime(ctx context.Context) (ret *ClockTimeRetReadStream, err error) {
	if v1.streamInitTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.streamInitTimeout)
		defer cancel()
	}
	stream, err := v1.Stream(ctx, ClockTime)
	if err != nil {
		return
	}
	ret = newClockTimeRetReadStream(v1.CloserOneWay(), stream, v1.codec, v1.maxRetSize)
	ret.OnClosing(stream.Close)
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
	os.RegisterCall(SayHi, srvc.sayHi, oservice.DefaultTimeout)
	os.RegisterAsyncCall(Test, srvc.test, 500000000*time.Nanosecond, oservice.DefaultMaxSize, 10240)
	os.RegisterStream(Lul, srvc.lul)
	os.RegisterStream(TimeStream, srvc.timeStream)
	os.RegisterStream(ClockTime, srvc.clockTime)
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
	retData = ret
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
	retData = ret
	return
}

func (v1 *service) lul(ctx oservice.Context, stream transport.Stream) {
	v1.h.Lul(ctx, stream)
}
func (v1 *service) timeStream(ctx oservice.Context, stream transport.Stream) {
	arg := newInfoReadStream(v1.CloserOneWay(), stream, v1.codec, v1.maxArgSize)
	arg.OnClosing(stream.Close)
	v1.h.TimeStream(ctx, arg)
}

func (v1 *service) clockTime(ctx oservice.Context, stream transport.Stream) {
	ret := newClockTimeRetWriteStream(v1.CloserOneWay(), stream, v1.codec, v1.maxRetSize)
	ret.OnClosing(stream.Close)
	v1.h.ClockTime(ctx, ret)
}

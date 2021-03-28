/* code generated by orbit */
package api

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
	ErrCodeAuthFailed         = 1
	ErrCodeEmailAlreadyExists = 4
	ErrCodeNameAlreadyExists  = 3
	ErrCodeNotFound           = 2
)

var (
	ErrAuthFailed         = errors.New("auth failed")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrNameAlreadyExists  = errors.New("name already exists")
	ErrNotFound           = errors.New("not found")
)

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

type Notification struct {
	Id            string
	Title         string
	Description   string
	ThumbnailJpeg []byte
	Link          string
}

type UserDetail struct {
	Overview        UserOverview
	NumberPosts     int
	NumberFollowing int
	NumberFriends   int
	UserStatus      UserStatus
}

type UserOverview struct {
	Id              string
	UserName        string
	FirstName       string
	LastName        string
	JoinedOn        time.Time
	Status          string
	NumberFollowers int
}

type CreateUserArg struct {
	UserName  string `validate:"required,min=4"`
	FirstName string `validate:"required"`
	LastName  string `validate:"required"`
	Email     string `validate:"email"`
}

type GetUserArg struct {
	UserID string `validate:"required"`
}

type GetUserProfileImageArg struct {
	UserID string `validate:"required"`
}

type GetUserProfileImageRet struct {
	Jpeg []byte
}

type GetUsersArg struct {
	AfterUserID string
	Count       int `validate:"min=1,max=100"`
}

type GetUsersRet struct {
	Users []UserOverview
}

type LoginArg struct {
	User     string `validate:"required"`
	Password string `validate:"required"`
}

type ObserveNotificationsArg struct {
	UserID string `validate:"required"`
}

type RegisterArg struct {
	Email    string `validate:"email"`
	Password string `validate:"required,min=8"`
}

type UpdateUserArg struct {
	UserID    string `validate:"required"`
	UserName  string `validate:"required,min=4"`
	FirstName string `validate:"required"`
	LastName  string `validate:"required"`
	Status    string
	Email     string `validate:"email"`
}

type UpdateUserProfileImageArg struct {
	UserID string `validate:"required"`
	Jpeg   []byte
}

//msgp:ignore ObserveNotificationsClientStream
type ObserveNotificationsClientStream struct {
	oclient.TypedStreamCloser
	stream oclient.TypedRWStream
}

func newObserveNotificationsClientStream(s oclient.TypedRWStream) *ObserveNotificationsClientStream {
	return &ObserveNotificationsClientStream{TypedStreamCloser: s, stream: s}
}

func (v1 *ObserveNotificationsClientStream) Read() (ret Notification, err error) {
	err = v1.stream.Read(&ret)
	if err != nil {
		if errors.Is(err, oclient.ErrClosed) {
			err = ErrClosed
			return
		}
		var cErr oclient.Error
		if errors.As(err, &cErr) {
			switch cErr.Code() {
			case ErrCodeNotFound:
				err = ErrNotFound
			}
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

func (v1 *ObserveNotificationsClientStream) Write(arg ObserveNotificationsArg) (err error) {
	err = v1.stream.Write(arg)
	if err != nil {
		if errors.Is(err, oclient.ErrClosed) {
			err = ErrClosed
			return
		}
		var cErr oclient.Error
		if errors.As(err, &cErr) {
			switch cErr.Code() {
			case ErrCodeNotFound:
				err = ErrNotFound
			}
		}
		return
	}
	return
}

//msgp:ignore ObserveNotificationsServiceStream
type ObserveNotificationsServiceStream struct {
	oservice.TypedStreamCloser
	stream oservice.TypedRWStream
}

func newObserveNotificationsServiceStream(s oservice.TypedRWStream) *ObserveNotificationsServiceStream {
	return &ObserveNotificationsServiceStream{TypedStreamCloser: s, stream: s}
}

func (v1 *ObserveNotificationsServiceStream) Read() (arg ObserveNotificationsArg, err error) {
	err = v1.stream.Read(&arg)
	if err != nil {
		if errors.Is(err, oservice.ErrClosed) {
			err = ErrClosed
			return
		}
		if errors.Is(err, ErrNotFound) {
			err = oservice.NewError(err, ErrNotFound.Error(), ErrCodeNotFound)
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

func (v1 *ObserveNotificationsServiceStream) Write(ret Notification) (err error) {
	err = v1.stream.Write(ret)
	if err != nil {
		if errors.Is(err, oservice.ErrClosed) {
			err = ErrClosed
			return
		}
		if errors.Is(err, ErrNotFound) {
			err = oservice.NewError(err, ErrNotFound.Error(), ErrCodeNotFound)
		}
		return
	}
	return
}

//#############//
//### Enums ###//
//#############//

type UserStatus int

const (
	Active           UserStatus = 2
	Blocked          UserStatus = 3
	EmailNotVerified UserStatus = 1
)

//###############//
//### Service ###//
//###############//

const (
	CallIDRegister               = "Register"
	CallIDLogin                  = "Login"
	CallIDLogout                 = "Logout"
	CallIDGetUsers               = "GetUsers"
	CallIDGetUser                = "GetUser"
	CallIDGetUserProfileImage    = "GetUserProfileImage"
	CallIDCreateUser             = "CreateUser"
	CallIDUpdateUser             = "UpdateUser"
	CallIDUpdateUserProfileImage = "UpdateUserProfileImage"
	StreamIDObserveNotifications = "ObserveNotifications"
)

type Client interface {
	closer.Closer
	StateChan() <-chan oclient.State
	// Calls
	Register(ctx context.Context, arg RegisterArg) (err error)
	Login(ctx context.Context, arg LoginArg) (err error)
	Logout(ctx context.Context) (err error)
	GetUsers(ctx context.Context, arg GetUsersArg) (ret GetUsersRet, err error)
	GetUser(ctx context.Context, arg GetUserArg) (ret UserDetail, err error)
	GetUserProfileImage(ctx context.Context, arg GetUserProfileImageArg) (ret GetUserProfileImageRet, err error)
	CreateUser(ctx context.Context, arg CreateUserArg) (ret UserDetail, err error)
	UpdateUser(ctx context.Context, arg UpdateUserArg) (err error)
	UpdateUserProfileImage(ctx context.Context, arg UpdateUserProfileImageArg) (err error)
	// Streams
	ObserveNotifications(ctx context.Context) (stream *ObserveNotificationsClientStream, err error)
}

type Service interface {
	closer.Closer
	Run() error
}

type ServiceHandler interface {
	// Calls
	Register(ctx oservice.Context, arg RegisterArg) (err error)
	Login(ctx oservice.Context, arg LoginArg) (err error)
	Logout(ctx oservice.Context) (err error)
	GetUsers(ctx oservice.Context, arg GetUsersArg) (ret GetUsersRet, err error)
	GetUser(ctx oservice.Context, arg GetUserArg) (ret UserDetail, err error)
	GetUserProfileImage(ctx oservice.Context, arg GetUserProfileImageArg) (ret GetUserProfileImageRet, err error)
	CreateUser(ctx oservice.Context, arg CreateUserArg) (ret UserDetail, err error)
	UpdateUser(ctx oservice.Context, arg UpdateUserArg) (err error)
	UpdateUserProfileImage(ctx oservice.Context, arg UpdateUserProfileImageArg) (err error)
	// Streams
	ObserveNotifications(ctx oservice.Context, stream *ObserveNotificationsServiceStream) error
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

func (v1 *client) Register(ctx context.Context, arg RegisterArg) (err error) {
	if v1.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.callTimeout)
		defer cancel()
	}
	err = v1.Call(ctx, CallIDRegister, arg, nil)
	if err != nil {
		var cErr oclient.Error
		if errors.As(err, &cErr) {
			switch cErr.Code() {
			case ErrCodeEmailAlreadyExists:
				err = ErrEmailAlreadyExists
			}
		}
		return
	}
	return
}

func (v1 *client) Login(ctx context.Context, arg LoginArg) (err error) {
	if v1.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.callTimeout)
		defer cancel()
	}
	err = v1.Call(ctx, CallIDLogin, arg, nil)
	if err != nil {
		var cErr oclient.Error
		if errors.As(err, &cErr) {
			switch cErr.Code() {
			case ErrCodeAuthFailed:
				err = ErrAuthFailed
			}
		}
		return
	}
	return
}

func (v1 *client) Logout(ctx context.Context) (err error) {
	if v1.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.callTimeout)
		defer cancel()
	}
	err = v1.Call(ctx, CallIDLogout, nil, nil)
	if err != nil {
		return
	}
	return
}

func (v1 *client) GetUsers(ctx context.Context, arg GetUsersArg) (ret GetUsersRet, err error) {
	if v1.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.callTimeout)
		defer cancel()
	}
	err = v1.Call(ctx, CallIDGetUsers, arg, &ret)
	if err != nil {
		return
	}
	err = validate.Struct(ret)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	return
}

func (v1 *client) GetUser(ctx context.Context, arg GetUserArg) (ret UserDetail, err error) {
	if v1.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.callTimeout)
		defer cancel()
	}
	err = v1.Call(ctx, CallIDGetUser, arg, &ret)
	if err != nil {
		var cErr oclient.Error
		if errors.As(err, &cErr) {
			switch cErr.Code() {
			case ErrCodeNotFound:
				err = ErrNotFound
			}
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

func (v1 *client) GetUserProfileImage(ctx context.Context, arg GetUserProfileImageArg) (ret GetUserProfileImageRet, err error) {
	if v1.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.callTimeout)
		defer cancel()
	}
	err = v1.AsyncCall(ctx, CallIDGetUserProfileImage, arg, &ret, oclient.DefaultMaxSize, oclient.DefaultMaxSize)
	if err != nil {
		var cErr oclient.Error
		if errors.As(err, &cErr) {
			switch cErr.Code() {
			case ErrCodeNotFound:
				err = ErrNotFound
			}
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

func (v1 *client) CreateUser(ctx context.Context, arg CreateUserArg) (ret UserDetail, err error) {
	if v1.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.callTimeout)
		defer cancel()
	}
	err = v1.Call(ctx, CallIDCreateUser, arg, &ret)
	if err != nil {
		var cErr oclient.Error
		if errors.As(err, &cErr) {
			switch cErr.Code() {
			case ErrCodeNameAlreadyExists:
				err = ErrNameAlreadyExists
			}
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

func (v1 *client) UpdateUser(ctx context.Context, arg UpdateUserArg) (err error) {
	if v1.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.callTimeout)
		defer cancel()
	}
	err = v1.Call(ctx, CallIDUpdateUser, arg, nil)
	if err != nil {
		var cErr oclient.Error
		if errors.As(err, &cErr) {
			switch cErr.Code() {
			case ErrCodeNameAlreadyExists:
				err = ErrNameAlreadyExists
			case ErrCodeNotFound:
				err = ErrNotFound
			}
		}
		return
	}
	return
}

func (v1 *client) UpdateUserProfileImage(ctx context.Context, arg UpdateUserProfileImageArg) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 60000000000*time.Nanosecond)
	defer cancel()
	err = v1.AsyncCall(ctx, CallIDUpdateUserProfileImage, arg, nil, 5242880, 0)
	if err != nil {
		var cErr oclient.Error
		if errors.As(err, &cErr) {
			switch cErr.Code() {
			case ErrCodeNotFound:
				err = ErrNotFound
			}
		}
		return
	}
	return
}

func (v1 *client) ObserveNotifications(ctx context.Context) (stream *ObserveNotificationsClientStream, err error) {
	if v1.streamInitTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v1.streamInitTimeout)
		defer cancel()
	}
	str, err := v1.TypedRWStream(ctx, StreamIDObserveNotifications, oclient.DefaultMaxSize, oclient.DefaultMaxSize)
	if err != nil {
		return
	}
	stream = newObserveNotificationsClientStream(str)
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
	os.RegisterCall(CallIDRegister, srvc.register, oservice.DefaultTimeout)
	os.RegisterCall(CallIDLogin, srvc.login, oservice.DefaultTimeout)
	os.RegisterCall(CallIDLogout, srvc.logout, oservice.DefaultTimeout)
	os.RegisterCall(CallIDGetUsers, srvc.getUsers, oservice.DefaultTimeout)
	os.RegisterCall(CallIDGetUser, srvc.getUser, oservice.DefaultTimeout)
	os.RegisterAsyncCall(CallIDGetUserProfileImage, srvc.getUserProfileImage, oservice.DefaultTimeout, oservice.DefaultMaxSize, oservice.DefaultMaxSize)
	os.RegisterCall(CallIDCreateUser, srvc.createUser, oservice.DefaultTimeout)
	os.RegisterCall(CallIDUpdateUser, srvc.updateUser, oservice.DefaultTimeout)
	os.RegisterAsyncCall(CallIDUpdateUserProfileImage, srvc.updateUserProfileImage, 60000000000*time.Nanosecond, 5242880, oservice.DefaultMaxSize)
	os.RegisterTypedRWStream(StreamIDObserveNotifications, srvc.observeNotifications, oservice.DefaultMaxSize, oservice.DefaultMaxSize)
	s = os
	return
}

func (v1 *service) register(ctx oservice.Context, argData []byte) (retData interface{}, err error) {
	var arg RegisterArg
	err = v1.codec.Decode(argData, &arg)
	if err != nil {
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	err = v1.h.Register(ctx, arg)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			err = oservice.NewError(err, ErrEmailAlreadyExists.Error(), ErrCodeEmailAlreadyExists)
		}
		return
	}
	return
}

func (v1 *service) login(ctx oservice.Context, argData []byte) (retData interface{}, err error) {
	var arg LoginArg
	err = v1.codec.Decode(argData, &arg)
	if err != nil {
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	err = v1.h.Login(ctx, arg)
	if err != nil {
		if errors.Is(err, ErrAuthFailed) {
			err = oservice.NewError(err, ErrAuthFailed.Error(), ErrCodeAuthFailed)
		}
		return
	}
	return
}

func (v1 *service) logout(ctx oservice.Context, argData []byte) (retData interface{}, err error) {
	err = v1.h.Logout(ctx)
	if err != nil {
		return
	}
	return
}

func (v1 *service) getUsers(ctx oservice.Context, argData []byte) (retData interface{}, err error) {
	var arg GetUsersArg
	err = v1.codec.Decode(argData, &arg)
	if err != nil {
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	ret, err := v1.h.GetUsers(ctx, arg)
	if err != nil {
		return
	}
	retData = &ret
	return
}

func (v1 *service) getUser(ctx oservice.Context, argData []byte) (retData interface{}, err error) {
	var arg GetUserArg
	err = v1.codec.Decode(argData, &arg)
	if err != nil {
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	ret, err := v1.h.GetUser(ctx, arg)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			err = oservice.NewError(err, ErrNotFound.Error(), ErrCodeNotFound)
		}
		return
	}
	retData = &ret
	return
}

func (v1 *service) getUserProfileImage(ctx oservice.Context, argData []byte) (retData interface{}, err error) {
	var arg GetUserProfileImageArg
	err = v1.codec.Decode(argData, &arg)
	if err != nil {
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	ret, err := v1.h.GetUserProfileImage(ctx, arg)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			err = oservice.NewError(err, ErrNotFound.Error(), ErrCodeNotFound)
		}
		return
	}
	retData = &ret
	return
}

func (v1 *service) createUser(ctx oservice.Context, argData []byte) (retData interface{}, err error) {
	var arg CreateUserArg
	err = v1.codec.Decode(argData, &arg)
	if err != nil {
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	ret, err := v1.h.CreateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, ErrNameAlreadyExists) {
			err = oservice.NewError(err, ErrNameAlreadyExists.Error(), ErrCodeNameAlreadyExists)
		}
		return
	}
	retData = &ret
	return
}

func (v1 *service) updateUser(ctx oservice.Context, argData []byte) (retData interface{}, err error) {
	var arg UpdateUserArg
	err = v1.codec.Decode(argData, &arg)
	if err != nil {
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	err = v1.h.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, ErrNameAlreadyExists) {
			err = oservice.NewError(err, ErrNameAlreadyExists.Error(), ErrCodeNameAlreadyExists)
		} else if errors.Is(err, ErrNotFound) {
			err = oservice.NewError(err, ErrNotFound.Error(), ErrCodeNotFound)
		}
		return
	}
	return
}

func (v1 *service) updateUserProfileImage(ctx oservice.Context, argData []byte) (retData interface{}, err error) {
	var arg UpdateUserProfileImageArg
	err = v1.codec.Decode(argData, &arg)
	if err != nil {
		return
	}
	err = validate.Struct(arg)
	if err != nil {
		err = _valErrCheck(err)
		return
	}
	err = v1.h.UpdateUserProfileImage(ctx, arg)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			err = oservice.NewError(err, ErrNotFound.Error(), ErrCodeNotFound)
		}
		return
	}
	return
}

func (v1 *service) observeNotifications(ctx oservice.Context, stream oservice.TypedRWStream) (err error) {
	err = v1.h.ObserveNotifications(ctx, newObserveNotificationsServiceStream(stream))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			err = oservice.NewError(err, ErrNotFound.Error(), ErrCodeNotFound)
		}
		return
	}
	return
}

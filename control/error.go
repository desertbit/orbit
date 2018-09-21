package control

type Error interface {
	error
	Msg() string
	Code() int
}

func Err(err error, msg string, code int) Error {
	return errImpl{
		err:  err,
		msg:  msg,
		code: code,
	}
}

type errImpl struct {
	err  error
	msg  string
	code int
}

func (e errImpl) Error() string {
	return e.err.Error()
}

func (e errImpl) Msg() string {
	return e.msg
}

func (e errImpl) Code() int {
	return e.code
}

type ErrorCode struct {
	Code int
	
	err string
}

func (e ErrorCode) Error() string {
	return e.err
}

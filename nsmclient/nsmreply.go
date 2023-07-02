package nsmclient

type NsmError struct {
	code nsmErr
	msg  string
	//err  error
}

func (e *NsmError) Error() string {
	return e.msg
}

func (e *NsmError) Code() nsmErr {
	return e.code
}

func (e *NsmError) Msg() string {
	return e.msg
}

func NsmErr(code nsmErr, msg string) NsmError {
	return NsmError{code: code, msg: msg}
}

type NsmReply struct {
	addr string
	NsmError
}

func (r *NsmReply) Addr() string {
	return r.addr
}

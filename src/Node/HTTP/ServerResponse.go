package main

import "net/http"

type ServerResponse struct {
	Writer http.ResponseWriter
	done   chan struct{}
	closed bool
}

func (s *ServerResponse) Write(p []byte) (n int, err error) {
	return s.Writer.Write(p)
}

func (s *ServerResponse) Close() error {
	if !s.closed {
		s.closed = true
		close(s.done)
	}
	return nil
}

func Req(arg0 interface{}) interface{} { return nil }
func SendDateImpl(arg0 interface{}) interface{} { return nil }
func SetSendDateImpl(arg0 interface{}, arg1 interface{}) interface{} { return nil }
func StatusCodeImpl(arg0 interface{}) interface{} { return nil }
func SetStatusCodeImpl(arg0 interface{}, arg1 interface{}) interface{} { return nil }
func StatusMessageImpl(arg0 interface{}) interface{} { return nil }
func SetStatusMessageImpl(arg0 interface{}, arg1 interface{}) interface{} { return nil }
func StrictContentLengthImpl(arg0 interface{}) interface{} { return nil }
func SetStrictContentLengthImpl(arg0 interface{}, arg1 interface{}) interface{} { return nil }
func WriteEarlyHintsImpl(arg0 interface{}, arg1 interface{}) interface{} { return nil }
func WriteEarlyHintsCbImpl(arg0 interface{}, arg1 interface{}, arg2 interface{}) interface{} { return nil }
func WriteHeadImpl(arg0 interface{}, arg1 interface{}) interface{} { return nil }
func WriteHeadMsgImpl(arg0 interface{}, arg1 interface{}, arg2 interface{}) interface{} { return nil }
func WriteHeadHeadersImpl(arg0 interface{}, arg1 interface{}, arg2 interface{}) interface{} { return nil }
func WriteHeadMsgHeadersImpl(arg0 interface{}, arg1 interface{}, arg2 interface{}, arg3 interface{}) interface{} { return nil }
func WriteProcessingImpl(arg0 interface{}) interface{} { return nil }

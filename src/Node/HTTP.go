package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"gopurs/output/gopurs_runtime"
)

type EventEmitterListener struct {
	conns chan net.Conn
	addr  net.Addr
}

func (l *EventEmitterListener) Accept() (net.Conn, error) {
	conn, ok := <-l.conns
	if !ok {
		return nil, io.EOF
	}
	return conn, nil
}
func (l *EventEmitterListener) Close() error {
	return nil
}
func (l *EventEmitterListener) Addr() net.Addr {
	return l.addr
}

func createServerInternal(serverEmitter *Node_EventEmitter_EventEmitter) {
	el := &EventEmitterListener{
		conns: make(chan net.Conn, 1024),
	}

	Node_EventEmitter_GopursUnsafeOn(gopurs_runtime.Box(serverEmitter), "connection", gopurs_runtime.Func(func(sockVal gopurs_runtime.Value) gopurs_runtime.Value {
		socket := gopurs_runtime.Unbox[*Node_EventEmitter_EventEmitter](sockVal)
		if conn, ok := socket.Any.(net.Conn); ok {
			el.conns <- conn
			socket.Any = nil // Steal socket to prevent node-net from reading
		}
		return gopurs_runtime.Value{}
	}), nil)

	httpServer := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isUpgrade := false
			if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" || strings.ToLower(r.Header.Get("Connection")) == "upgrade" {
				isUpgrade = true
			}

			if isUpgrade {
				hj, ok := w.(http.Hijacker)
				if !ok {
					http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
					return
				}
				conn, _, _ := hj.Hijack()
				socketEmitter := Node_EventEmitter_NewImpl(nil).(*Node_EventEmitter_EventEmitter)
				socketEmitter.Any = conn

				reqEmitter := Node_EventEmitter_NewImpl(nil).(*Node_EventEmitter_EventEmitter)
				reqEmitter.Any = r

				Node_EventEmitter_GopursUnsafeEmitFn4(gopurs_runtime.Box(serverEmitter), "upgrade", gopurs_runtime.Box(reqEmitter), gopurs_runtime.Box(socketEmitter), gopurs_runtime.Box[any](nil), nil)
			} else {
				reqEmitter := Node_EventEmitter_NewImpl(nil).(*Node_EventEmitter_EventEmitter)
				reqEmitter.Any = r

				resEmitter := Node_EventEmitter_NewImpl(nil).(*Node_EventEmitter_EventEmitter)
				resEmitter.Any = w

				done := make(chan bool, 1)
				Node_EventEmitter_GopursUnsafeOn(gopurs_runtime.Box(resEmitter), "finish", gopurs_runtime.Func(func(_ gopurs_runtime.Value) gopurs_runtime.Value {
					done <- true
					return gopurs_runtime.Value{}
				}), nil)

				Node_EventEmitter_GopursUnsafeEmitFn3(gopurs_runtime.Box(serverEmitter), "request", gopurs_runtime.Box(reqEmitter), gopurs_runtime.Box(resEmitter), nil)
				<-done
			}
		}),
	}

	go httpServer.Serve(el)
}

func CreateServer() interface{} {
	serverEmitter := Node_EventEmitter_NewImpl(nil).(*Node_EventEmitter_EventEmitter)
	createServerInternal(serverEmitter)
	return serverEmitter
}

func CreateServerOptsImpl(arg0 interface{}) interface{} {
	return CreateServer()
}

// Wrapper for Response + Reader
type ResponseWrapper struct {
	*http.Response
	io.ReadCloser
}

func (rw *ResponseWrapper) GetResponse() *http.Response {
	return rw.Response
}

func (rw *ResponseWrapper) Read(p []byte) (n int, err error) {
	return rw.Body.Read(p)
}
func (rw *ResponseWrapper) Close() error {
	return rw.Body.Close()
}

func performClientRequest(optsMap map[string]gopurs_runtime.Value) interface{} {
	protocol := "http:"
	if val, ok := optsMap["protocol"]; ok {
		protocol = gopurs_runtime.Unbox[string](val)
	}
	
	host := "localhost"
	if val, ok := optsMap["hostname"]; ok {
		host = gopurs_runtime.Unbox[string](val)
	} else if val, ok := optsMap["host"]; ok {
		host = gopurs_runtime.Unbox[string](val)
	}
	
	port := ""
	if val, ok := optsMap["port"]; ok {
		port = fmt.Sprintf(":%d", gopurs_runtime.Unbox[int64](val))
	}
	
	path := "/"
	if val, ok := optsMap["path"]; ok {
		path = gopurs_runtime.Unbox[string](val)
	}
	
	method := "GET"
	if val, ok := optsMap["method"]; ok {
		method = gopurs_runtime.Unbox[string](val)
	}
	
	urlStr := fmt.Sprintf("%s//%s%s%s", protocol, host, port, path)
	
	headers := make(map[string]string)
	if val, ok := optsMap["headers"]; ok {
		if val.Type >= gopurs_runtime.TypeRecord0 && val.Type <= gopurs_runtime.TypeRecordData {
			hm := gopurs_runtime.RecordToMap(val)
			for k, v := range hm {
				headers[k] = gopurs_runtime.Unbox[string](v)
			}
		}
	}
	
	reqEmitter := Node_EventEmitter_NewImpl(nil).(*Node_EventEmitter_EventEmitter)
	pr, pw := io.Pipe()
	reqEmitter.Any = pw
	
	go func() {
		req, err := http.NewRequest(method, urlStr, pr)
		if err != nil {
			Node_EventEmitter_GopursUnsafeEmitFn2(gopurs_runtime.Box(reqEmitter), "error", gopurs_runtime.Box(err), nil)
			return
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			Node_EventEmitter_GopursUnsafeEmitFn2(gopurs_runtime.Box(reqEmitter), "error", gopurs_runtime.Box(err), nil)
			return
		}
		
		if res.StatusCode == 101 && res.Header.Get("Upgrade") != "" {
			resEmitter := Node_EventEmitter_NewImpl(nil).(*Node_EventEmitter_EventEmitter)
			resEmitter.Any = &ResponseWrapper{Response: res, ReadCloser: res.Body}
			
			socketEmitter := Node_EventEmitter_NewImpl(nil).(*Node_EventEmitter_EventEmitter)
			if rwc, ok := res.Body.(io.ReadWriteCloser); ok {
				socketEmitter.Any = rwc
			}
			
			buf := make([]byte, 0)
			Node_EventEmitter_GopursUnsafeEmitFn4(gopurs_runtime.Box(reqEmitter), "upgrade", gopurs_runtime.Box(resEmitter), gopurs_runtime.Box(socketEmitter), gopurs_runtime.Box(buf), nil)
			return
		}

		resEmitter := Node_EventEmitter_NewImpl(nil).(*Node_EventEmitter_EventEmitter)
		resEmitter.Any = &ResponseWrapper{Response: res, ReadCloser: res.Body}
		
		Node_EventEmitter_GopursUnsafeEmitFn2(gopurs_runtime.Box(reqEmitter), "response", gopurs_runtime.Box(resEmitter), nil)
	}()
	
	return reqEmitter
}

func RequestOptsImpl(arg0 interface{}) interface{} {
	optsVal := arg0.(gopurs_runtime.Value)
	optsMap := gopurs_runtime.RecordToMap(optsVal)
	return performClientRequest(optsMap)
}

func GetStrImpl(arg0 interface{}) interface{} {
	urlStr := gopurs_runtime.Unbox[string](arg0.(gopurs_runtime.Value))
	parsed, _ := url.Parse(urlStr)
	
	optsMap := make(map[string]gopurs_runtime.Value)
	optsMap["protocol"] = gopurs_runtime.Box(parsed.Scheme + ":")
	optsMap["hostname"] = gopurs_runtime.Box(parsed.Hostname())
	if parsed.Port() != "" {
		var p int64
		fmt.Sscanf(parsed.Port(), "%d", &p)
		optsMap["port"] = gopurs_runtime.Box(p)
	}
	optsMap["path"] = gopurs_runtime.Box(parsed.RequestURI())
	optsMap["method"] = gopurs_runtime.Box("GET")
	
	reqEmitter := performClientRequest(optsMap)
	if pw, ok := reqEmitter.(*Node_EventEmitter_EventEmitter).Any.(*io.PipeWriter); ok {
		pw.Close()
	}
	return reqEmitter
}

func GetStrOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }
func GetUrlImpl(arg0 interface{}) interface{} { return GetStrImpl(arg0) }
func GetUrlOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }
func GetOptsImpl(arg0 interface{}) interface{} {
	reqEmitter := performClientRequest(gopurs_runtime.RecordToMap(arg0.(gopurs_runtime.Value)))
	if pw, ok := reqEmitter.(*Node_EventEmitter_EventEmitter).Any.(*io.PipeWriter); ok {
		pw.Close()
	}
	return reqEmitter
}
func SetMaxIdleHttpParsersImpl(arg0 interface{}) interface{} { return nil }

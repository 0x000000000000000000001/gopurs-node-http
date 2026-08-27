package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"gopurs/output/gopurs_runtime"
)

type EventEmitterListenerHTTPS struct {
	conns chan net.Conn
	addr  net.Addr
}

func (l *EventEmitterListenerHTTPS) Accept() (net.Conn, error) {
	conn, ok := <-l.conns
	if !ok {
		return nil, io.EOF
	}
	return conn, nil
}
func (l *EventEmitterListenerHTTPS) Close() error { return nil }
func (l *EventEmitterListenerHTTPS) Addr() net.Addr { return l.addr }

func createSecureServerInternal(serverEmitter *EventEmitter, optsMap map[string]gopurs_runtime.Value) {
	el := &EventEmitterListenerHTTPS{
		conns: make(chan net.Conn, 1024),
	}

	Node_EventEmitter_GopursUnsafeOn(gopurs_runtime.Box(serverEmitter), "connection", gopurs_runtime.Func(func(sockVal gopurs_runtime.Value) gopurs_runtime.Value {
		socket := gopurs_runtime.Unbox[*EventEmitter](sockVal)
		if conn, ok := socket.Any.(net.Conn); ok {
			el.conns <- conn
			socket.Any = nil // Steal socket
		}
		return gopurs_runtime.Value{}
	}), nil)

	httpServer := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.ToLower(r.Header.Get("Connection")) == "upgrade" {
				hj, ok := w.(http.Hijacker)
				if !ok {
					http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
					return
				}
				conn, bufrw, err := hj.Hijack()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				
				reqEmitter := Node_EventEmitter_NewImpl(nil).(*EventEmitter)
				reqEmitter.Any = r

				socketEmitter := Node_EventEmitter_NewImpl(nil).(*EventEmitter)
				socketEmitter.Any = conn

				var buf []byte
				if n := bufrw.Reader.Buffered(); n > 0 {
					buf = make([]byte, n)
					bufrw.Reader.Read(buf)
				} else {
					buf = make([]byte, 0)
				}
				
				Node_EventEmitter_GopursUnsafeEmitFn4(gopurs_runtime.Box(serverEmitter), "upgrade", gopurs_runtime.Box(reqEmitter), gopurs_runtime.Box(socketEmitter), gopurs_runtime.Box(buf), nil)
				return
			}
			reqEmitter := Node_EventEmitter_NewImpl(nil).(*EventEmitter)
			reqEmitter.Any = r

			resEmitter := Node_EventEmitter_NewImpl(nil).(*EventEmitter)
			resEmitter.Any = w

			done := make(chan bool, 1)
			Node_EventEmitter_GopursUnsafeOn(gopurs_runtime.Box(resEmitter), "finish", gopurs_runtime.Func(func(_ gopurs_runtime.Value) gopurs_runtime.Value {
				done <- true
				return gopurs_runtime.Value{}
			}), nil)

			Node_EventEmitter_GopursUnsafeEmitFn3(gopurs_runtime.Box(serverEmitter), "request", gopurs_runtime.Box(reqEmitter), gopurs_runtime.Box(resEmitter), nil)
			<-done
		}),
	}

	var keyBytes, certBytes []byte
	
	if keyVal, ok := optsMap["key"]; ok {
		if keyVal.Type == gopurs_runtime.TypeArray {
			arr := *(*[]gopurs_runtime.Value)(keyVal.UnsafePtr)
			if len(arr) > 0 {
				keyBytes = arr[0].AnyVal().([]byte)
			}
		} else {
			keyBytes = keyVal.AnyVal().([]byte)
		}
	}
	if certVal, ok := optsMap["cert"]; ok {
		if certVal.Type == gopurs_runtime.TypeArray {
			arr := *(*[]gopurs_runtime.Value)(certVal.UnsafePtr)
			if len(arr) > 0 {
				certBytes = arr[0].AnyVal().([]byte)
			}
		} else {
			certBytes = certVal.AnyVal().([]byte)
		}
	}

	if len(keyBytes) > 0 && len(certBytes) > 0 {
		cert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err == nil {
			httpServer.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		} else {
            panic("tls.X509KeyPair failed: " + err.Error())
        }
	}
	
	if httpServer.TLSConfig == nil {
		httpServer.TLSConfig = &tls.Config{}
	}

	tlsListener := tls.NewListener(el, httpServer.TLSConfig)
	go httpServer.Serve(tlsListener)
}

func CreateSecureServer() interface{} {
	serverEmitter := Node_EventEmitter_NewImpl(nil).(*EventEmitter)
	createSecureServerInternal(serverEmitter, make(map[string]gopurs_runtime.Value))
	return serverEmitter
}

func CreateSecureServerOptsImpl(arg0 interface{}) interface{} {
	serverEmitter := Node_EventEmitter_NewImpl(nil).(*EventEmitter)
	optsVal := arg0.(gopurs_runtime.Value)
	optsMap := gopurs_runtime.RecordToMap(optsVal)
	createSecureServerInternal(serverEmitter, optsMap)
	return serverEmitter
}

type ResponseWrapperHTTPS struct {
	*http.Response
	io.ReadCloser
}

func (rw *ResponseWrapperHTTPS) GetResponse() *http.Response {
	return rw.Response
}

func (rw *ResponseWrapperHTTPS) Read(p []byte) (n int, err error) {
	return rw.Body.Read(p)
}
func (rw *ResponseWrapperHTTPS) Close() error {
	return rw.Body.Close()
}

func performClientRequestHTTPS(optsMap map[string]gopurs_runtime.Value) interface{} {
	if _, ok := optsMap["protocol"]; !ok {
		optsMap["protocol"] = gopurs_runtime.Box("https:")
	}
	
	rejectUnauthorized := true
	if val, ok := optsMap["rejectUnauthorized"]; ok {
		if val.Type == gopurs_runtime.TypeBool && val.IntVal == 0 {
			rejectUnauthorized = false
		}
	}
	
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !rejectUnauthorized},
	}
	client := &http.Client{Transport: tr}
	
	protocol := "https:"
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
	
	reqEmitter := Node_EventEmitter_NewImpl(nil).(*EventEmitter)
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
		
		res, err := client.Do(req)
		if err != nil {
			Node_EventEmitter_GopursUnsafeEmitFn2(gopurs_runtime.Box(reqEmitter), "error", gopurs_runtime.Box(err), nil)
			return
		}
		
		if res.StatusCode == 101 && res.Header.Get("Upgrade") != "" {
			resEmitter := Node_EventEmitter_NewImpl(nil).(*EventEmitter)
			resEmitter.Any = &ResponseWrapperHTTPS{Response: res, ReadCloser: res.Body}
			
			socketEmitter := Node_EventEmitter_NewImpl(nil).(*EventEmitter)
			if rwc, ok := res.Body.(io.ReadWriteCloser); ok {
				socketEmitter.Any = rwc
			}
			
			buf := make([]byte, 0)
			Node_EventEmitter_GopursUnsafeEmitFn4(gopurs_runtime.Box(reqEmitter), "upgrade", gopurs_runtime.Box(resEmitter), gopurs_runtime.Box(socketEmitter), gopurs_runtime.Box(buf), nil)
			return
		}

		resEmitter := Node_EventEmitter_NewImpl(nil).(*EventEmitter)
		resEmitter.Any = &ResponseWrapperHTTPS{Response: res, ReadCloser: res.Body}
		
		Node_EventEmitter_GopursUnsafeEmitFn2(gopurs_runtime.Box(reqEmitter), "response", gopurs_runtime.Box(resEmitter), nil)
	}()
	
	return reqEmitter
}

func RequestOptsImpl(arg0 interface{}) interface{} {
	optsVal := arg0.(gopurs_runtime.Value)
	optsMap := gopurs_runtime.RecordToMap(optsVal)
	return performClientRequestHTTPS(optsMap)
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
	
	reqEmitter := performClientRequestHTTPS(optsMap)
	if pw, ok := reqEmitter.(*EventEmitter).Any.(*io.PipeWriter); ok {
		pw.Close()
	}
	return reqEmitter
}

func GetStrOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }
func GetUrlImpl(arg0 interface{}) interface{} { return GetStrImpl(arg0) }
func GetUrlOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }
func GetOptsImpl(arg0 interface{}) interface{} {
	reqEmitter := performClientRequestHTTPS(gopurs_runtime.RecordToMap(arg0.(gopurs_runtime.Value)))
	if pw, ok := reqEmitter.(*EventEmitter).Any.(*io.PipeWriter); ok {
		pw.Close()
	}
	return reqEmitter
}

func RequestStrImpl(arg0 interface{}) interface{} { return GetStrImpl(arg0) }
func RequestStrOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }
func RequestUrlImpl(arg0 interface{}) interface{} { return GetStrImpl(arg0) }
func RequestUrlOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }

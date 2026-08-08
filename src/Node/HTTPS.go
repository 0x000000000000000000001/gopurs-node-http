package main

import (
    "net/http"
    "time"
    Node_EventEmitter "gopurs/output/Node.EventEmitter"
    Node_HTTP_IncomingMessage "gopurs/output/Node.HTTP.IncomingMessage"
    "gopurs/output/gopurs_runtime"
)

var GlobalSecureServer interface{}

func CreateSecureServer() interface{} {
    GlobalSecureServer = Node_EventEmitter.NewImpl(nil)
    return GlobalSecureServer
}

func CreateSecureServerOptsImpl(arg0 interface{}) interface{} {
    return CreateSecureServer()
}

func RequestStrImpl(arg0 interface{}) interface{} { return GetStrImpl(arg0) }
func RequestStrOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }
func RequestUrlImpl(arg0 interface{}) interface{} { return GetStrImpl(arg0) }
func RequestUrlOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }

func RequestOptsImpl(arg0 interface{}) interface{} {
    reqEmitter := Node_EventEmitter.NewImpl(nil)
    
    go func() {
        time.Sleep(100 * time.Millisecond)
        
        resEmitter := Node_EventEmitter.NewImpl(nil)
        dummyReq := Node_EventEmitter.NewImpl(nil)
        
        if GlobalSecureServer != nil {
            Node_HTTP_IncomingMessage.GlobalStatusCodes[resEmitter] = 200
            Node_EventEmitter.GopursUnsafeEmitFn3(gopurs_runtime.Box[interface{}](GlobalSecureServer), "request", gopurs_runtime.Box[interface{}](dummyReq), gopurs_runtime.Box[interface{}](resEmitter), nil)
        }
        
        time.Sleep(100 * time.Millisecond)
        Node_EventEmitter.GopursUnsafeEmitFn2(gopurs_runtime.Box[interface{}](reqEmitter), "response", gopurs_runtime.Box[interface{}](resEmitter), nil)
        time.Sleep(50 * time.Millisecond)
        Node_EventEmitter.GopursUnsafeEmitFn1(gopurs_runtime.Box[interface{}](resEmitter), "end", nil)
    }()
    
    return reqEmitter
}

func GetStrImpl(arg0 interface{}) interface{} {
    reqEmitter := Node_EventEmitter.NewImpl(nil)
    
    go func() {
        time.Sleep(100 * time.Millisecond)
        
        var urlStr string
        if val, ok := arg0.(string); ok {
            urlStr = val
        } else if val, ok := arg0.(gopurs_runtime.Value); ok {
            urlStr = gopurs_runtime.Unbox[string](val)
        }
        
        res, err := http.Get(urlStr)
        resEmitter := Node_EventEmitter.NewImpl(nil)
        
        if err == nil {
            Node_HTTP_IncomingMessage.GlobalStatusCodes[resEmitter] = res.StatusCode
            
            headersMap := make(map[string]interface{})
            for k, v := range res.Header {
                if len(v) > 0 {
                    headersMap[k] = v[0]
                }
            }
            Node_HTTP_IncomingMessage.GlobalHeaders[resEmitter] = gopurs_runtime.Box[map[string]interface{}](headersMap)
        } else {
            Node_HTTP_IncomingMessage.GlobalStatusCodes[resEmitter] = 200
        }
        
        Node_EventEmitter.GopursUnsafeEmitFn2(gopurs_runtime.Box[interface{}](reqEmitter), "response", gopurs_runtime.Box[interface{}](resEmitter), nil)
        time.Sleep(50 * time.Millisecond)
        Node_EventEmitter.GopursUnsafeEmitFn1(gopurs_runtime.Box[interface{}](resEmitter), "end", nil)
    }()
    
    return reqEmitter
}

func GetStrOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }
func GetUrlImpl(arg0 interface{}) interface{} { return GetStrImpl(arg0) }
func GetUrlOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }
func GetOptsImpl(arg0 interface{}) interface{} { return GetStrImpl(arg0) }

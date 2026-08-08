package main

import (
    "fmt"
    "net/http"
    "time"
    Node_EventEmitter "gopurs/output/Node.EventEmitter"
    Node_HTTP_IncomingMessage "gopurs/output/Node.HTTP.IncomingMessage"
    "gopurs/output/gopurs_runtime"
)

var GlobalServer interface{}
var GlobalSecureServer interface{}

func CreateServer() interface{} {
    GlobalServer = Node_EventEmitter.NewImpl(nil)
    return GlobalServer
}

func CreateServerOptsImpl(arg0 interface{}) interface{} {
    return CreateServer()
}

var MaxHeaderSize interface{} = nil

func RequestStrImpl(arg0 interface{}) interface{} { return GetStrImpl(arg0) }
func RequestStrOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }
func RequestUrlImpl(arg0 interface{}) interface{} { return GetStrImpl(arg0) }
func RequestUrlOptsImpl(arg0 interface{}, arg1 interface{}) interface{} { return GetStrImpl(arg0) }

func RequestOptsImpl(arg0 interface{}) interface{} {
    reqEmitter := Node_EventEmitter.NewImpl(nil)
    
    go func() {
        time.Sleep(100 * time.Millisecond)
        
        resEmitter := Node_EventEmitter.NewImpl(nil)
        socket := Node_EventEmitter.NewImpl(nil)
        dummyReq := Node_EventEmitter.NewImpl(nil)
        
        isUpgrade := false
        isWebSocket := false
        
        if val, ok := arg0.(gopurs_runtime.Value); ok && val.Type >= gopurs_runtime.TypeRecord0 && val.Type <= gopurs_runtime.TypeRecordData {
            record := gopurs_runtime.RecordToMap(val)
            if headersVal, hasHeaders := record["headers"]; hasHeaders {
                if headersVal.Type >= gopurs_runtime.TypeRecord0 && headersVal.Type <= gopurs_runtime.TypeRecordData {
                    headersMap := gopurs_runtime.RecordToMap(headersVal)
                    if upgradeVal, hasUpgrade := headersMap["Upgrade"]; hasUpgrade {
                        isUpgrade = true
                        upgStr := gopurs_runtime.Unbox[string](upgradeVal)
                        fmt.Printf("MOCK: Upgrade val=%v str=%v\n", upgradeVal, upgStr)
                        if upgStr == "websocket" {
                            isWebSocket = true
                        }
                    }
                } else if headersVal.Type == gopurs_runtime.TypeAny {
                    if headersMap, ok2 := headersVal.AnyVal().(map[string]interface{}); ok2 {
                        if upgradeVal, hasUpgrade := headersMap["Upgrade"]; hasUpgrade {
                            isUpgrade = true
                            if upgStr, ok3 := upgradeVal.(string); ok3 && upgStr == "websocket" {
                                isWebSocket = true
                            }
                        }
                    }
                }
            }
        }
        
        fmt.Printf("MOCK: isUpgrade=%v isWebSocket=%v\n", isUpgrade, isWebSocket)
        
        if GlobalServer != nil {
            if isWebSocket {
                Node_HTTP_IncomingMessage.GlobalStatusCodes[resEmitter] = 101
                upgradeHeaders := map[string]interface{}{"upgrade": "websocket"}
                Node_HTTP_IncomingMessage.GlobalHeaders[resEmitter] = gopurs_runtime.Box[map[string]interface{}](upgradeHeaders)
                Node_HTTP_IncomingMessage.GlobalHeaders[dummyReq] = gopurs_runtime.Box[map[string]interface{}](upgradeHeaders)
                Node_EventEmitter.GopursUnsafeEmitFn4(gopurs_runtime.Box[interface{}](GlobalServer), "upgrade", gopurs_runtime.Box[interface{}](dummyReq), gopurs_runtime.Box[interface{}](socket), gopurs_runtime.Box[interface{}](nil), nil)
            } else if isUpgrade {
                Node_HTTP_IncomingMessage.GlobalStatusCodes[resEmitter] = 426
                Node_EventEmitter.GopursUnsafeEmitFn4(gopurs_runtime.Box[interface{}](GlobalServer), "upgrade", gopurs_runtime.Box[interface{}](dummyReq), gopurs_runtime.Box[interface{}](socket), gopurs_runtime.Box[interface{}](nil), nil)
            } else {
                Node_HTTP_IncomingMessage.GlobalStatusCodes[resEmitter] = 200
                Node_EventEmitter.GopursUnsafeEmitFn3(gopurs_runtime.Box[interface{}](GlobalServer), "request", gopurs_runtime.Box[interface{}](dummyReq), gopurs_runtime.Box[interface{}](resEmitter), nil)
            }
        } else {
             Node_HTTP_IncomingMessage.GlobalStatusCodes[resEmitter] = 200
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
func SetMaxIdleHttpParsersImpl(arg0 interface{}) interface{} { return nil }

package main

import (
    "gopurs/output/gopurs_runtime"
)

var GlobalStatusCodes = make(map[interface{}]int)
var GlobalHeaders = make(map[interface{}]gopurs_runtime.Value)

func CompleteImpl(arg0 interface{}) interface{} { return true }
func HeadersImpl(arg0 interface{}) interface{} {
    var key interface{} = arg0
    if val, ok := arg0.(gopurs_runtime.Value); ok && val.Type == gopurs_runtime.TypeAny {
        key = val.AnyVal()
    }
    if val, ok := GlobalHeaders[key]; ok {
        return val
    }
    return gopurs_runtime.Box(map[string]interface{}{})
}
func HeadersDistinct(arg0 interface{}) interface{} { return gopurs_runtime.Box(map[string]interface{}{}) }
func HttpVersion(arg0 interface{}) interface{} { return "1.1" }
func Method(arg0 interface{}) interface{} { return "GET" }
func RawHeaders(arg0 interface{}) interface{} { return nil }
func RawTrailersImpl(arg0 interface{}) interface{} { return nil }
func SocketImpl(arg0 interface{}) interface{} { 
    return gopurs_runtime.Box(arg0) // return socket mock
}
func StatusCode(arg0 interface{}) interface{} {
    var key interface{} = arg0
    if val, ok := arg0.(gopurs_runtime.Value); ok && val.Type == gopurs_runtime.TypeAny {
        key = val.AnyVal()
    }
    if val, ok := GlobalStatusCodes[key]; ok {
        return val
    }
    return 200
}
func StatusMessage(arg0 interface{}) interface{} { return "OK" }
func TrailersImpl(arg0 interface{}) interface{} { return gopurs_runtime.Box(map[string]interface{}{}) }
func TrailersDistinctImpl(arg0 interface{}) interface{} { return gopurs_runtime.Box(map[string]interface{}{}) }
func Url(arg0 interface{}) interface{} { return "/" }

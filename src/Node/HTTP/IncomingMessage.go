package main

import (
    "net/http"
    "strings"

    "gopurs/output/gopurs_runtime"
)

type HttpResponseGetter interface {
	GetResponse() *http.Response
}

func getReqOrRes(arg0 interface{}) (*http.Request, *http.Response) {
    emitter := gopurs_runtime.Unbox[*EventEmitter](arg0)
    if req, ok := emitter.Any.(*http.Request); ok {
        return req, nil
    }
    if res, ok := emitter.Any.(*http.Response); ok {
        return nil, res
    }
    if getter, ok := emitter.Any.(HttpResponseGetter); ok {
        return nil, getter.GetResponse()
    }
    return nil, nil
}

func CompleteImpl(arg0 interface{}) interface{} { return true }

func HeadersImpl(arg0 interface{}) interface{} {
    req, res := getReqOrRes(arg0)
    var header http.Header
    if req != nil {
        header = req.Header
    } else if res != nil {
        header = res.Header
    }
    
    headersMap := make(map[string]gopurs_runtime.Value)
    if header != nil {
        for k, v := range header {
            if len(v) > 0 {
                lowerK := strings.ToLower(k)
                if lowerK == "set-cookie" {
                    arr := make([]gopurs_runtime.Value, len(v))
                    for i, s := range v {
                        arr[i] = gopurs_runtime.Box(s)
                    }
                    headersMap[lowerK] = gopurs_runtime.Box(arr)
                } else {
                    headersMap[lowerK] = gopurs_runtime.Box(v[0])
                }
            }
        }
    }
    return gopurs_runtime.Box(headersMap)
}

func HeadersDistinct(arg0 interface{}) interface{} { return HeadersImpl(arg0) }

func HttpVersion(arg0 interface{}) interface{} {
    req, res := getReqOrRes(arg0)
    if req != nil {
        return strings.TrimPrefix(req.Proto, "HTTP/")
    } else if res != nil {
        return strings.TrimPrefix(res.Proto, "HTTP/")
    }
    return "1.1"
}

func Method(arg0 interface{}) interface{} {
    req, _ := getReqOrRes(arg0)
    if req != nil {
        return req.Method
    }
    return ""
}

func RawHeaders(arg0 interface{}) interface{} { return nil }
func RawTrailersImpl(arg0 interface{}) interface{} { return nil }
func SocketImpl(arg0 interface{}) interface{} {
    emitter := gopurs_runtime.Unbox[*EventEmitter](arg0)
    return gopurs_runtime.Box(emitter) // Return the emitter itself as a dummy socket if none available
}

func StatusCode(arg0 interface{}) interface{} {
    _, res := getReqOrRes(arg0)
    if res != nil {
        return int64(res.StatusCode)
    }
    return int64(200)
}

func StatusMessage(arg0 interface{}) interface{} {
    _, res := getReqOrRes(arg0)
    if res != nil {
        return strings.TrimPrefix(res.Status, "200 ")
    }
    return "OK"
}

func TrailersImpl(arg0 interface{}) interface{} { return gopurs_runtime.Box(map[string]interface{}{}) }
func TrailersDistinctImpl(arg0 interface{}) interface{} { return gopurs_runtime.Box(map[string]interface{}{}) }

func Url(arg0 interface{}) interface{} {
    req, _ := getReqOrRes(arg0)
    if req != nil {
        if req.RequestURI != "" {
            return req.RequestURI
        }
        return req.URL.String()
    }
    return ""
}

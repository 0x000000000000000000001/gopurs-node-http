package main

import (
	"fmt"
	"gopurs/output/Node.EventEmitter"
	"gopurs/output/gopurs_runtime"
	"unsafe"
    "io"
)

func main() {
	emitter := Node_EventEmitter.NewImpl(nil).(*Node_EventEmitter.EventEmitter)
    pr, pw := io.Pipe()
    _ = pr
    emitter.Any = pw
	var writable interface{} = emitter

	// How gopurs boxes interfaces:
	val := gopurs_runtime.Value{Type: gopurs_runtime.TypeForeign, UnsafePtr: unsafe.Pointer(&writable)}

	w := gopurs_runtime.Unbox[any](val)
	
	fmt.Printf("Type of w: %T\n", w)
    if e, ok := w.(*Node_EventEmitter.EventEmitter); ok {
        fmt.Printf("Cast to EventEmitter ok! Any is %T\n", e.Any)
        if ic, ok := e.Any.(io.Closer); ok {
            fmt.Printf("Cast to io.Closer ok!\n")
            ic.Close()
            fmt.Println("Closed successfully")
        }
    } else {
        fmt.Println("Cast failed")
    }
}

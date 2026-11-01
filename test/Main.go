package Test_Main

import "os"

func testStdout() interface{} {
	emitter := Node_EventEmitter_NewImpl(nil).(*EventEmitter)
	emitter.Any = os.Stdout
	return emitter
}

var Stdout = testStdout()

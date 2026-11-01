package Node_EventEmitter

import (
	"errors"
	"gopurs/output/gopurs_runtime"
	"io"
	"net/http"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestStdoutIsNativeWritableStream(t *testing.T) {
	emitter := nodeEventEmitter_extractEventEmitter(gopurs_runtime.Box(Stdout))
	writer, ok := emitter.Any.(io.Writer)
	if !ok || writer != os.Stdout {
		t.Fatalf("test stdout must carry os.Stdout as a native writer, got %T", emitter.Any)
	}
}

type responseReadStep struct {
	data string
	err  error
}

type responseBodyProbe struct {
	mu     sync.Mutex
	steps  []responseReadStep
	reads  int
	closes int
	read   chan struct{}
	closed chan struct{}
}

func (b *responseBodyProbe) Read(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.reads++
	select {
	case b.read <- struct{}{}:
	default:
	}
	if len(b.steps) == 0 {
		return 0, io.EOF
	}
	step := b.steps[0]
	b.steps = b.steps[1:]
	return copy(p, step.data), step.err
}

func (b *responseBodyProbe) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closes++
	if b.closes == 1 {
		close(b.closed)
	}
	return nil
}

func responseFixture(steps ...responseReadStep) (*ResponseWrapperHTTPS, *EventEmitter, *responseBodyProbe) {
	body := &responseBodyProbe{steps: steps, read: make(chan struct{}, 8), closed: make(chan struct{})}
	emitter := NewImpl(nil).(*EventEmitter)
	response := &ResponseWrapperHTTPS{
		Response:   &http.Response{Body: body},
		ReadCloser: body,
		emitter:    emitter,
	}
	emitter.Any = response
	return response, emitter, body
}

type responseEvents struct {
	order  []string
	chunks [][]byte
	err    error
}

func observeResponse(emitter *EventEmitter) *responseEvents {
	events := &responseEvents{}
	boxed := gopurs_runtime.Box(emitter)
	GopursUnsafeOn(boxed, "end", gopurs_runtime.Func(func(_ gopurs_runtime.Value) gopurs_runtime.Value {
		events.order = append(events.order, "end")
		return gopurs_runtime.Value{}
	}), nil)
	GopursUnsafeOn(boxed, "error", gopurs_runtime.Func(func(value gopurs_runtime.Value) gopurs_runtime.Value {
		events.order = append(events.order, "error")
		events.err = gopurs_runtime.Unbox[error](value)
		return gopurs_runtime.Value{}
	}), nil)
	return events
}

func (events *responseEvents) listenForData(emitter *EventEmitter) {
	GopursUnsafeOn(gopurs_runtime.Box(emitter), "data", gopurs_runtime.Func(func(value gopurs_runtime.Value) gopurs_runtime.Value {
		events.order = append(events.order, "data")
		// Keep the actual chunk: a reused read buffer would corrupt this value.
		events.chunks = append(events.chunks, gopurs_runtime.Unbox[[]byte](value))
		return gopurs_runtime.Value{}
	}), nil)
}

func assertNoResponseRead(t *testing.T, body *responseBodyProbe) {
	t.Helper()
	select {
	case <-body.read:
		t.Fatal("response body read before both readiness and a data listener")
	case <-time.After(20 * time.Millisecond):
	}
}

func awaitResponseClosed(t *testing.T, body *responseBodyProbe, reads int) {
	t.Helper()
	select {
	case <-body.closed:
	case <-time.After(time.Second):
		t.Fatal("response body did not finish and close")
	}
	body.mu.Lock()
	defer body.mu.Unlock()
	if body.reads != reads || body.closes != 1 {
		t.Fatalf("body reads/closes = %d/%d, want %d/1", body.reads, body.closes, reads)
	}
}

func TestHTTPSResponseWaitsForReady(t *testing.T) {
	response, emitter, body := responseFixture(responseReadStep{"payload", io.EOF})
	events := observeResponse(emitter)
	events.listenForData(emitter)
	assertNoResponseRead(t, body)
	response.markResponseReady()
	awaitResponseClosed(t, body, 1)
	if !reflect.DeepEqual(events.order, []string{"data", "end"}) || string(events.chunks[0]) != "payload" {
		t.Fatalf("unexpected response events: %+v", events)
	}
}

func TestHTTPSResponseStartsWhenDataListenerArrivesLate(t *testing.T) {
	response, emitter, body := responseFixture(responseReadStep{"late", io.EOF})
	events := observeResponse(emitter)
	response.markResponseReady()
	assertNoResponseRead(t, body)
	events.listenForData(emitter)
	awaitResponseClosed(t, body, 1)
	if !reflect.DeepEqual(events.order, []string{"data", "end"}) || string(events.chunks[0]) != "late" {
		t.Fatalf("late listener did not receive the response: %+v", events)
	}
}

func TestHTTPSResponseKeepsChunksAndEmitsDataBeforeEOF(t *testing.T) {
	response, emitter, body := responseFixture(
		responseReadStep{"first", nil},
		responseReadStep{"SECOND", io.EOF},
	)
	events := observeResponse(emitter)
	events.listenForData(emitter)
	response.markResponseReady()
	awaitResponseClosed(t, body, 2)
	if !reflect.DeepEqual(events.order, []string{"data", "data", "end"}) {
		t.Fatalf("events = %v", events.order)
	}
	if got := []string{string(events.chunks[0]), string(events.chunks[1])}; !reflect.DeepEqual(got, []string{"first", "SECOND"}) {
		t.Fatalf("retained chunks were changed or lost: %v", got)
	}
}

func TestHTTPSResponseReadErrorDoesNotEmitEnd(t *testing.T) {
	wantError := errors.New("body read failed")
	response, emitter, body := responseFixture(responseReadStep{"partial", wantError})
	events := observeResponse(emitter)
	events.listenForData(emitter)
	response.markResponseReady()
	awaitResponseClosed(t, body, 1)
	if !reflect.DeepEqual(events.order, []string{"data", "error"}) || !errors.Is(events.err, wantError) {
		t.Fatalf("read error was changed or emitted end: %+v", events)
	}
	if string(events.chunks[0]) != "partial" {
		t.Fatalf("data returned with read error was lost: %q", events.chunks)
	}
}

func TestHTTPSResponseEmptyBodyEmitsEndAndCloses(t *testing.T) {
	response, emitter, body := responseFixture(responseReadStep{"", io.EOF})
	events := observeResponse(emitter)
	events.listenForData(emitter)
	response.markResponseReady()
	awaitResponseClosed(t, body, 1)
	if !reflect.DeepEqual(events.order, []string{"end"}) {
		t.Fatalf("empty body events = %v, want end only", events.order)
	}
}

func TestHTTPSResponseStartsOnlyOneReadLoop(t *testing.T) {
	response, emitter, body := responseFixture(responseReadStep{"once", io.EOF})
	events := observeResponse(emitter)
	events.listenForData(emitter)
	var ready sync.WaitGroup
	for i := 0; i < 16; i++ {
		ready.Add(1)
		go func() {
			defer ready.Done()
			response.markResponseReady()
			response.OnListenerAdded("data")
		}()
	}
	ready.Wait()
	awaitResponseClosed(t, body, 1)
	response.markResponseReady()
	response.OnListenerAdded("data")
	awaitResponseClosed(t, body, 1)
	if !reflect.DeepEqual(events.order, []string{"data", "end"}) {
		t.Fatalf("duplicate read loop events: %v", events.order)
	}
}

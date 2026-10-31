import { execFileSync } from "node:child_process";
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

const runtimeRoot = process.env.GOPURS_ROOT ?? fileURLToPath(new URL("../../gopurs/", import.meta.url));
const libraryRoot = process.env.GOPURS_LIBRARY_ROOT ?? fileURLToPath(new URL("../../", import.meta.url));
const source = readFileSync(fileURLToPath(new URL("../src/Node/HTTPS.go", import.meta.url)), "utf8");
const start = source.indexOf("type ResponseWrapperHTTPS struct {");
const end = source.indexOf("func performClientRequestHTTPS(");
if (start < 0 || end <= start) throw new Error("Cannot locate the HTTPS response wrapper in its source");

const workspace = mkdtempSync(join(tmpdir(), "gopurs-https-response-"));
try {
  mkdirSync(join(workspace, "gopurs_runtime"));
  copyFileSync(join(runtimeRoot, "runtime/runtime.go"), join(workspace, "gopurs_runtime/runtime.go"));
  copyFileSync(join(libraryRoot, "gopurs-node-event-emitter/src/Node/EventEmitter.go"), join(workspace, "emitter.go"));
  // Compile the actual wrapper and emitter together; aliases only reproduce the
  // compiler's names for the three emitter functions referenced by this wrapper.
  writeFileSync(join(workspace, "response.go"), `package Node_EventEmitter
import (
  "io"
  "net/http"
  "sync"
  "gopurs/output/gopurs_runtime"
)
var Node_EventEmitter_ListenerCountImpl = ListenerCountImpl
var Node_EventEmitter_GopursUnsafeEmitFn1 = GopursUnsafeEmitFn1
var Node_EventEmitter_GopursUnsafeEmitFn2 = GopursUnsafeEmitFn2
${source.slice(start, end)}`);
  copyFileSync(fileURLToPath(new URL("./https-response-native_test.go", import.meta.url)), join(workspace, "response_test.go"));
  writeFileSync(join(workspace, "go.mod"), "module gopurs/output\n\ngo 1.22\n");
  const race = process.env.GOPURS_NATIVE_TEST_RACE === "1" ? ["-race"] : [];
  execFileSync("go", ["test", ...race, "-count=1", "-timeout=10s", "-v", "."], {
    cwd: workspace,
    stdio: "inherit",
    timeout: 120_000,
    env: {
      ...process.env,
      GOCACHE: process.env.GOCACHE ?? join(workspace, "go-cache"),
      GOWORK: "off",
      GOTOOLCHAIN: "local",
      GOPROXY: "off",
      GOSUMDB: "off",
    },
  });
} finally {
  rmSync(workspace, { recursive: true, force: true });
}

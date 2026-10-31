#!/usr/bin/env node
import { spawn } from "node:child_process";
import { readFileSync } from "node:fs";
import https from "node:https";
import { fileURLToPath } from "node:url";

const certificate = fileURLToPath(new URL("./fixtures/localhost-cert.pem", import.meta.url));
const privateKey = fileURLToPath(new URL("./fixtures/localhost-key.pem", import.meta.url));
const timeoutMs = Number(process.env.GOPURS_TEST_HTTPS_TIMEOUT_MS ?? 30_000);
const outputLimit = 1024 * 1024;
const setupTimeoutMs = 5_000;
const sockets = new Set();
const requests = { https: 0, cookies: 0 };
const stdout = [];
let outputBytes = 0;
let child;
let server;
let cancelStage;
let receivedSignal;

function interrupted(signal) {
  if (receivedSignal) return;
  receivedSignal = signal;
  if (cancelStage) cancelStage(new Error(`Interrupted by ${signal}`));
}

const onInterrupt = () => interrupted("SIGINT");
const onTerminate = () => interrupted("SIGTERM");

function signalChild(signal) {
  if (!child || !child.pid || child.exitCode !== null || child.signalCode !== null) return;
  try {
    if (process.platform === "win32") child.kill(signal);
    else process.kill(-child.pid, signal);
  } catch (error) {
    if (error.code !== "ESRCH") throw error;
  }
}

function listen() {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => finish(new Error("HTTPS fixture setup timed out")), setupTimeoutMs);
    const finish = error => {
      clearTimeout(timer);
      cancelStage = undefined;
      server.removeListener("error", finish);
      if (error) reject(error);
      else resolve();
    };
    cancelStage = finish;
    server.once("error", finish);
    server.listen(0, "127.0.0.1", () => finish());
  });
}

function run(command, args) {
  return new Promise((resolve, reject) => {
    let finished = false;
    const timer = setTimeout(() => finish(new Error(`HTTPS fixture child timed out after ${timeoutMs} ms`)), timeoutMs);
    const finish = error => {
      if (finished) return;
      finished = true;
      clearTimeout(timer);
      cancelStage = undefined;
      if (error) reject(error);
      else resolve();
    };
    cancelStage = finish;
    child = spawn(command, args, {
      detached: process.platform !== "win32",
      stdio: ["inherit", "pipe", "pipe"],
      env: {
        ...process.env,
        GOPURS_TEST_HTTPS_BASE_URL: `https://127.0.0.1:${server.address().port}`,
        SSL_CERT_FILE: certificate,
        NODE_EXTRA_CA_CERTS: certificate,
      },
    });
    const forward = (chunk, destination, capture) => {
      if (finished) return;
      outputBytes += chunk.length;
      if (outputBytes > outputLimit) {
        finish(new Error(`HTTPS fixture child exceeded ${outputLimit} bytes of output`));
        return;
      }
      if (capture) stdout.push(chunk);
      destination.write(chunk);
    };
    child.stdout.on("data", chunk => forward(chunk, process.stdout, true));
    child.stderr.on("data", chunk => forward(chunk, process.stderr, false));
    child.once("error", finish);
    child.once("close", (code, signal) => {
      if (code !== 0) finish(new Error(`HTTPS fixture child failed (${signal ?? `exit ${code}`})`));
      else finish();
    });
    server.on("error", finish);
  });
}

async function cleanup() {
  cancelStage = undefined;
  signalChild("SIGTERM");
  for (const socket of sockets) socket.destroy();
  if (server) {
    await new Promise(resolve => {
      const timer = setTimeout(resolve, 1_000);
      server.close(() => {
        clearTimeout(timer);
        resolve();
      });
    });
  }
  if (child && child.pid) {
    await new Promise(resolve => setTimeout(resolve, 300));
    signalChild("SIGKILL");
    child.stdout.destroy();
    child.stderr.destroy();
  }
  process.removeListener("SIGINT", onInterrupt);
  process.removeListener("SIGTERM", onTerminate);
}

async function main() {
  const [command, ...args] = process.argv.slice(2);
  if (!command) throw new Error("Usage: node test/https-fixture.mjs COMMAND [ARG ...]");
  if (!Number.isSafeInteger(timeoutMs) || timeoutMs < 1 || timeoutMs > 3_600_000) {
    throw new Error("GOPURS_TEST_HTTPS_TIMEOUT_MS must be an integer from 1 to 3600000");
  }
  if (process.env.NODE_TLS_REJECT_UNAUTHORIZED === "0") {
    throw new Error("Remove NODE_TLS_REJECT_UNAUTHORIZED=0: this fixture requires certificate verification");
  }
  process.on("SIGINT", onInterrupt);
  process.on("SIGTERM", onTerminate);
  server = https.createServer({
    key: readFileSync(privateKey),
    cert: readFileSync(certificate),
    handshakeTimeout: 3_000,
  }, (request, response) => {
    if (request.method !== "GET") {
      response.writeHead(405, { Allow: "GET" });
      response.end("Method not allowed\n");
      return;
    }
    let route;
    let body;
    if (request.url === "/https?case=tls") {
      route = "https";
      body = "local HTTPS response\n";
    } else if (request.url === "/cookies") {
      route = "cookies";
      body = "local cookies response\n";
      response.setHeader("Set-Cookie", [
        "cookie1=firstcookie; Path=/",
        "cookie2=secondcookie; HttpOnly; Secure",
      ]);
    } else {
      response.writeHead(404);
      response.end("Not found\n");
      return;
    }
    response.setHeader("Content-Type", "text/plain; charset=utf-8");
    response.setHeader("Content-Length", Buffer.byteLength(body));
    response.once("finish", () => { requests[route] += 1; });
    response.end(body);
  });
  server.setTimeout(3_000);
  server.headersTimeout = 3_000;
  server.requestTimeout = 3_000;
  server.on("connection", socket => {
    sockets.add(socket);
    socket.once("close", () => sockets.delete(socket));
  });
  await listen();
  if (receivedSignal) throw new Error(`Interrupted by ${receivedSignal}`);
  await run(command, args);
  const output = Buffer.concat(stdout).toString("utf8");
  for (const [route, marker] of [["https", "PASS local HTTPS"], ["cookies", "PASS local cookies"]]) {
    if (requests[route] === 0) throw new Error(`No successful request to the local ${route} route`);
    if (!output.split(/\r?\n/).includes(marker)) throw new Error(`Child did not report ${marker}`);
  }
  console.log("Local HTTPS fixture: TLS and cookies checks completed.");
}

try {
  await main();
} catch (error) {
  console.error(`Local HTTPS fixture: ${error.message}`);
  process.exitCode = receivedSignal === "SIGINT" ? 130 : receivedSignal === "SIGTERM" ? 143 : 1;
} finally {
  try {
    await cleanup();
  } catch (error) {
    console.error(`Local HTTPS fixture cleanup: ${error.message}`);
    process.exitCode = 1;
  }
}

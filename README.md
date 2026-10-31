# gopurs-node-http

## Local Go development

This checkout is part of the gopurs library family. Use the
[local Go development guide](../gopurs/README.md#develop-one-library-locally)
for toolchain setup, sibling dependencies, Spago configuration and Go commands.
The existing npm, Bower and Dhall commands below retain their JavaScript or
upstream roles.

The Go `bin/test` and JavaScript `npm test` commands run the HTTPS client tests
through `test/https-fixture.mjs`. It starts an independent HTTPS server on an
ephemeral loopback port, trusts its test certificate only in the child process,
and requires successful status/body/cookie assertions. These tests do not call
public HTTP services. See [the fixture notes](test/fixtures/README.md) for the
certificate's scope and direct runner usage.

Before compilation, `bin/test` also runs `test/https-response-native.mjs`
against the Go response wrapper and sibling event-emitter source. It checks
body delivery, empty responses, read errors, chunk ownership and one-time
startup. Set `GOPURS_NATIVE_TEST_RACE=1` to include Go's race detector when
the local Go/C toolchain supports it. This covers HTTPS response `data`/`end`
delivery; it does not establish support for the other stream APIs such as
pause/resume, piping or pull-based reads.


[![Latest release](http://img.shields.io/github/release/purescript-node/purescript-node-http.svg)](https://github.com/purescript/purescript-node-http/releases)
[![Build status](https://github.com/purescript-node/purescript-node-http/workflows/CI/badge.svg?branch=master)](https://github.com/purescript-node/purescript-node-http/actions?query=workflow%3ACI+branch%3Amaster)
[![Pursuit](https://pursuit.purescript.org/packages/purescript-node-http/badge)](https://pursuit.purescript.org/packages/purescript-node-http)

A wrapper for Node's [HTTP](https://nodejs.org/dist/latest-v18.x/docs/api/http.html) and [HTTPS](https://nodejs.org/dist/latest-v18.x/docs/api/https.html) APIs.

## Installation

```
spago install node-http
```

## Documentation

Module documentation is [published on Pursuit](http://pursuit.purescript.org/packages/purescript-node-http).

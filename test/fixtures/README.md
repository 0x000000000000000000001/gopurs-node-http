# Local HTTPS test fixture

`../https-fixture.mjs` requires Node.js **14.18 or newer** and uses only its
standard library. It binds an ephemeral port on `127.0.0.1` and starts the direct
test command given after the script name. Run compilation before this wrapper;
its timeout covers execution of the test command. It closes the server when the
command finishes or the wrapper receives SIGINT/SIGTERM. While the direct child
is still running, cleanup signals its process group on POSIX, or only the direct
child on Windows. Once the direct child has exited, no further signal is sent:
cleanup does **not** guarantee termination of descendants that outlive it. Use
this wrapper for a direct Go test binary or a direct Node test process, not a
launcher that exits while leaving tests running.

Run from the repository root:

```sh
node test/https-fixture.mjs ./output/go_test_app
node test/https-fixture.mjs node --input-type=module -e "import { main } from './output/Test.Main/index.js'; main();"
```

The child receives these variables:

- `GOPURS_TEST_HTTPS_BASE_URL`: `https://127.0.0.1:<ephemeral-port>`.
- `SSL_CERT_FILE`: absolute path to `localhost-cert.pem`, for Go's TLS verifier.
- `NODE_EXTRA_CA_CERTS`: the same certificate, for Node's TLS verifier.

Certificate verification stays enabled. The wrapper rejects an inherited
`NODE_TLS_REJECT_UNAUTHORIZED=0`. It does not install certificates into any
system trust store or contact a public endpoint.

The only successful routes are:

| Request | Status | Body | Set-Cookie |
| --- | --- | --- | --- |
| `GET /https?case=tls` | 200 | `local HTTPS response\n` | absent |
| `GET /cookies` | 200 | `local cookies response\n` | `cookie1=firstcookie; Path=/` and `cookie2=secondcookie; HttpOnly; Secure` |

Other GET routes return 404; other methods return 405. A successful run requires
the child to exit zero, receive a successful response from both routes, and
print both exact stdout lines **after checking the responses**:

```text
PASS local HTTPS
PASS local cookies
```

The wrapper checks the route counts and marker lines; the child is responsible
for asserting the response status, bodies, cookies, and TLS behavior. Child
stdout and stderr are streamed, with a combined 1 MiB limit; stdout is also
captured within that limit to check the markers. The child has 30 seconds by
default. `GOPURS_TEST_HTTPS_TIMEOUT_MS` accepts an integer from 1 to 3600000 for
slower commands or timeout checks. Server setup is limited to 5 seconds, and
idle sockets/TLS handshakes to 3 seconds. Cleanup signals the spawned child's
process group only while that direct child is still running. SIGKILL follows a
short SIGTERM grace period only if the direct child has not yet exited.

## Public test certificate and private key

`localhost-key.pem` is an **intentionally public, test-only private key**. It is
not a credential and must never protect a deployed service or real data.
`localhost-cert.pem` is its self-signed RSA-2048 certificate, signed with SHA-256.
It is a test CA and server certificate with these explicit properties:

- Valid from **2026-10-31 17:35:42 UTC** through **2126-10-07 17:35:42 UTC**.
- Subject alternative names: **DNS:localhost**, **IP:127.0.0.1**.
- Basic constraints: `CA:TRUE`; extended key usage: `serverAuth`.
- Key usage: digital signature, key encipherment, certificate signing.

The long validity avoids routine expiry in this isolated fixture. Trust is
supplied only to the test child through the environment above. The key and
certificate are committed test data; no download or certificate generation is
needed during tests.

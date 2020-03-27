# TODO:
- reset or close the client session if the auth hook performs a logout, reloading, ... 
- log cancel calls? client and session?
- send cancel events over an internal stream. otherwise sending huge payloads will block the stream.
- in both rpc implementations the error handling is bad: wrap errors and shorten messages.
- catch panic: check if message with stack trace and use a different logging method.
- disconnect session after TTL
- more logs (client & service)
- use tls.dialContext in yamux, once the new go version has come out
- add maxmsgsize to calls (and streams?), but optional!!! take config value as default.
- update the comments and remove unneeded parts (like old pkg).
- finish documenting
- write tests for packages:
  - internal.flusher
- Walk through TODOs in code and resolve them
- Include go report in readme (and fix issues that it reports beforehand)
- Options must allow to set Timeout == 0, to prevent any timeout
- add orbit fmt cmd for .orbit files

### Testing
- Add reflection based tests that test the API structs and instantiate them with each field initialized, then marshal them using msgp and check the result with require.Exactly. This way, we can detect whether go generate has been executed for newly added structs/properties



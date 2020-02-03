# TODO:
- use tls.dialContext in yamux, once the new go version has come out
- add maxmsgsize to calls (and streams?), but optional!!! take config value as default.
- update the comments and remove unneeded parts (like old pkg).
- add & update tests

## OLD:

### General
- use the orbit logger for the control and event package
- finish documenting
- clarify orbit/config.go, since logger defaults to os.Stderr (shouldn't that be zerolog?)
- discuss logging in general, make sure that orbit uses a default config level. Right now, zerolog prints '???' as level
- write tests for packages:
  - orbit
  - internal.flusher
- Walk through TODOs in code and resolve them
- Add load balancing interface
- Include go report in readme (and fix issues that it reports beforehand)

### Testing
- Add reflection based tests that test the API structs and instantiate them with each field initialized, then marshal them using msgp and check the result with require.Exactly. This way, we can detect whether go generate has been executed for newly added structs/properties

### Samples 
- Add sample that shows error handling
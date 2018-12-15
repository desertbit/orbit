### General
- change license? at least for the examples?
- use the orbit logger for the control and event package
- global events
- finish documenting
- clarify orbit/config.go, since logger defaults to os.Stderr (shouldn't that be zerolog?)
- discuss logging in general, make sure that orbit uses a default config level. Right now, zerolog prints '???' as level
- write tests for packages:
  - orbit
  - flusher
- Walk through TODOs in code and resolve them
- Add request cancellation (e.g. with context)
- Add load balancing interface

### Samples
- Add "chatroom" example, where one client writes to console, and others receive event (multicast), needs global events pkg first 
- Add sample that shows error handling
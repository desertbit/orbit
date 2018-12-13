### General
- change license to MIT
- use the orbit logger for the control and event package
- finish documenting
- clarify orbit/config.go, since logger defaults to os.Stderr (shouldn't that be zerolog?)
- discuss logging in general, make sure that orbit uses a default config level. Right now, zerolog prints '???' as level
- write tests for packages:
  - orbit
  - internal.flusher
- Walk through TODOs in code and resolve them
- Add request cancellation (e.g. with context)

### Samples 
- Add sample that shows error handling
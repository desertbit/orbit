### General
- server config and add workers option value
- change license? at least for the examples?
- use the orbit logger for the control and event package
- global events
- finish documenting
- clarify orbit/config.go, since Codec is never used and logger defaults to os.Stderr (should'nt that be zerolog?)
- write tests
- think about adding all timeouts into the config.

### Samples
- use go survey instead of go input
- Rework completely
- Add sample that uses control pkg
- Add sample that uses signaler pkg (timebomb scenario, where countdown is signaled)
- Add "chatroom" example, where one client writes to console, and others receive event (multicast), needs global events pkg first 
- Add sample that shows error handling
- Add sample that shows HTTP usage
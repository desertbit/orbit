### General
- change license? at least for the examples?
- use the orbit logger for the control and event package
- global events
- finish documenting
- clarify orbit/config.go, since logger defaults to os.Stderr (shouldn't that be zerolog?)
- write tests
- think about adding all timeouts into the config.

### Control
- Some promise-like feature for CallAsync? Otherwise, signaler is always kind of the better alternative and CallAsync is not needed

### Samples
- Rework completely
- Add "chatroom" example, where one client writes to console, and others receive event (multicast), needs global events pkg first 
- Add sample that shows error handling
- Add sample that shows HTTP usage
### General
- change license? at least for the examples?
- use the orbit logger for the control and event package
- global events
- finish documenting
- clarify orbit/config.go, since logger defaults to os.Stderr (shouldn't that be zerolog?)
- write tests

### Samples
- Rework completely
- Add "chatroom" example, where one client writes to console, and others receive event (multicast), needs global events pkg first 
- Add sample that shows error handling
- Add sample that shows HTTP usage
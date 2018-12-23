### General
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
  - Client-side must be able to cancel an ongoing request
    - Hand in a request context to each Call
    - The context can contain a closer (for cancellation) and/or a timeout (for timeouts, replaces the current timeout methods)
    - When the closer is closed, the server must be notified of it, therefore additional data must be sent over the wire
  - Server-side must be able to listen to a cancel from the client to abort its request handling
    - Include a closer in the current context
    - Close this closer, if client has sent a cancel message
- Add load balancing interface

### Samples
- Add "chatroom" example, where one client writes to console, and others receive event (multicast), needs global events pkg first 
- Add sample that shows error handling
# Orbit - Interlink Remote Applications

Orbit provides a powerful backend to interlink remote applications with each other.  
It replaces connectionless RPC solutions with a multiplexed session-based solution that includes convenient features (such as raw streams, signals and more ...).  

Orbit generally does not expect you to use a strict client-server architecture. Both peers can call functions and trigger signals at the other peer. Therefore, it is possible to implement both client-server and peer-to-peer applications.


### Features

- Session-based
- Multiplexed connections with multiple channel streams and Keep-Alive (using [yamux](https://github.com/hashicorp/yamux))
- Plugable custom codecs for encoding and decoding (defaulting to MessagePack using [msgpack](https://github.com/msgpack/msgpack]))
- Use raw tcp streams to implement your own protocols, and/or
  - Use the control package for RPC-like approaches
  - Use efficient signals for event-based approaches
- Provide an authentication hook to easily authenticate your peers.
- Configure everything to your needs, such as:
  - your preferred logger
  - allowed message sizes
  - timeouts
  - ...
- Easy setup (check out the [sample](https://github.com/desertbit/orbit/tree/master/sample))

### Control - RPC
The control package provides an implementation for Remote Procedure Calls. This allows a peer to define functions that other connected peers can then call.

##### Setup
First, you need to setup a control on each peer. It is best to use either the provided `Init` or `InitMany` functions to do this.
```go
ctrl, _, err := orbitSession.Init(&orbit.Init{
    Control: orbit.InitControl{
        Funcs: map[string]control.Func{
            api.Action1: handleAction1,
        },
    },
})
if err != nil {
    return err
}

ctrl.Ready()
```
To use a custom initialization, check out the source code of the `InitMany` function to get a grasp of what needs to be done.

##### Synchronous Call
To make a synchronous call, simply trigger a call on _peer1_ to _peer2_ and then wait for the incoming response.
```go
// Call
ctx, err := s.ctrl.Call(api.Action1, &api.Action1Args{
	ID: 28
})
if err != nil {
    return
}

// Response
var response api.Action1Ret
err = ctx.Decode(&response)
if err != nil {
    return
}

print(response.SomeData)
```

_peer2_ might handle this call like this:
```go
func handleAction1(ctx *control.Context) (v interface{}, err error) {
	var args api.Action1Args
	err = ctx.Decode(&args)
	if err != nil {
		return
	}
	
	// handle the request ...
	
	v = &api.Action1Ret{
		SomeData: someData,
	}
	return
}
```
Note, that the `handleAction1` func of _peer2_ must have been added to the control of _peer2_ for the correct key. Check out the [Control Setup](#Setup) to see how to do this

##### Asynchronous Call
```go
// Call Async
ctx, err := s.ctrl.CallAsync(api.Action2, &api.Action2Args{
	ID: 28
})
if err != nil {
    return
}
```
In order to receive a result, the signaler package can be used, by triggering a signal that carries the response. Check out the next section for an example.

### Signaler - Events
TODO
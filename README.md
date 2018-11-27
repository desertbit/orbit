# Orbit - Interlink Remote Applications

Orbit provides a powerful backend to interlink remote applications with each other.  
It replaces connectionless RPC solutions with a multiplexed session-based solution that includes convenient features (such as raw streams, signals and more ...).  

Orbit generally does not expect you to use a strict client-server architecture. Both peers can call functions and trigger signals at the other peer. Therefore, it is possible to implement both client-server and peer-to-peer applications.


### Features

- Session-based
- Multiplexed connections with multiple channel streams and Keep-Alive (using [yamux](https://github.com/hashicorp/yamux))
- Plugable custom codecs for encoding and decoding (defaulting to MessagePack using [msgpack](https://github.com/msgpack/msgpack))
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
type Action1Args struct {
    ID int
}

type Action1Ret struct {
    SomeData string
}

func Action1() (data string, err error) {
    // Call
    ctx, err := s.ctrl.Call("Action1", &Action1Args{
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
    
    data = response.SomeData
    return 
}
```

_peer2_ might handle this call like this:
```go
func handleAction1(ctx *control.Context) (v interface{}, err error) {
    var args Action1Args
    err = ctx.Decode(&args)
    if err != nil {
        return
    }
    
    // handle the request ...
    
    v = &Action1Ret{
        SomeData: someData,
    }
    return
}
```
Note, that the `handleAction1` func of _peer2_ must have been added to the control of _peer2_ for the correct key. Check out the [Control Setup](#Setup) to see how to do this

##### Asynchronous Call
An asynchronous call is very similar to its synchronous counterpart. 
```go
type Action2Args struct {
    ID int
}

type Action2Ret struct {
    SomeData string
}

// Call Async
func Action2() error {
    callback := func(data interface{}, err error) {
        if err != nil {
            log.Fatal(err)
        }
        
        // Response
        var response Action2Ret
        err = ctx.Decode(&response)
        if err != nil {
            return
        }
        
        // handle data...
        println(response.SomeData)
    }
    
    return s.ctrl.CallAsync(
        "Action2", 
        &Action2Args{
            ID: 28,
        }, 
        callback,
    )
}
```
Inside the callback, you receive the response (or an error) and can handle it the same way as with the synchronous call.

### Signaler - Events
TODO
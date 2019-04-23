# Orbit - Interlink Remote Applications

[![GoDoc](https://godoc.org/github.com/desertbit/desertbit?status.svg)](https://godoc.org/github.com/desertbit/desertbit)
[![coverage](https://codecov.io/gh/desertbit/orbit/branch/master/graph/badge.svg)](https://codecov.io/gh/desertbit/orbit/branch/master)
[![license](https://img.shields.io/github/license/desertbit/orbit.svg)](https://opensource.org/licenses/MIT)

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

## Signaler - Events
The signaler package provides an implementation for sending events to remote peers. Under the hood, it uses the control package's `CallOneWay` function tFirst, you need to setup a control on each peer. It is best to use either the provided `Init` or `InitMany` functions to do this.o make calls without expecting a response.  

The signaler package adds a lot of convenient stuff, such as allowing peers to set filters on their events, or unregister from an event completely.

The code in the following sections is taken from the [sample](github.com/desertbit/orbit/sample).

### Setup
First, you need to setup a signaler on each peer. It is best to use either the provided `Init` or `InitMany` functions to do this.

We start with the peer that emits the signal, the **sender**:
```go
// Initialize the signaler and declare, which signals
// can be triggered on it.
_, sig, err := orbitSesion.Init(&orbit.Init{
    Signaler: orbit.InitSignaler{
        Signals: []orbit.InitSignal{
            {
                ID: "TimeBomb",
            },
        },
    },
})
if err != nil {
    return
}

// Start the signaler.
sig.Ready()
```
First, we initialize an orbit session using `Init` and we register a signaler on it that can emit the `"TimeBomb"` signal.   
If this succeeds, we start the signaler by calling its `Ready()` method, which starts the listen routines of the signaler.  

Now let us move on to the peer that receives the signal, the **receiver**:
```go
// Initialize the signaler and declare, which signals
// can be triggered on it.
_, sig, err := orbitSesion.Init(&orbit.Init{
    Signaler: orbit.InitSignaler{
        Signals: []orbit.InitSignal{},
    },
})
if err != nil {
    return
}

// Register handlers for events from the remote peer
_ = sig.OnSignalFunc("TimeBomb", onEventTimeBomb)

// Start the signaler.
sig.Ready()
``` 
Again, we need to initialize the signaler for this peer as well, however, we do not register any signals on it, since we only want to receive signals from the remote peer right now.  
Afterwards, we register a handler func for the `"TimeBomb"` signal, the `onEventTimeBomb` function.  
In the end, we start the signaler.

Here is the implementation of the `onEventTimeBomb` handler func:
```go
func onEventTimeBomb(ctx *signaler.Context) {
	var args api.TimeBombData
	err := ctx.Decode(&args)
	if err != nil {
		log.Printf("onEventTimeBomb error: %v", err)
		return
	}

	// Do something with the signal data...
}
```
It is identical to the control handler funcs, only that we do not return something to the caller, as signals are unidirectional.

Now, if we want to finally trigger our signal on the **sender**, we can do it like this:
```go
// Trigger the event.
args := &api.TimeBombData{
    Countdown: 5,
}

err = sig.TriggerSignal("TimeBomb", &args)
if err != nil {
    log.Printf("triggerSignal TimeBomb: %v", err)
    return
}
```
We call the `TriggerSignal` method on our signaler we defined at the beginning of this section. This sends the given arguments over the wire to our **receiver**, where the `onEventTimeBomb` handler func will be triggered.
# Orbit File
This file describes the syntax of `.orbit` files used for code generation.

## Table of Contents
1. [Service](#service)
    1. [Call](#call)
    2. [Stream](#stream)
2. [Type](#type)
    1. [Basic Type](#basic-type)
    1. [Reference Type](#reference-type)
    1. [Inline Type](#inline-type)
3. [Enum](#enum)
4. [Error](#error)

## Service
Per `.orbit` file, you must declare exactly one service.
```
service {
    url: "example.com:4848"
    call sayHi { ... }
    stream messages { ... }
}
```

- **url**  
Defines the url of the service, e.g. "127.0.0.1:4587", "examplecom/my-service:9999"  
Usage: `url: 'example.com/my-service:9999'`

### Call
Per service, you can declare as many calls as you want.
```
service {
    call sayHi {
        async
        timeout: 5s
        arg: {
            name string 'required,min=1'
        }
        maxArgSize: 50KB
        ret: someType
        maxRetSize: 10MiB
        errors: myError
    }
}
```
- **async** (default: false)  
Each async call gets executed on a separate stream, thus, it does not block other calls.  
Usage: `async`
- **arg** (default: none)  
The argument data sent to the service. Can either be an inline or reference type.  
Usage: `arg: { ... }` or `arg: refType`
- **maxArgSize** (default: MaxArgSize from options)  
The maximum allowed size of the arg data. Must only be used in conjunction with `async`  
Usage: `maxArgSize: <size>`, where _\<size\>_ is a [bytefmt string](https://github.com/cloudfoundry/bytefmt)  
Special value: `-1` -> no limit
- **ret** (default: none)  
The return data sent from the service. Can either be an inline or reference type.  
Usage: `ret: { ... }` or `ret: refType`
- **maxRetSize** (default: MaxRetSize from options)  
The maximum allowed size of the ret data. Must only be used in conjunction with `async`  
Usage: `maxRetSize: <size>`, where _\<size\>_ is a [bytefmt string](https://github.com/cloudfoundry/bytefmt)  
Special value: `-1` -> no limit
- **timeout** (default: CallTimeout from options)  
The maximum time a call may take to finish.  
Usage: `timeout: <duration>`, where _\<duration\>_ is a [go time duration formatted string](https://golang.org/pkg/time/#ParseDuration)
- **errors** (default: none)  
The [errors](#error) this call may return. A comma-separated list of error names.  
Usage: `errors: err1, err2, ...`  

### Stream
Per service, you can declare as many streams as you want.
```
service {
    stream messages {
        arg: {
            name string 'required,min=1'
        }
        maxArgSize: 50KB
        ret: someType
        maxRetSize: 10MiB
        errors: myError
    }
}
```
- **arg** (default: none)  
The argument data streamed to the service. Can either be an inline or reference type.  
Usage: `arg: { ... }` or `arg: refType`
- **maxArgSize** (default: MaxArgSize from options)  
The maximum allowed size of the arg data.  
Usage: `maxRetSize: <size>`, where _\<size\>_ is a [bytefmt string](https://github.com/cloudfoundry/bytefmt)  
Special value: `-1` -> no limit
- **ret** (default: none)  
The return data streamed from the service. Can either be an inline or reference type.  
Usage: `ret: { ... }` or `ret: refType`
- **maxRetSize** (default: MaxRetSize from options)  
The maximum allowed size of the ret data.  
Usage: `maxRetSize: <size>`, where _\<size\>_ is a [bytefmt string](https://github.com/cloudfoundry/bytefmt)  
Special value: `-1` -> no limit
- **errors** (default: none)  
The [errors](#error) this call may return. A comma-separated list of error names.  
Usage: `errors: err1, err2, ...`  

## Type
Per `.orbit` file, you can declare as many types as you want.
```
type someType {
    name string 'validate:"required,min=1"'
    ...
}
```
Each field of a type must have the following syntax: `<identifier> <datatype> '<struct-tag>'`  
- **identifier** (mandatory)  
The name of the field, must be unique within the type.  
- **datatype** (mandatory)  
The data type of the field. Can either be a basic type or a reference type.  
- **struct-tag** (optional)  
The struct tag field. This is directly converted to a go struct-tag. Use it to define JSON formatting etc. Orbit supports the [go-playground validator](https://github.com/go-playground/validator/), so you can write validation logic directly inside the tags.

### Basic Type
The following basic types are available.
- `bool`, `byte`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`, `string`
- `time`: Equal to go's [time.Time](https://golang.org/pkg/time/#Time)
- `duration`: Equal to go's [time.Duration](https://golang.org/pkg/time/#Duration)

### Reference Type
A type that has been declared using the above syntax can be referenced elsewhere by its `identifier`
```
type A {
    someField int
}

type B {
    a A
}

service {
    call test {
        arg: B
    }
}
```

### Inline Type
Calls and Streams may define types inline. Such types receive a generated `identifier` of the form `<Call/Stream Name><Arg/Ret>`, but they can not be referenced elsewhere.
```
service {
    call test {
        arg: {
            name string 'required'
        }
        ret: {
            age int
        }
    }
}
```

## Enum
Per `.orbit` file, you can declare as many enums as you want.
```
enum carBrand {
    bmw = 1
    audi = 2
    volvo = 3
    toyota = 4
    vw = 5
}
```
Each field of an enum must have the following syntax: `<name> = <identifier>`  
- **name** (mandatory)  
The name of the field, must be unique within the enum.
- **identifier** (mandatory)  
The identifier of the field, must be an unsigned integer and unique within the enum.  

An enum can be used similar to basic types.
```
type A {
    brand carBrand 'required'
}

service {
    call getBrand {
        ret: {
            carBrand
        }
    }
}
```

## Error
Per `.orbit` file, you can declare as many errors as you want.
```
errors {
    myFirstError = 1
    anotherOne = 2
}
```
Each field of an errors block must have the following syntax: `<name> = <number>`
- **name** (mandatory)  
The name of the error, must be unique across all error blocks.  
The go error text will contain the name of the error split up by CamelCase, e.g. error "myFirstError" is constructed as `errors.New("my first error")`
- **number** (mandatory)  
The status code used to transmit the error over the network, must be unique across all error blocks.
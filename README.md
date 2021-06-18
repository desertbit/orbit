# Orbit - Interlink Services

[![GoDoc](https://godoc.org/github.com/desertbit/orbit?status.svg)](https://godoc.org/github.com/desertbit/orbit)
[![coverage](https://codecov.io/gh/desertbit/orbit/branch/master/graph/badge.svg)](https://codecov.io/gh/desertbit/orbit/branch/master)
[![license](https://img.shields.io/github/license/desertbit/orbit.svg)](https://opensource.org/licenses/MIT)

Orbit provides a powerful, MIT-licensed networking backend to interlink services. It offers **RPC** like features and is primarily stateless. It aims to be light-weight, customizable and an alterantive to gRPC.

![LOGO](images/logo_small.png)

## Current Status
This project is under heavy development. The API may change at any time. 

## Features
- RPC Calls and Streams
- Easy API declaration using custom syntax in `.orbit` files
- Code generation
- Pluggable transport protocols (quic, yamux, custom)
- Pluggable codecs (msgpack, custom)
- Field validation using [go-playground validator](https://github.com/go-playground/validator/)
  
## Documentation
- [Orbit File](docs/orbit-file.md)

## Similar projects
- [gRPC](https://github.com/grpc/grpc-go)

## Thanks to
- `gopherize.me` for our great logo
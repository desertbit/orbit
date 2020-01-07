/* code generated by orbit */
package api

import (
	"net"
	"context"
	"sync"
	"github.com/desertbit/orbit/internal/packet"
	"errors"
	"time"
	"github.com/desertbit/orbit/pkg/orbit"
)

//##############//
//### Errors ###//
//##############//

const (
ErrCodeNotFound = 1
ErrCodeDatasetDoesNotExist = 2
)
var (
ErrNotFound = errors.New("not found")
orbitErrNotFound = orbit.Err(ErrNotFound, ErrNotFound.Error(), ErrCodeNotFound)
ErrDatasetDoesNotExist = errors.New("dataset does not exist")
orbitErrDatasetDoesNotExist = orbit.Err(ErrDatasetDoesNotExist, ErrDatasetDoesNotExist.Error(), ErrCodeDatasetDoesNotExist)
)
//#############//
//### Types ###//
//#############//

type Char struct {
lol string
}

type Plate struct {
name string
rect *Rect
test map[int]*Rect
test2 []*Rect
test3 []float32
test4 map[string]map[int][]*Rect
ts time.Time
version int
}

type Rect struct {
c *Char
x1 float32
x2 float32
y1 float32
y2 float32
}

type Test3Args struct {
c map[int][]*Rect
i int
v float64
}

type Test3Ret struct {
lol string
}

//################//
//### Services ###//
//################//

// Example  ---------------------
const (
ExampleTest = "ExampleTest"
ExampleTest2 = "ExampleTest2"
ExampleTest3 = "ExampleTest3"
ExampleTest4 = "ExampleTest4"
ExampleHello = "ExampleHello"
ExampleHello2 = "ExampleHello2"
ExampleHello3 = "ExampleHello3"
ExampleHello4 = "ExampleHello4"
)

type ExampleConsumerCaller interface {
// Calls
Test(ctx context.Context, args *Plate) (ret *Rect, err error)
Test2(ctx context.Context, args *Rect) (err error)
// Streams
Hello() (conn net.Conn, err error)
Hello2(args <-chan *Char) (err error)
}
type ExampleConsumerHandler interface {
// Calls
Test3(ctx context.Context, args *Test3Args) (ret *Test3Ret, err error)
Test4(ctx context.Context) (ret *Rect, err error)
// Streams
Hello3() (ret <-chan *Plate, err error)
Hello4(args <-chan *Char) (ret <-chan *Plate, err error)
}
type ExampleProviderCaller interface {
// Calls
Test3(ctx context.Context, args *Test3Args) (ret *Test3Ret, err error)
Test4(ctx context.Context) (ret *Rect, err error)
// Streams
Hello3() (ret <-chan *Plate, err error)
Hello4(args <-chan *Char) (ret <-chan *Plate, err error)
}
type ExampleProviderHandler interface {
// Calls
Test(ctx context.Context, args *Plate) (ret *Rect, err error)
Test2(ctx context.Context, args *Rect) (err error)
// Streams
Hello() (conn net.Conn, err error)
Hello2(args <-chan *Char) (err error)
}

type exampleConsumer struct {
h ExampleConsumerHandler
os *orbit.Session
}

func RegisterExampleConsumer(os *orbit.Session, h ExampleConsumerHandler) ExampleConsumerCaller {
cc := &exampleConsumer{h: h, os: os}
return cc
}
func (v1 *exampleConsumer) Test(ctx context.Context, args *Plate) (ret *Rect, err error) {
retData, err := v1.os.Call(ctx, ExampleTest, args)
if err != nil {
var cErr *orbit.ErrorCode
if errors.As(err, &cErr) {
switch cErr.Code {
case 1:
err = ErrNotFound
case 2:
err = ErrDatasetDoesNotExist
}
}
return
}
err = retData.Decode(ret)
if err != nil { return }
return
}

func (v1 *exampleConsumer) Test2(ctx context.Context, args *Rect) (err error) {
_, err = v1.os.Call(ctx, ExampleTest2, args)
if err != nil {
var cErr *orbit.ErrorCode
if errors.As(err, &cErr) {
switch cErr.Code {
case 1:
err = ErrNotFound
case 2:
err = ErrDatasetDoesNotExist
}
}
return
}
return
}

func (v1 *exampleConsumer) test3(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
var args *Test3Args
err = ad.Decode(args)
if err != nil { return }
ret, err := v1.h.Test3(ctx, args)
if err != nil {
if errors.Is(err, ErrNotFound) {
err = orbitErrNotFound
} else if errors.Is(err, ErrDatasetDoesNotExist) {
err = orbitErrDatasetDoesNotExist
}
return
}
r = ret
return
}

func (v1 *exampleConsumer) test4(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
ret, err := v1.h.Test4(ctx)
if err != nil {
if errors.Is(err, ErrNotFound) {
err = orbitErrNotFound
} else if errors.Is(err, ErrDatasetDoesNotExist) {
err = orbitErrDatasetDoesNotExist
}
return
}
r = ret
return
}

func (v1 *exampleConsumer) Hello() (conn net.Conn, err error) {
return v1.os.OpenStream(ExampleHello)
}
type ChanCharArgs struct {
C chan<- *Char
c chan *Char
close chan struct{}
mx sync.Mutex
err error
}

func newChanCharArgs(chanSize int) *ChanCharArgs {
c := &ChanCharArgs{c: make(chan *Char, chanSize), close: make(chan struct{})}
c.C = c.c
return c
}

func (c *ChanCharArgs) setError(err error)
c.mx.Lock()
c.err = err
c.mx.Unlock()
close(c.c)
}

func (c *ChanCharArgs) Close()
close(c.close)
}

func (c *ChanCharArgs) Err() (err error)
c.mx.Lock()
err = c.err
c.mx.Unlock()
return
}

func (v1 *exampleConsumer) Hello2(args <-chan *Char) (err error) {
conn, err := v1.os.OpenStream(ExampleHello2)
if err != nil { return }
go func() {
closingChan := v1.os.ClosingChan()
for {
select {
case <- closingChan:
return
case arg := <-args:
err := packet.WriteEncode(conn, arg, v1.os.Codec())
if err != nil && !v1.os.IsClosing() {
v1.os.Log().Error().Err(err).Str("channel", ExampleHello2).Msg("writing packet")
}
}
}
}()
return
}

type exampleProvider struct {
h ExampleProviderHandler
os *orbit.Session
}

func RegisterExampleProvider(os *orbit.Session, h ExampleProviderHandler) ExampleProviderCaller {
cc := &exampleProvider{h: h, os: os}
return cc
}
func (v1 *exampleProvider) Test3(ctx context.Context, args *Test3Args) (ret *Test3Ret, err error) {
retData, err := v1.os.Call(ctx, ExampleTest3, args)
if err != nil {
var cErr *orbit.ErrorCode
if errors.As(err, &cErr) {
switch cErr.Code {
case 1:
err = ErrNotFound
case 2:
err = ErrDatasetDoesNotExist
}
}
return
}
err = retData.Decode(ret)
if err != nil { return }
return
}

func (v1 *exampleProvider) Test4(ctx context.Context) (ret *Rect, err error) {
retData, err := v1.os.Call(ctx, ExampleTest4, nil)
if err != nil {
var cErr *orbit.ErrorCode
if errors.As(err, &cErr) {
switch cErr.Code {
case 1:
err = ErrNotFound
case 2:
err = ErrDatasetDoesNotExist
}
}
return
}
err = retData.Decode(ret)
if err != nil { return }
return
}

func (v1 *exampleProvider) test(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
var args *Plate
err = ad.Decode(args)
if err != nil { return }
ret, err := v1.h.Test(ctx, args)
if err != nil {
if errors.Is(err, ErrNotFound) {
err = orbitErrNotFound
} else if errors.Is(err, ErrDatasetDoesNotExist) {
err = orbitErrDatasetDoesNotExist
}
return
}
r = ret
return
}

func (v1 *exampleProvider) test2(ctx context.Context, s *orbit.Session, ad *orbit.Data) (r interface{}, err error) {
var args *Rect
err = ad.Decode(args)
if err != nil { return }
err = v1.h.Test2(ctx, args)
if err != nil {
if errors.Is(err, ErrNotFound) {
err = orbitErrNotFound
} else if errors.Is(err, ErrDatasetDoesNotExist) {
err = orbitErrDatasetDoesNotExist
}
return
}
return
}

type ChanPlateArgs struct {
C <-chan *Plate
c chan *Plate
close chan struct{}
mx sync.Mutex
err error
}

func newChanPlateArgs(chanSize int) *ChanPlateArgs {
c := &ChanPlateArgs{c: make(chan *Plate, chanSize), close: make(chan struct{})}
c.C = c.c
return c
}

func (c *ChanPlateArgs) setError(err error)
c.mx.Lock()
c.err = err
c.mx.Unlock()
close(c.c)
}

func (c *ChanPlateArgs) Close()
close(c.close)
}

func (c *ChanPlateArgs) Err() (err error)
c.mx.Lock()
err = c.err
c.mx.Unlock()
return
}

func (v1 *exampleProvider) Hello3() (ret <-chan *Plate, err error) {
conn, err := v1.os.OpenStream(ExampleHello3)
if err != nil { return }
retChan := make(chan *Plate, v1.os.StreamChanSize())
ret = retChan
go func() {
closingChan := v1.os.ClosingChan()
for {
select {
case <- closingChan:
return
default:
var data *Plate
err := packet.ReadDecode(conn, data, v1.os.Codec())
if err != nil && !v1.os.IsClosing() {
v1.os.Log().Error().Err(err).Str("channel", ExampleHello3).Msg("reading packet")
}
select {
case <-closingChan:
return
case retChan <- data:
}
}
}
}()
return
}

type ChanCharArgs struct {
C chan<- *Char
c chan *Char
close chan struct{}
mx sync.Mutex
err error
}

func newChanCharArgs(chanSize int) *ChanCharArgs {
c := &ChanCharArgs{c: make(chan *Char, chanSize), close: make(chan struct{})}
c.C = c.c
return c
}

func (c *ChanCharArgs) setError(err error)
c.mx.Lock()
c.err = err
c.mx.Unlock()
close(c.c)
}

func (c *ChanCharArgs) Close()
close(c.close)
}

func (c *ChanCharArgs) Err() (err error)
c.mx.Lock()
err = c.err
c.mx.Unlock()
return
}

type ChanPlateArgs struct {
C <-chan *Plate
c chan *Plate
close chan struct{}
mx sync.Mutex
err error
}

func newChanPlateArgs(chanSize int) *ChanPlateArgs {
c := &ChanPlateArgs{c: make(chan *Plate, chanSize), close: make(chan struct{})}
c.C = c.c
return c
}

func (c *ChanPlateArgs) setError(err error)
c.mx.Lock()
c.err = err
c.mx.Unlock()
close(c.c)
}

func (c *ChanPlateArgs) Close()
close(c.close)
}

func (c *ChanPlateArgs) Err() (err error)
c.mx.Lock()
err = c.err
c.mx.Unlock()
return
}

func (v1 *exampleProvider) Hello4(args <-chan *Char) (ret <-chan *Plate, err error) {
conn, err := v1.os.OpenStream(ExampleHello4)
if err != nil { return }
go func() {
closingChan := v1.os.ClosingChan()
for {
select {
case <- closingChan:
return
case arg := <-args:
err := packet.WriteEncode(conn, arg, v1.os.Codec())
if err != nil && !v1.os.IsClosing() {
v1.os.Log().Error().Err(err).Str("channel", ExampleHello4).Msg("writing packet")
}
}
}
}()
retChan := make(chan *Plate, v1.os.StreamChanSize())
ret = retChan
go func() {
closingChan := v1.os.ClosingChan()
for {
select {
case <- closingChan:
return
default:
var data *Plate
err := packet.ReadDecode(conn, data, v1.os.Codec())
if err != nil && !v1.os.IsClosing() {
v1.os.Log().Error().Err(err).Str("channel", ExampleHello4).Msg("reading packet")
}
select {
case <-closingChan:
return
case retChan <- data:
}
}
}
}()
return
}

// ---------------------



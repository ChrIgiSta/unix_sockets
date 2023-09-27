![go-test](https://github.com/ChrIgiSta/unix_sockets/actions/workflows/tests.yml/badge.svg)

# GoLang Unix Sockets
This package contains easy to use unix sockets.

## Test
`go test -v ./...`

## Build Echo Unix Socket Server
`go build cmd/server/echo.go`

## Usage

### Unix Socket Server
```go
package main

import (
	"log"
	"sync"

	"github.com/ChrIgiSta/unix_sockets"
)

func main() {
	var wg sync.WaitGroup = sync.WaitGroup{}
	defer wg.Wait()

	s := unix_sockets.NewUnixSocketServer("/tmp/mySocket.sock") // create a socket server
	wg.Add(1)
	rxCh, err := s.ListenReceive(&wg) // start listen
	if err != nil {
		log.Fatal(err)
	}
	defer s.Shutdown()

	msg, ok := <-rxCh // recieve some data
	if !ok {
		return
	}

	s.SendAll(msg) // send some data
}
```

### Unix Socket Client
```go
package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/ChrIgiSta/unix_sockets"
)

func main() {
	var wg sync.WaitGroup = sync.WaitGroup{}
	defer wg.Wait()

	c := unix_sockets.NewUnixSocketClient()
	wg.Add(1)
	rxCh, err := c.Connect(&wg, "/tmp/mySocket.sock") // create a socket client
	if err != nil {
		log.Fatal(err)
	}
	defer c.Disconnect()

	err = c.Send([]byte("Hello World!")) // send some data
	if err != nil {
		log.Fatal(err)
	}

	msg := <-rxCh // recieve some data
	fmt.Println(msg)
}
```
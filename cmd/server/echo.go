package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/ChrIgiSta/unix_sockets"
)

const (
	SOCKET_PATH = "/tmp/myunixsock.sock"
)

func main() {
	var wg sync.WaitGroup = sync.WaitGroup{}
	defer wg.Wait()

	s := unix_sockets.NewUnixSocketServer(SOCKET_PATH)
	wg.Add(1)
	rxCh, err := s.ListenReceive(&wg)
	if err != nil {
		log.Fatal(err)
	}

	for {
		msg, ok := <-rxCh
		if !ok {
			return
		}
		fmt.Println(msg)
		fmt.Println(string(msg))
		s.SendAll(msg)

		if string(msg) == "exit" {
			break
		}
	}

	err = s.Shutdown()
	if err != nil {
		log.Println(err)
	}
}

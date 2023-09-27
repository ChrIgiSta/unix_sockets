/**
 * Copyright © 2023, Staufi Tech - Switzerland
 * All rights reserved.
 *
 *  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 *  AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 *  IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 *  ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 *  LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 *  CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 *  SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 *  INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 *  CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 *  ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 *  POSSIBILITY OF SUCH DAMAGE.
 */

package unix_sockets

import (
	"log"
	"net"
	"os"
	"sync"

	"github.com/ChrIgiSta/unix_sockets/utils"
)

const (
	UNIX_SOCKET_NETWORK_STRING = "unix"
	UNIX_SOCKET_BUFFER_SIZE    = 2048
	UNIX_SOCKET_MAX_CLIENTS    = 10
	NEW_LINE_BYTE              = 0x0a
)

type UnixSocketServer struct {
	listener    net.Listener
	interrupted bool
	clientPool  *utils.ClientPool
	socketPath  string

	listenerErr error
	clientErr   error
}

func NewUnixSocketServer(socketPath string) *UnixSocketServer {

	var (
		unixSock UnixSocketServer = UnixSocketServer{}
	)

	os.Remove(socketPath)

	unixSock.interrupted = false
	unixSock.clientErr = nil
	unixSock.listenerErr = nil
	unixSock.socketPath = socketPath
	unixSock.clientPool = utils.NewClientPool(UNIX_SOCKET_MAX_CLIENTS)

	return &unixSock
}

func (s *UnixSocketServer) GetPath() string {

	return s.socketPath
}

func (s *UnixSocketServer) IsInterrupted() bool {
	return s.interrupted
}

func (s *UnixSocketServer) ClientErr() error {
	return s.clientErr
}

func (s *UnixSocketServer) ListenerErr() error {
	return s.listenerErr
}

func (s *UnixSocketServer) SendAll(message []byte) []error {
	var errors []error

	message = AppendNewLineIfNecessary(message)

	for _, c := range s.clientPool.GetAll() {
		_, err := c.Write(message)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (s *UnixSocketServer) ListenReceive(wg *sync.WaitGroup) (<-chan []byte, error) {
	var err error

	s.interrupted = false
	s.listenerErr = nil
	s.clientErr = nil

	rxChannel := make(chan []byte, UNIX_SOCKET_BUFFER_SIZE)

	s.listener, err = net.Listen(UNIX_SOCKET_NETWORK_STRING, s.socketPath)
	if err != nil {
		defer close(rxChannel)
		return rxChannel, err
	}

	go s.waitForClient(wg, rxChannel)

	return rxChannel, err
}

func (s *UnixSocketServer) Shutdown() error {
	s.interrupted = true
	return s.listener.Close()
}

func (s *UnixSocketServer) waitForClient(wg *sync.WaitGroup, rxChannel chan<- []byte) {
	defer wg.Done()

	var wgClients sync.WaitGroup = sync.WaitGroup{}

	for !s.interrupted {
		client, err := s.listener.Accept()
		if err != nil {
			defer close(rxChannel)
			s.interrupted = true
			log.Println("error: client accept, ", err)
			s.listenerErr = err
			wgClients.Wait()
			return
		} else {
			wgClients.Add(1)
			go s.clientHandler(client, rxChannel, &wgClients)
		}
	}
}

func (s *UnixSocketServer) clientHandler(client net.Conn, rxChannel chan<- []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := s.clientPool.Add(client); err != nil {
		log.Println("error: client add to pool, ", err)
		return
	}
	defer func() {
		if err := s.clientPool.Drop(client); err != nil {
			log.Println("error: client drop from pool, ", err)
		}
	}()

	buffer := make([]byte, UNIX_SOCKET_BUFFER_SIZE)

	log.Println("info: client from ", client.RemoteAddr().String())
	for !s.interrupted {
		n, err := client.Read(buffer)
		if err != nil {
			s.clientErr = err
			log.Println("error: client read, ", err)
			return
		}

		rxChannel <- RemoveNewLineIfNecessary(buffer[:n])
	}
}

func AppendNewLineIfNecessary(in []byte) []byte {
	if in[len(in)-1] != NEW_LINE_BYTE {
		in = append(in, NEW_LINE_BYTE)
	}
	return in
}

func RemoveNewLineIfNecessary(in []byte) []byte {
	if in[len(in)-1] == NEW_LINE_BYTE {
		return in[:len(in)-1]
	}
	return in
}

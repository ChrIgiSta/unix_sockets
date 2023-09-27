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
	"errors"
	"log"
	"net"
	"sync"
)

type UnixSocketClient struct {
	connection  net.Conn
	interrupted bool
	buffer      []byte
}

func NewUnixSocketClient() *UnixSocketClient {
	return &UnixSocketClient{
		interrupted: false,
		buffer:      make([]byte, UNIX_SOCKET_BUFFER_SIZE),
	}
}

func (c *UnixSocketClient) Connect(wg *sync.WaitGroup, socketPath string) (<-chan []byte, error) {
	var err error

	c.interrupted = false
	c.connection, err = net.Dial(UNIX_SOCKET_NETWORK_STRING, socketPath)
	if err != nil {
		c.interrupted = true
		return nil, err
	}

	rxChannel := make(chan []byte, UNIX_SOCKET_BUFFER_SIZE)
	go c.readerLoop(wg, rxChannel)

	return rxChannel, err
}

func (c *UnixSocketClient) Disconnect() error {
	c.interrupted = true
	return c.connection.Close()
}

func (c *UnixSocketClient) Send(message []byte) error {
	var err error

	if !c.interrupted {
		_, err = c.connection.Write(AppendNewLineIfNecessary(message))
	} else {
		err = errors.New("not connected")
	}
	return err
}

func (c *UnixSocketClient) readerLoop(wg *sync.WaitGroup, rxChannel chan<- []byte) {

	defer wg.Done()
	defer close(rxChannel)

	for !c.interrupted {
		n, err := c.connection.Read(c.buffer)
		if err != nil {
			c.interrupted = true
			log.Println("error: client reader loop exited, ", err)
			return
		}
		rxChannel <- RemoveNewLineIfNecessary(c.buffer[:n])
	}
}

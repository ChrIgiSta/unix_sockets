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
	"sync"
	"testing"
	"time"
)

const TEST_SOCKETS_SOCKET_PATH = "/tmp/socket-test.sock"

func TestSockets_TestHelperFunctions(t *testing.T) {

	test := "hello\n"
	out := AppendNewLineIfNecessary([]byte(test))
	if string(out) != test {
		t.Error("Append new line would not be necessary")
	}
	test = "hello"
	out = AppendNewLineIfNecessary([]byte(test))
	if string(out) != test+"\n" {
		t.Error("Append new line would be necessary")
	}

	test = "hello\n"
	out = RemoveNewLineIfNecessary([]byte(test))
	if string(out)+"\n" != test {
		t.Error("Remove new line would be necessary")
	}
	test = "hello"
	out = RemoveNewLineIfNecessary([]byte(test))
	if string(out) != test {
		t.Error("Some unexpected changes")
	}
}

func TestSockets_ServerInit(t *testing.T) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	unixSockServer := NewUnixSocketServer(TEST_SOCKETS_SOCKET_PATH)
	wg.Add(1)
	_, err := unixSockServer.ListenReceive(&wg)
	if err != nil {
		t.Error("init unix socket server, ", err)
	}

	err = unixSockServer.Shutdown()
	if err != nil {
		t.Error("stop unix socket server, ", err)
	}
}

func TestSockets_ClientConnectToServer(t *testing.T) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	unixSockServer := NewUnixSocketServer(TEST_SOCKETS_SOCKET_PATH)
	wg.Add(1)
	rxServerCh, err := unixSockServer.ListenReceive(&wg)
	if err != nil {
		t.Error("listen unix socket server, ", err)
	}
	t.Log("server socket created")

	unixSockClient := NewUnixSocketClient()
	wg.Add(1)
	rxClientCh, err := unixSockClient.Connect(&wg, TEST_SOCKETS_SOCKET_PATH)
	if err != nil {
		t.Error("connect unix socket client, ", err)
	}
	t.Log("client socket created and connected")
	time.Sleep(1 * time.Second)

	errs := unixSockServer.SendAll([]byte("Hello Client!"))
	for _, err := range errs {
		t.Error("server send all err, ", err)
	}

	rxAtClient := <-rxClientCh
	if string(rxAtClient) != "Hello Client!" {
		t.Error("unexpected rx at client")
		t.Error(string(rxAtClient))
	}

	t.Log("sent server to client")

	err = unixSockClient.Send([]byte("Hello Server!"))
	if err != nil {
		t.Error("client send, ", err)
	}

	rxAtServer := <-rxServerCh
	if string(rxAtServer) != "Hello Server!" {
		t.Error("unexpected rx at server")
		t.Error(string(rxAtServer))
	}

	t.Log("sent client to server")

	err = unixSockClient.Disconnect()
	if err != nil {
		t.Error("disconnecting client, ", err)
	}

	err = unixSockServer.Shutdown()
	if err != nil {
		t.Error("shutdown server, ", err)
	}
}

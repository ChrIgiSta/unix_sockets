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

package utils

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/go-test/deep"
)

func TestSockets_ClientPool(t *testing.T) {
	cPool := NewClientPool(10)

	client := NewFakeUnixConn("localhost", "/temp/xy.sock")

	err := cPool.Add(client)
	if err != nil {
		t.Error(err)
	}

	gottenConn, err := cPool.Get("/temp/xy.sock")
	if err != nil {
		t.Error("pool get single client, ", err)
	}
	if diff := deep.Equal(*gottenConn, client); diff != nil {
		t.Error("get single client from pool, ", diff)
	}

	allConns := cPool.GetAll()
	if len(allConns) != 1 {
		t.Error("pool size mismatch")
	}

	if diff := deep.Equal(allConns[0], client); diff != nil {
		t.Error("get all client from pool, ", diff)
	}

	err = cPool.Drop(client)
	if err != nil {
		t.Error("drop client, ", err)
	}

	allConns = cPool.GetAll()
	if len(allConns) != 0 {
		t.Error("not an empty pool after drop")
	}
}

func TestSockets_ClientPoolCheckLimit(t *testing.T) {
	var limit uint = 11

	cPool := NewClientPool(limit)

	for i := 0; i < int(limit); i++ {
		client := NewFakeUnixConn("localhost", strconv.Itoa(i))
		err := cPool.Add(client)
		if err != nil {
			t.Error("add client to pool, ", err)
		}
	}
	client := NewFakeUnixConn("localhost", strconv.Itoa(int(limit)))
	err := cPool.Add(client)
	if err == nil {
		t.Error("pool not limited")
	}

	for i := 0; i < int(limit); i += 2 {
		client := NewFakeUnixConn("localhost", strconv.Itoa(i))
		err := cPool.Drop(client)
		if err != nil {
			t.Error("drop some client from pool, ", err)
		}
	}

	allClients := cPool.GetAll()
	if len(allClients) != int(limit/2) {
		t.Error("size mismatch")
	}
}

type FakeUnixConn struct {
	localAddress  string
	remoteAddress string
}

func NewFakeUnixConn(localAddress string, remoteAddress string) *FakeUnixConn {
	return &FakeUnixConn{
		localAddress:  localAddress,
		remoteAddress: remoteAddress,
	}
}

func (f *FakeUnixConn) Read(b []byte) (n int, err error) {
	return len(b), nil
}
func (f *FakeUnixConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}
func (f *FakeUnixConn) Close() error {
	return nil
}
func (f *FakeUnixConn) LocalAddr() net.Addr {
	return &net.UnixAddr{
		Name: f.localAddress,
		Net:  f.localAddress,
	}
}
func (f *FakeUnixConn) RemoteAddr() net.Addr {
	return &net.UnixAddr{
		Name: f.remoteAddress,
		Net:  f.remoteAddress,
	}
}
func (f *FakeUnixConn) SetDeadline(t time.Time) error {
	return nil
}
func (f *FakeUnixConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (f *FakeUnixConn) SetWriteDeadline(t time.Time) error {
	return nil
}

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
	"errors"
	"net"
	"sync"
)

const CLIENT_POOL_SIZE_UNLIMITED = 0

type ClientPool struct {
	clients []net.Conn
	access  sync.Mutex
	maxSize uint
}

func NewClientPool(maxSize uint) *ClientPool {
	return &ClientPool{
		clients: make([]net.Conn, 0),
		access:  sync.Mutex{},
		maxSize: maxSize,
	}
}

func (p *ClientPool) Add(client net.Conn) error {
	p.access.Lock()
	defer p.access.Unlock()

	if p.maxSize != CLIENT_POOL_SIZE_UNLIMITED &&
		len(p.clients) >= int(p.maxSize) {
		return errors.New("max size reached")
	}

	for _, c := range p.clients {
		if c.RemoteAddr().String() == client.RemoteAddr().String() {
			return errors.New("client already exist")
		}
	}

	p.clients = append(p.clients, client)

	return nil
}

func (p *ClientPool) Drop(client net.Conn) error {
	p.access.Lock()
	defer p.access.Unlock()

	var clientToDelete int = -1

	for i, c := range p.clients {
		if c.RemoteAddr().String() == client.RemoteAddr().String() {
			clientToDelete = i
			break
		}
	}

	if clientToDelete < 0 {
		return errors.New("client to delete not found")
	}

	// not order sensitiv
	p.clients[clientToDelete] = p.clients[len(p.clients)-1]
	p.clients[len(p.clients)-1] = nil
	p.clients = p.clients[:len(p.clients)-1]

	return nil
}

func (p *ClientPool) GetAll() []net.Conn {
	return p.clients
}

func (p *ClientPool) Get(address string) (*net.Conn, error) {
	p.access.Lock()
	defer p.access.Unlock()

	for _, c := range p.clients {
		if c.RemoteAddr().String() == address {
			return &c, nil
		}
	}
	return nil, errors.New("client not found")
}

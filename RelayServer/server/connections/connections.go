package connections

import (
	"net"
	"net/netip"
	"sync"

	"github.com/google/uuid"
)

type RelayAddr struct {
	ExpectedIP    netip.Addr
	UDPReturnAddr *net.UDPAddr
}

// Thread safe connection tracker for active users
type RelayConnections struct {
	connections map[uuid.UUID]RelayAddr
	userIPs     map[netip.Addr]uuid.UUID
	mtx         sync.RWMutex
}

func (rc *RelayConnections) Set(user uuid.UUID, conn RelayAddr) {
	rc.mtx.Lock()
	defer rc.mtx.Unlock()
	rc.connections[user] = conn
	rc.userIPs[conn.ExpectedIP] = user
}

func (rc *RelayConnections) Remove(user uuid.UUID) {
	rc.mtx.Lock()
	defer rc.mtx.Unlock()
	addr, exists := rc.connections[user]
	delete(rc.connections, user)
	if exists {
		delete(rc.userIPs, addr.ExpectedIP)
	}
}

func (rc *RelayConnections) Get(user uuid.UUID) (RelayAddr, bool) {
	rc.mtx.RLock()
	defer rc.mtx.RUnlock()
	addr, exists := rc.connections[user]
	return addr, exists
}

func (rc *RelayConnections) UserFromAddress(addr netip.Addr) (uuid.UUID, bool) {
	rc.mtx.RLock()
	defer rc.mtx.RUnlock()
	user, exists := rc.userIPs[addr]
	return user, exists
}

func NewConnectionsMap() *RelayConnections {
	return &RelayConnections{
		connections: make(map[uuid.UUID]RelayAddr),
	}
}

package router

import (
	"RIMB/relayserver/server/connections"
	"RIMB/relayserver/server/peers"
	"encoding/binary"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

type UDPPacket struct {
	data []byte
	addr *net.UDPAddr
}

// Routes packets for peer groups
type Router struct {
	udpSocket        *net.UDPConn
	peerGroupManager *peers.PeerGroupManager
	connections      *connections.RelayConnections
	udpBuffer        chan UDPPacket
	closeSignal      chan struct{}
	wg               sync.WaitGroup
}

func (r *Router) Close() {
	close(r.closeSignal)
	r.udpSocket.Close()
	r.wg.Wait()
}

func (r *Router) run(numWorkers int) {
	defer r.wg.Done()
	defer close(r.udpBuffer)
	for range numWorkers {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			for packet := range r.udpBuffer {
				if len(packet.data) == 16 {
					r.handleHeartbeat(packet)
				} else if len(packet.data) > 20 {
					r.handlePacket(packet)
				}
			}
		}()
	}
	for {
		select {
		case <-r.closeSignal:
			return
		default:
			r.udpSocket.SetReadDeadline(time.Now().Add(2 * time.Second))
			buffer := make([]byte, 2048)
			n, remoteAddr, err := r.udpSocket.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Println("Listener failed:", err)
				return
			}
			select {
			case r.udpBuffer <- UDPPacket{data: buffer[:n], addr: remoteAddr}:
			default:
				log.Println("UDP buffer full: dropping packet")
			}
		}
	}
}

func (r *Router) handlePacket(p UDPPacket) {
	peerID, err := uuid.FromBytes(p.data[:16])
	if err != nil {
		log.Printf("Error parsing UUID from %v: %v", p.addr.IP, err)
		return
	}
	if !r.peerGroupManager.HasPeer(peerID) {
		return // Peer does not exist
	}
	userID := r.peerGroupManager.PeerToUser(peerID)
	if userID == uuid.Nil {
		return
	}
	expectedAddr, exists := r.connections.Get(userID)
	if !exists || !expectedAddr.ExpectedIP.Compare() == p.addr.IP {
		return
	}
	clientID := binary.BigEndian.Uint32(p.data[16:20])
	destinations, err := r.peerGroupManager.ResolveDestinationUUIDs(peerID, clientID)
	if err != nil {
		log.Printf("Error resolving destinations for %v: %v", userID, err)
		return
	}
	payload := p.data[20:]
	for _, user := range destinations {
		dst, exists := r.connections.Get(user)
		if !exists || dst.UDPReturnAddr == nil {
			continue
		}
		r.udpSocket.WriteToUDP(payload, dst.UDPReturnAddr)
	}
	if expectedAddr.UDPReturnAddr == nil || expectedAddr.UDPReturnAddr.Port != p.addr.Port {
		expectedAddr.UDPReturnAddr = &net.UDPAddr{
			IP:   append(net.IP(nil), p.addr.IP...),
			Port: p.addr.Port,
			Zone: p.addr.Zone,
		}
		r.connections.Set(userID, expectedAddr)
	}
}

func (r *Router) handleHeartbeat(p UDPPacket) {
	peerID, err := uuid.FromBytes(p.data)
	if err != nil {
		log.Printf("Error parsing UUID from %v: %v", p.addr.IP, err)
		return
	}
	if !r.peerGroupManager.HasPeer(peerID) {
		return // Peer does not exist
	}
	userID := r.peerGroupManager.PeerToUser(peerID)
	if userID == uuid.Nil {
		return
	}
	expectedAddr, exists := r.connections.Get(userID)
	if !exists || !expectedAddr.ExpectedIP.Equal(p.addr.IP) {
		return
	}
	if expectedAddr.UDPReturnAddr == nil || expectedAddr.UDPReturnAddr.Port != p.addr.Port {
		expectedAddr.UDPReturnAddr = &net.UDPAddr{
			IP:   append(net.IP(nil), p.addr.IP...),
			Port: p.addr.Port,
			Zone: p.addr.Zone,
		}
		r.connections.Set(userID, expectedAddr)
	}
}

func StartRouter(address *net.UDPAddr, pgm *peers.PeerGroupManager, connections *connections.RelayConnections, bufferSize int, numWorkers int) (*Router, error) {
	conn, err := net.ListenUDP("udp", address)
	if err != nil {
		return nil, err
	}
	r := &Router{
		udpSocket:        conn,
		peerGroupManager: pgm,
		connections:      connections,
		udpBuffer:        make(chan UDPPacket, bufferSize),
		closeSignal:      make(chan struct{}),
	}
	r.wg.Add(1)
	go r.run(numWorkers)
	return r, nil
}

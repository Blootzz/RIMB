package server

import (
	"RIMB/relayserver/server/connections"
	"RIMB/relayserver/server/control"
	"RIMB/relayserver/server/peers"
	"RIMB/relayserver/server/router"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/google/uuid"
)

type TCPPacket struct {
	data []byte
	addr *net.TCPAddr
}

type RelayServer struct {
	router  *router.Router
	control *control.Control
}

func (s *RelayServer) Close() error {
	s.router.Close()
	s.contol.Close()
	return nil
}

func (s *RelayServer) Wait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Termination signal received. Shutting down server...")
	s.Close()
}

func StartRelay(address string) (*RelayServer, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	pgm := peers.NewPeerGroupManager(256, func(u uuid.UUID) error { return nil })
	connections := connections.NewConnectionsMap()
	router, err := router.StartRouter(udpAddr, pgm, connections, 2000, runtime.NumCPU())
	if err != nil {
		return nil, err
	}
	s := &RelayServer{
		router:    router,
		tcpSocket: nil,
	}
	log.Printf("Relay server started on %s", address)
	return s, nil
}

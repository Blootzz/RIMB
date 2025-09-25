package control

import (
	"RIMB/relayserver/server/connections"
	"RIMB/relayserver/server/peers"
	"bufio"
	"log"
	"net"
	"sync"
	"time"
)

type TCPPacket struct {
	data []byte
	addr *net.TCPAddr
}

// Control server used to establish identity and manage peer groups
type Control struct {
	tcpListener       *net.TCPListener
	peerGroupManager  *peers.PeerGroupManager
	connections       *connections.RelayConnections
	connectionTimeout time.Duration // Time to reclaim a user UUID after disconnect before it becomes stale
	tcpBuffer         chan TCPPacket
	closeSignal       chan struct{}
	wg                sync.WaitGroup
}

func (c *Control) run(numWorkers int) {
	defer c.wg.Done()
	defer close(c.tcpBuffer)
	for range numWorkers {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			for packet := range c.tcpBuffer {
				c.handlePacket(packet)
			}
		}()
	}
	for {
		select {
		case <-c.closeSignal:
			return
		default:
			c.tcpListener.SetDeadline(time.Now().Add(2 * time.Second))
			conn, err := c.tcpListener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Println("Listener failed:", err)
				return
			}
			c.wg.Add(1)
			go func(conn net.Conn) {
				defer c.wg.Done()
				defer conn.Close()
				addr := conn.RemoteAddr().(*net.TCPAddr)
				scanner := bufio.NewScanner(conn)
				buf := make([]byte, 0, 256)
				scanner.Buffer(buf, 256)
				for scanner.Scan() {
					msg := append([]byte(nil), scanner.Bytes()...)
					packet := TCPPacket{
						data: msg,
						addr: addr,
					}
					select {
					case c.tcpBuffer <- packet:
					case <-c.closeSignal:
						return
					}
				}
				if err := scanner.Err(); err != nil {
					if err == bufio.ErrTooLong {
						conn.Write([]byte("ERROR: message too large"))
					} else {
						log.Printf("Read error from %v: %v", conn.RemoteAddr(), err)
					}
				}
				
			}(conn)
		}
	}
}

func (c *Control) handlePacket(p TCPPacket) {

}

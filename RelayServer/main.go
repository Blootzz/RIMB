package main

import (
	"RIMB/relayserver/server/peers"
	server "RIMB/relayserver/server/relay"
	"fmt"
	"os"

	"github.com/google/uuid"
	"gopkg.in/ini.v1"
)

func main() {

	inidata, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	section := inidata.Section("Server")
	ip := section.Key("IPAddress").String()
	port, err := section.Key("Port").Uint()
	if err != nil {
		fmt.Printf("Failed to parse field in config file: %v", err)
		os.Exit(1)
	}
	if len(ip) == 0 {
		fmt.Printf("IP address not set")
		os.Exit(1)
	}
	relay, err := server.StartRelay(fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Printf("Failed to start relay: %v", err)
		os.Exit(1)
	}
	relay.Wait()

	for range 100 {
		pgm := peers.NewPeerGroupManager(256, func(user uuid.UUID) error {
			print("piss")
			return nil
		})
		user, err := uuid.NewRandom()
		if err != nil {
			return
		}
		user2, _ := uuid.NewRandom()
		groupID, err := pgm.CreateGroup(user, false)
		if err != nil {
			return
		}
		groupID2, err := pgm.CreateGroup(user2, true)
		if err != nil {
			return
		}
		println(groupID2.String())
		println(pgm.NumGroups())
		pgm.AssignUser(user2, groupID)
		println(pgm.NumGroups())

	}
}

package peers

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/google/uuid"
)

type Peer struct {
	UserUUID uuid.UUID
	PeerUUID uuid.UUID
	ClientID uint32
}

type PeerGroup struct {
	groupID          uuid.UUID
	userPeerMap      map[uuid.UUID]uuid.UUID // Bidirectional map setup for peer ID
	peerUserMap      map[uuid.UUID]uuid.UUID
	peerClientIDMap  map[uuid.UUID]uint32 // Ditto with peer <-> clientID
	clientIDPeerMap  map[uint32]uuid.UUID
	owner            uint32
	cidCounter       uint32
	canReassignOwner bool
}

// Registers a new user UUID to the peer group. Immediately returns true if the user is already registered.
func (pg *PeerGroup) RegisterUser(user uuid.UUID) (bool, error) {
	if _, exists := pg.userPeerMap[user]; exists {
		return true, nil
	}
	peerID, err := uuid.NewRandom()
	if err != nil {
		return false, fmt.Errorf("failed to generate new peer id for user %s: %w", user, err)
	}
	cid := pg.cidCounter
	pg.userPeerMap[user] = peerID
	pg.peerUserMap[peerID] = user
	pg.peerClientIDMap[peerID] = cid
	pg.clientIDPeerMap[cid] = peerID
	pg.cidCounter++
	return true, nil
}

// Removes the provided user UUID from this peer group.
// If the user is an owner, the next user with the lowest client ID will be promoted as the new owner.
// If the canReassignOwner flag is false and the user is an owner, the peer group will instead enter an invalid state.
func (pg *PeerGroup) RemoveUser(user uuid.UUID) bool {
	peerID, exists := pg.userPeerMap[user]
	if !exists {
		return false
	}
	clientID := pg.peerClientIDMap[peerID]
	delete(pg.userPeerMap, user)
	delete(pg.peerUserMap, peerID)
	delete(pg.peerClientIDMap, peerID)
	delete(pg.clientIDPeerMap, clientID)
	if pg.IsEmpty() || clientID != pg.owner || !pg.canReassignOwner {
		return true
	}
	clients := pg.GetClients()
	pg.owner = clients[0]
	for i := 1; i < len(clients); i++ {
		if pg.owner > clients[i] {
			pg.owner = clients[i]
		}
	}
	return true
}

// Checks if the provided peer UUID is associated with the owner client ID.
func (pg *PeerGroup) IsOwner(peer uuid.UUID) bool {
	clientID, exists := pg.peerClientIDMap[peer]
	if !exists {
		return false
	}
	return clientID == pg.owner
}

// Returns the destination user UUID, given that both the sending peer UUID and destination client ID are valid.
func (pg *PeerGroup) ResolveDestinationUUID(peer uuid.UUID, clientID uint32) (uuid.UUID, error) {
	srcCID, exists := pg.peerClientIDMap[peer]
	if !exists {
		return uuid.Nil, errors.New("invalid peer UUID")
	}
	if clientID == 0 {
		clientID = pg.owner
	}
	if srcCID == clientID {
		return uuid.Nil, errors.New("source and destination are the same")
	}
	dstPeer, exists := pg.clientIDPeerMap[clientID]
	if !exists {
		return uuid.Nil, errors.New("invalid client ID")
	}

	return pg.peerUserMap[dstPeer], nil
}

// When this returns false, the peer group has entered a state where it should be destroyed.
// A peer group is invalid when it has no members or the owner is not in the client list.
func (pg *PeerGroup) IsValid() bool {
	if pg.IsEmpty() {
		return false
	}
	_, exists := pg.clientIDPeerMap[pg.owner]
	return exists
}

func (pg *PeerGroup) GetClients() []uint32 {
	return slices.Collect(maps.Keys(pg.clientIDPeerMap))
}

func (pg *PeerGroup) NumUsers() int {
	return len(pg.userPeerMap)
}

func (pg *PeerGroup) IsEmpty() bool {
	return pg.NumUsers() <= 0
}

func (pg *PeerGroup) GetRegisteredUsers() []uuid.UUID {
	return slices.Collect(maps.Keys(pg.userPeerMap))
}

func (pg *PeerGroup) HasUser(user uuid.UUID) bool {
	_, exists := pg.userPeerMap[user]
	return exists
}

func (pg *PeerGroup) PeerLookup(user uuid.UUID) *Peer {
	peerUUID, exists := pg.userPeerMap[user]
	if !exists {
		return nil
	}
	cid := pg.peerClientIDMap[peerUUID]
	return &Peer{
		UserUUID: user,
		PeerUUID: peerUUID,
		ClientID: cid,
	}
}

func NewPeerGroup(owner uuid.UUID, canReassignOwner bool) (*PeerGroup, error) {
	pgID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new id for peer group %s: %w", pgID, err)
	}
	pg := PeerGroup{
		groupID:          pgID,
		userPeerMap:      make(map[uuid.UUID]uuid.UUID),
		peerUserMap:      make(map[uuid.UUID]uuid.UUID),
		peerClientIDMap:  make(map[uuid.UUID]uint32),
		clientIDPeerMap:  make(map[uint32]uuid.UUID),
		canReassignOwner: canReassignOwner,
	}
	_, err = pg.RegisterUser(owner)
	if err != nil {
		return nil, fmt.Errorf("failed to register owner for new peer group %s: %w", pgID, err)
	}
	return &pg, nil
}

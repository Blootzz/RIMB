package peers

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/google/uuid"
)

type NotifyOnDestroyCallback func(uuid.UUID) error

// Unlike PeerGroup, PGM is designed to be thread safe.
type PeerGroupManager struct {
	userPeerGroupMap     map[uuid.UUID]uuid.UUID
	peerToGroupMap       map[uuid.UUID]uuid.UUID
	peerGroupInstanceMap map[uuid.UUID]*PeerGroup
	notifyCallback       NotifyOnDestroyCallback
	peerGroupSizeLimit   uint16
	mtx                  sync.RWMutex
}

func (pgm *PeerGroupManager) CreateGroup(owner uuid.UUID, canReassignOwner bool) (uuid.UUID, error) {
	pg, err := NewPeerGroup(owner, canReassignOwner)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create peer group: %w", err)
	}
	pgm.mtx.Lock()
	notify := pgm.remove(owner)
	pgm.peerGroupInstanceMap[pg.groupID] = pg
	pgm.userPeerGroupMap[owner] = pg.groupID
	if p := pg.PeerLookup(owner); p != nil {
		pgm.peerToGroupMap[p.PeerUUID] = pg.groupID
	}
	pgm.mtx.Unlock()
	pgm.notifyUsers(notify)
	return pg.groupID, nil
}

func (pgm *PeerGroupManager) remove(user uuid.UUID) []uuid.UUID {
	pgUUID, exists := pgm.userPeerGroupMap[user]
	if !exists {
		return nil
	}
	pg := pgm.peerGroupInstanceMap[pgUUID]
	if p := pg.PeerLookup(user); p != nil {
		delete(pgm.peerToGroupMap, p.PeerUUID)
	}
	pg.RemoveUser(user)
	notifyList := make([]uuid.UUID, 0, pg.NumUsers())
	if !pg.IsValid() {
		for _, peer := range pg.GetRegisteredUsers() {
			delete(pgm.userPeerGroupMap, peer)
			if p := pg.PeerLookup(peer); p != nil {
				delete(pgm.peerToGroupMap, p.PeerUUID)
				notifyList = append(notifyList, peer)
			}
		}
		delete(pgm.peerGroupInstanceMap, pgUUID)
	}
	delete(pgm.userPeerGroupMap, user)
	return notifyList
}

func (pgm *PeerGroupManager) notifyUsers(notifyList []uuid.UUID) {
	if notifyList == nil {
		return
	}
	for _, peer := range notifyList {
		if err := pgm.notifyCallback(peer); err != nil {
			log.Printf("Failed to notify user %s during cleanup of peer group: %v", peer, err)
		}
	}
}

func (pgm *PeerGroupManager) RemoveUser(user uuid.UUID) bool {
	pgm.mtx.Lock()
	notify := pgm.remove(user)
	pgm.mtx.Unlock()
	pgm.notifyUsers(notify)
	return notify != nil
}

func (pgm *PeerGroupManager) AssignUser(user uuid.UUID, group uuid.UUID) (bool, error) {
	pgm.mtx.Lock()
	pg, exists := pgm.peerGroupInstanceMap[group]
	if !exists {
		pgm.mtx.Unlock()
		return false, nil
	}
	if pg.HasUser(user) {
		pgm.mtx.Unlock()
		return true, nil
	}
	if pg.NumUsers() >= int(pgm.peerGroupSizeLimit) {
		pgm.mtx.Unlock()
		return false, nil
	}
	notify := pgm.remove(user)
	_, err := pg.RegisterUser(user)
	if err != nil {
		pgm.mtx.Unlock()
		return false, fmt.Errorf("failed to register user %s to peer group %s: %w", user, group, err)
	}
	pgm.userPeerGroupMap[user] = group
	pgm.peerToGroupMap[pg.PeerLookup(user).PeerUUID] = group
	pgm.mtx.Unlock()
	pgm.notifyUsers(notify)
	return true, nil
}

func (pgm *PeerGroupManager) HasUser(user uuid.UUID) bool {
	pgm.mtx.RLock()
	defer pgm.mtx.RUnlock()
	_, exists := pgm.userPeerGroupMap[user]
	return exists
}

func (pgm *PeerGroupManager) HasPeer(peer uuid.UUID) bool {
	pgm.mtx.RLock()
	defer pgm.mtx.RUnlock()
	_, exists := pgm.peerToGroupMap[peer]
	return exists
}

func (pgm *PeerGroupManager) PGLookupUser(user uuid.UUID) uuid.UUID {
	pgm.mtx.RLock()
	defer pgm.mtx.RUnlock()
	pgUUID, exists := pgm.userPeerGroupMap[user]
	if !exists {
		return uuid.Nil
	}
	return pgUUID
}

func (pgm *PeerGroupManager) PGLookupPeer(peer uuid.UUID) uuid.UUID {
	pgm.mtx.RLock()
	defer pgm.mtx.RUnlock()
	pgUUID, exists := pgm.peerToGroupMap[peer]
	if !exists {
		return uuid.Nil
	}
	return pgUUID
}

func (pgm *PeerGroupManager) UserToPeer(user uuid.UUID) uuid.UUID {
	pgm.mtx.RLock()
	defer pgm.mtx.RUnlock()
	pgUUID, exists := pgm.userPeerGroupMap[user]
	if !exists {
		return uuid.Nil
	}
	pg := pgm.peerGroupInstanceMap[pgUUID]
	peer := pg.PeerLookup(user)
	if peer == nil {
		return uuid.Nil
	}
	return peer.PeerUUID
}

func (pgm *PeerGroupManager) PeerToUser(peer uuid.UUID) uuid.UUID {
	pgm.mtx.RLock()
	defer pgm.mtx.RUnlock()
	pgUUID, exists := pgm.peerToGroupMap[peer]
	if !exists {
		return uuid.Nil
	}
	pg := pgm.peerGroupInstanceMap[pgUUID]
	user, exists := pg.peerUserMap[peer]
	if !exists {
		return uuid.Nil
	}
	return user
}

func (pgm *PeerGroupManager) NumGroups() int {
	pgm.mtx.RLock()
	defer pgm.mtx.RUnlock()
	return len(pgm.peerGroupInstanceMap)
}

func (pgm *PeerGroupManager) NumUsers() int {
	pgm.mtx.RLock()
	defer pgm.mtx.RUnlock()
	return len(pgm.userPeerGroupMap)
}

func (pgm *PeerGroupManager) ResolveDestinationUUIDs(peer uuid.UUID, clientID uint32) ([]uuid.UUID, error) {
	pgm.mtx.RLock()
	defer pgm.mtx.RUnlock()
	pgUUID, exists := pgm.peerToGroupMap[peer]
	if !exists {
		return nil, errors.New("invalid peer UUID")
	}
	pg := pgm.peerGroupInstanceMap[pgUUID]
	if clientID != math.MaxUint32 {
		dst, err := pg.ResolveDestinationUUID(peer, clientID)
		if err != nil {
			return nil, err
		}
		return []uuid.UUID{dst}, nil
	}
	srcCID := pg.peerClientIDMap[peer]
	dstUsers := make([]uuid.UUID, 0, pg.NumUsers())
	for _, cid := range pg.GetClients() {
		if srcCID == cid {
			continue
		}
		dst, err := pg.ResolveDestinationUUID(peer, cid)
		if err != nil {
			return nil, err
		}
		dstUsers = append(dstUsers, dst)
	}
	return dstUsers, nil
}

func NewPeerGroupManager(groupSizeLimit uint16, notifyCallback NotifyOnDestroyCallback) *PeerGroupManager {
	return &PeerGroupManager{
		userPeerGroupMap:     make(map[uuid.UUID]uuid.UUID),
		peerToGroupMap:       make(map[uuid.UUID]uuid.UUID),
		peerGroupInstanceMap: make(map[uuid.UUID]*PeerGroup),
		notifyCallback:       notifyCallback,
		peerGroupSizeLimit:   groupSizeLimit,
	}
}

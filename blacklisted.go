package main

import (
	"sync"
	"time"
)

type blacklistedPeers struct {
	peers map[string]int64 // map of blacklisted peers and the time the peer was blacklisted
	mux   sync.RWMutex
}

func newBlacklistedPeers() *blacklistedPeers {
	return &blacklistedPeers{
		peers: make(map[string]int64),
	}
}

func (bp *blacklistedPeers) add(peer string) {
	bp.mux.Lock()
	defer bp.mux.Unlock()

	// if it already exists, then do nothing as it's already blacklisted
	if _, ok := bp.peers[peer]; ok {
		return
	}

	bp.peers[peer] = time.Now().UTC().Unix()
}

func (bp *blacklistedPeers) remove(peer string) {
	bp.mux.Lock()
	defer bp.mux.Unlock()

	delete(bp.peers, peer)
}

func (bp *blacklistedPeers) isBlacklisted(peer string) bool {
	bp.mux.RLock()
	defer bp.mux.RUnlock()

	_, ok := bp.peers[peer]

	return ok
}

func (bp *blacklistedPeers) get() []string {
	bp.mux.RLock()
	defer bp.mux.RUnlock()

	var peers []string
	for peer := range bp.peers {
		peers = append(peers, peer)
	}

	return peers
}

func (bp *blacklistedPeers) isEmpty() bool {
	bp.mux.RLock()
	defer bp.mux.RUnlock()

	return len(bp.peers) == 0
}

// reset resets blacklisted peers
func (bp *blacklistedPeers) reset() {
	bp.mux.Lock()
	defer bp.mux.Unlock()

	bp.peers = make(map[string]int64)
}

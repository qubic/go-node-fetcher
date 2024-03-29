package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	qubic "github.com/qubic/go-node-connector"
	"math/rand"
	"sync"
	"time"
)

const qubicPort = "21841"
const cutDownInterval = 30

type reliablePeers struct {
	bp        *blacklistedPeers
	peers     []string
	updatedAt time.Time
	mu        sync.RWMutex
}

func newReliablePeers(bp *blacklistedPeers) *reliablePeers {
	return &reliablePeers{
		bp: bp,
	}
}

func (rp *reliablePeers) get() []string {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	return rp.peers
}

func (rp *reliablePeers) getResponse() PeersResponse {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	return PeersResponse{
		Peers:     rp.peers,
		UpdatedAt: rp.updatedAt.UTC().Unix(),
	}
}

func (rp *reliablePeers) set(peers []string) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	rp.updatedAt = time.Now()
	rp.peers = peers
}

func (rp *reliablePeers) build(peers []string) error {
	filteredPeers := rp.getPeersCurrentTick(peers)
	if len(filteredPeers) == 0 {
		return errors.New("no reliable peers found")
	}

	fmt.Printf("found %d reliable peers\n", len(filteredPeers))
	rp.set(filteredPeers)

	return nil
}

func (rp *reliablePeers) getOneRandom() string {
	peers := rp.get()

	if len(peers) == 0 {
		return ""
	}

	return peers[rand.Intn(len(peers))]
}

func (rp *reliablePeers) getPeersCurrentTick(peers []string) []string {
	filteredPeers := make([]string, 0, len(peers))
	peersTicks := make(map[string]int)
	maxTick := 0
	emptyTickPeers := 0
	for _, p := range peers {
		tick, err := rp.getPeerCurrentTick(p)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if tick == 0 {
			emptyTickPeers++
		}
		if tick > maxTick {
			maxTick = tick
		}
		peersTicks[p] = tick
	}
	for p, t := range peersTicks {
		if maxTick-t >= cutDownInterval {
			fmt.Printf("Peer %s has tick %d, which is %d ticks behind, cutting it off. Proceed to blacklist\n", p, t, maxTick-t)
			rp.bp.add(p)
		} else {
			filteredPeers = append(filteredPeers, p)
		}
	}

	fmt.Println("Empty tick peers: ", emptyTickPeers)

	return filteredPeers
}

func (rp *reliablePeers) getPeerCurrentTick(peer string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	qc, err := qubic.NewClient(ctx, peer, "21841")
	if err != nil {
		return 0, errors.Wrap(err, "creating qubic connection")
	}
	defer qc.Close()

	currentTick, err := qc.GetTickInfo(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "getting tick info")
	}

	return int(currentTick.Tick), nil
}

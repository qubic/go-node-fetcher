package main

import (
	"context"
	"github.com/pkg/errors"
	qubic "github.com/qubic/go-node-connector"
	"log"
	"math/rand"
	"sync"
	"time"
)

type distinctPeers struct {
	bp                        *blacklistedPeers
	peers                     map[string]struct{}
	maxPeers                  int
	mux                       sync.RWMutex
	exchangeConnectionTimeout time.Duration
}

func newDistinctPeers(startingPeer string, maxPeers int, exchangeConnectionTimeout time.Duration, bp *blacklistedPeers) *distinctPeers {
	dp := distinctPeers{
		bp:                        bp,
		peers:                     make(map[string]struct{}, maxPeers),
		maxPeers:                  maxPeers,
		exchangeConnectionTimeout: exchangeConnectionTimeout,
	}
	dp.setPeers([]string{startingPeer})

	return &dp
}

func (p *distinctPeers) build() ([]string, error) {
	peer := p.getRandomPeer()
	err := p.exchangePeerList(peer)
	if err != nil {
		return nil, errors.Wrap(err, "exchanging peer list")
	}

	if p.isEmpty() {
		if p.bp.isEmpty() {
			log.Println("No distinct peers found, no blacklisted peers found")
		}
		return p.bp.get(), nil
	}

	return p.get(), nil
}

func (p *distinctPeers) exchangePeerList(peer string) error {
	if p.isFull() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.exchangeConnectionTimeout)
	defer cancel()

	qc, err := qubic.NewClient(ctx, peer, qubicPort)
	if err != nil {
		return errors.Wrap(err, "creating new connection")
	}
	qc.Close()

	unmetPeers := p.getUnmetPeers(qc.Peers)
	p.setPeers(unmetPeers)

	for _, peer := range unmetPeers {
		p.exchangePeerList(peer)
	}

	return nil
}

func (p *distinctPeers) getUnmetPeers(peers []string) []string {
	var unmetPeers []string
	for _, peer := range peers {
		if _, ok := p.peers[peer]; !ok {
			unmetPeers = append(unmetPeers, peer)
		}
	}

	return unmetPeers
}

func (p *distinctPeers) get() []string {
	p.mux.RLock()
	defer p.mux.RUnlock()

	var peers []string
	for peer := range p.peers {
		peers = append(peers, peer)
	}

	return peers
}

func (p *distinctPeers) setPeers(peers []string) {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, peer := range peers {
		p.peers[peer] = struct{}{}
	}
}

func (p *distinctPeers) isFull() bool {
	p.mux.RLock()
	defer p.mux.RUnlock()

	return len(p.peers) >= p.maxPeers
}

func (p *distinctPeers) isEmpty() bool {
	p.mux.RLock()
	defer p.mux.RUnlock()

	return len(p.peers) == 0
}

func (p *distinctPeers) getRandomPeer() string {
	p.mux.RLock()
	defer p.mux.RUnlock()

	if p.isEmpty() {
		return ""
	}

	peers := p.get()

	return peers[rand.Intn(len(p.peers))]
}

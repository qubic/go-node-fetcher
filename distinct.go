package main

import (
	"context"
	"github.com/cockroachdb/pebble"
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
	db                        *pebble.DB
	mux                       sync.RWMutex
	exchangeConnectionTimeout time.Duration
}

func newDistinctPeers(startingPeers []string, maxPeers int, exchangeConnectionTimeout time.Duration, bp *blacklistedPeers, db *pebble.DB) *distinctPeers {
	dp := distinctPeers{
		bp:                        bp,
		peers:                     make(map[string]struct{}, maxPeers),
		maxPeers:                  maxPeers,
		db:                        db,
		exchangeConnectionTimeout: exchangeConnectionTimeout,
	}
	dp.setPeers(startingPeers)

	return &dp
}

func (p *distinctPeers) build() ([]string, error) {
	peer, err := p.getRandomPeer()
	if err != nil {
		return nil, errors.Wrap(err, "getting random peer")
	}
	if len(peer) == 0 {
		peer = p.get()[0]
	}

	err = p.exchangePeerList(peer)
	if err != nil {
		return nil, errors.Wrap(err, "exchanging peer list")
	}

	if p.isEmpty() {
		if p.bp.isEmpty() {
			log.Println("No distinct peers found, no blacklisted peers found")
		}
		return p.bp.get(), nil
	}

	err = storePeers(p.db, p.get())
	if err != nil {
		return nil, errors.Wrap(err, "storing peers")
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

func (p *distinctPeers) getRandomPeer() (string, error) {
	peers, err := retrievePeers(p.db)
	if err != nil {
		return "", errors.Wrap(err, "retrieving peers from store")
	}

	if len(peers) == 0 {
		return "", nil
	}

	return peers[rand.Intn(len(p.peers))], nil
}

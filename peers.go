package main

import (
	"github.com/cockroachdb/pebble"
	"github.com/pkg/errors"
	"time"
)

type Peers struct {
	dp *distinctPeers
	rp *reliablePeers
}

type PeersResponse struct {
	Peers     []string
	UpdatedAt int64
}

func NewPeers(startingPeer string, maxPeers int, exchangeConnectionTimeout time.Duration, db *pebble.DB) (*Peers, error) {
	storedPeers, err := retrievePeers(db)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving peers from store")
	}

	initialPeers := append(storedPeers, startingPeer)
	bp := newBlacklistedPeers()
	p := Peers{
		dp: newDistinctPeers(initialPeers, maxPeers, exchangeConnectionTimeout, bp, db),
		rp: newReliablePeers(bp),
	}

	return &p, nil
}

func (p *Peers) Compute() error {
	peers, err := p.dp.build()
	if err != nil {
		return errors.Wrapf(err, "building distinct peers")
	}

	err = p.rp.build(peers)
	if err != nil {
		return errors.Wrap(err, "building reliable peers list")
	}

	return nil
}

func (p *Peers) Get() PeersResponse {
	return p.rp.getResponse()
}

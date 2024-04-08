package main

import (
	"github.com/cockroachdb/pebble"
	"github.com/pkg/errors"
	"time"
)

type Peers struct {
	dp            *distinctPeers
	rp            *reliablePeers
	fixedPeerList bool
}

type PeersResponse struct {
	Peers     []string
	UpdatedAt int64
}

func NewPeers(fixedPeerList bool, startingPeers []string, whitelistedPeers []string, maxPeers int, exchangeConnectionTimeout time.Duration, db *pebble.DB) (*Peers, error) {

	if !fixedPeerList {
		storedPeers, err := retrievePeers(db)
		if err != nil {
			return nil, errors.Wrap(err, "retrieving peers from store")
		}

		startingPeers = append(startingPeers, storedPeers...)
	}

	bp := newBlacklistedPeers()
	p := Peers{
		dp:            newDistinctPeers(startingPeers, whitelistedPeers, maxPeers, exchangeConnectionTimeout, bp, db),
		rp:            newReliablePeers(bp),
		fixedPeerList: fixedPeerList,
	}

	return &p, nil
}

func (p *Peers) Compute() error {
	peers, err := p.dp.build(p.fixedPeerList)
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

package main

import (
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

func NewPeers(startingPeer string, maxPeers int, exchangeConnectionTimeout time.Duration) *Peers {
	bp := newBlacklistedPeers()
	p := Peers{
		dp: newDistinctPeers(startingPeer, maxPeers, exchangeConnectionTimeout, bp),
		rp: newReliablePeers(bp),
	}

	return &p
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

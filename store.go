package main

import (
	"bytes"
	"encoding/json"
	"github.com/cockroachdb/pebble"
	"github.com/pkg/errors"
)

func storePeers(db *pebble.DB, peers []string) error {
	serialized, err := json.Marshal(peers)
	if err != nil {
		return errors.Wrap(err, "serializing peers")
	}

	err = db.Set([]byte("peers"), serialized, &pebble.WriteOptions{Sync: true})
	if err != nil {
		return errors.Wrap(err, "setting peers")
	}

	return nil
}

func retrievePeers(db *pebble.DB) ([]string, error) {
	value, closer, err := db.Get([]byte("peers"))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return []string{}, nil
		}

		return nil, errors.Wrap(err, "getting peers")
	}
	defer closer.Close()

	peers := make([]string, 0)
	err = json.NewDecoder(bytes.NewReader(value)).Decode(&peers)
	if err != nil {
		return nil, errors.Wrap(err, "decoding peers")
	}

	return peers, err
}

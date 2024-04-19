package main

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	rp *Peers
}

type response struct {
	Peers       []string `json:"peers"`
	Length      int      `json:"length"`
	LastUpdated int64    `json:"last_updated"`
	MaxTick     int      `json:"max_tick"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	p := h.rp.Get()
	res := response{
		Peers:       p.Peers,
		Length:      len(p.Peers),
		LastUpdated: p.UpdatedAt,
		MaxTick:     p.MaxTick,
	}
	b, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(b)
}

package main

import (
	"fmt"
	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"time"
)

const prefix = "NODE_FETCHER"

func main() {
	err := run()
	if err != nil {
		log.Fatal(err.Error())
	}

}

func run() error {
	var cfg struct {
		Server struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Qubic struct {
			StartingPeerIP  string        `conf:"default:92.50.108.113"`
			MaxPeers        int           `conf:"default:50"`
			ExchangeTimeout time.Duration `conf:"default:2s"`
		}
	}

	if err := conf.Parse(os.Args[1:], prefix, &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage(prefix, &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString(prefix, &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config version")
			}
			fmt.Println(version)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main: Config :\n%v\n", out)

	rp := NewPeers(cfg.Qubic.StartingPeerIP, cfg.Qubic.MaxPeers, cfg.Qubic.ExchangeTimeout)
	err = rp.Compute()
	if err != nil {
		return errors.Wrap(err, "computing first batch of reliable peers")
	}

	h := Handler{rp: rp}

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-ticker.C:
				err := rp.Compute()
				if err != nil {
					log.Printf("Computing reliable peers: %s", err.Error())
				}
			}
		}
	}()

	fmt.Println("Server started")
	http.HandleFunc("/peers", h.Handle)

	log.Fatal(http.ListenAndServe(":8080", nil))

	return nil
}

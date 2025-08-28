package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/geekp2p/pixelval/internal/config"
	"github.com/geekp2p/pixelval/internal/gateway"
	"github.com/geekp2p/pixelval/internal/p2p"
)

func main() {
	_ = godotenv.Load(".env")

	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start P2P node (host + pubsub + topic/sub)
	host, psub, topic, sub := p2p.StartNode(ctx, cfg)

	// optional embedded relay (CircuitV2) in the SAME binary
	if cfg.EnableEmbeddedRelay {
		p2p.StartEmbeddedRelay(ctx, cfg)
	}

	// start Web UI + WS gateway
	go gateway.StartWeb(ctx, host, psub, topic, sub, cfg)

	// start GUI (Electron via astilectron) if enabled
	if cfg.GUI {
		go gateway.OpenGUI(cfg)
	}

	// wait for signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("PixelVal: shutting down...")
}

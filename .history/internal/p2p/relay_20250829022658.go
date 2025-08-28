<<<<<<< HEAD
package p2p

import (
	"context"
	"fmt"
	"strings"

	libp2p "github.com/libp2p/go-libp2p"
	relayv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/geekp2p/pixelval/internal/config"
)

func StartEmbeddedRelay(ctx context.Context, cfg config.Config) {
	announce := []ma.Multiaddr{}
	for _, s := range cfg.AnnounceAddrs {
		if s = strings.TrimSpace(s); s == "" {
			continue
		}
		if m, err := ma.NewMultiaddr(s); err == nil {
			announce = append(announce, m)
		}
	}

	opts := []libp2p.Option{libp2p.ListenAddrStrings(cfg.RelayListen)}
	if len(announce) > 0 {
		opts = append(opts, libp2p.AddrsFactory(func(in []ma.Multiaddr) []ma.Multiaddr {
			return append(in, announce...)
		}))
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		fmt.Println("embedded relay error:", err)
		return
	}
	go func() {
		<-ctx.Done()
		_ = h.Close()
	}()

	_, err = relayv2.New(h)
	if err != nil {
		fmt.Println("relay init:", err)
		return
	}

	fmt.Printf("✅ Embedded Relay PeerID: %s\n", h.ID())
	for _, a := range h.Addrs() {
		fmt.Printf("📡 Relay Listening: %s/p2p/%s\n", a, h.ID())
	}
}
=======
package p2p

import (
	"context"
	"fmt"
	"strings"

	libp2p "github.com/libp2p/go-libp2p"
	relayv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/geekp2p/pixelval/internal/config"
)

func StartEmbeddedRelay(ctx context.Context, cfg config.Config) {
	announce := []ma.Multiaddr{}
	for _, s := range cfg.AnnounceAddrs {
		if s = strings.TrimSpace(s); s == "" {
			continue
		}
		if m, err := ma.NewMultiaddr(s); err == nil {
			announce = append(announce, m)
		}
	}

	opts := []libp2p.Option{libp2p.ListenAddrStrings(cfg.RelayListen)}
	if len(announce) > 0 {
		opts = append(opts, libp2p.AddrsFactory(func(in []ma.Multiaddr) []ma.Multiaddr {
			return append(in, announce...)
		}))
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		fmt.Println("embedded relay error:", err)
		return
	}
	go func() {
		<-ctx.Done()
		_ = h.Close()
	}()

	_, err = relayv2.New(h)
	if err != nil {
		fmt.Println("relay init:", err)
		return
	}

	fmt.Printf("✅ Embedded Relay PeerID: %s\n", h.ID())
	for _, a := range h.Addrs() {
		fmt.Printf("📡 Relay Listening: %s/p2p/%s\n", a, h.ID())
	}
}
>>>>>>> 5afbf8a8359422fae8cbab6291d9f50e384f4ba8

package p2p

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	clientv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	tcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	ma "github.com/multiformats/go-multiaddr"

	"pixelval/internal/config"
)

type mdnsNotifee struct{ h host.Host }

func (n *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	_ = n.h.Connect(context.Background(), pi)
}

func StartNode(ctx context.Context, cfg config.Config) (host.Host, *pubsub.PubSub, *pubsub.Topic, *pubsub.Subscription) {
	opts := []libp2p.Option{
		libp2p.Muxer(yamux.ID, yamux.DefaultTransport),
		libp2p.Transport(tcp.NewTCPTransport),
	}
	if cfg.ListenTCP != "" {
		opts = append(opts, libp2p.ListenAddrStrings(cfg.ListenTCP))
	}
	if cfg.ListenQUIC != "" {
		opts = append(opts,
			libp2p.Transport(quic.NewTransport),
			libp2p.ListenAddrStrings(cfg.ListenQUIC),
		)
	}
	if cfg.EnableUPnP {
		opts = append(opts, libp2p.NATPortMap())
	}
	if cfg.EnableRelayClient {
		opts = append(opts, libp2p.EnableRelay())
	}
	if cfg.EnableHolePunch {
		opts = append(opts, libp2p.EnableHolePunching())
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		log.Fatalf("libp2p.New: %v", err)
	}

	// print addrs
	log.Printf("PeerID: %s", h.ID())
	for _, a := range h.Addrs() {
		log.Printf("Listen: %s/p2p/%s", a, h.ID())
	}

	// mDNS for LAN
	svc := mdns.NewMdnsService(h, cfg.AppRoom, &mdnsNotifee{h: h})
	go func() {
		<-ctx.Done()
		_ = svc.Close()
	}()

	// connect to relay if provided
	if strings.TrimSpace(cfg.RelayAddr) != "" {
		if m, err := ma.NewMultiaddr(cfg.RelayAddr); err == nil {
			if pi, err := peer.AddrInfoFromP2pAddr(m); err == nil {
				h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
				_ = h.Connect(ctx, *pi)
				// reserve a slot (so we can be relayed)
				_, _ = clientv2.Reserve(ctx, h, *pi)
			}
		}
	}

	// connect to bootstrap peers
	for _, s := range cfg.BootstrapPeers {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		m, err := ma.NewMultiaddr(s)
		if err != nil {
			continue
		}
		pi, err := peer.AddrInfoFromP2pAddr(m)
		if err != nil {
			continue
		}
		if pi.ID == h.ID() {
			continue
		}
		h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
		_ = h.Connect(ctx, *pi)
	}

	// log peer changes
	h.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, c network.Conn) {
			log.Printf("↔ connected %s", c.RemotePeer())
		},
		DisconnectedF: func(n network.Network, c network.Conn) {
			log.Printf("× disconnected %s", c.RemotePeer())
		},
	})

	// pubsub
	psub, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		log.Fatalf("GossipSub: %v", err)
	}
	topic, err := psub.Join("room:" + cfg.AppRoom)
	if err != nil {
		log.Fatalf("Join topic: %v", err)
	}
	sub, err := topic.Subscribe()
	if err != nil {
		log.Fatalf("Subscribe: %v", err)
	}

	// simple keepalive log
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				fmt.Printf("Peers: %d\n", len(h.Network().Peers()))
			}
		}
	}()

	return h, psub, topic, sub
}

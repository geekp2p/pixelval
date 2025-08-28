package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AppName string `yaml:"app_name"`
	AppRoom string `yaml:"app_room"`
	WebAddr string `yaml:"web_addr"`
	GUI     bool   `yaml:"gui"`

	// P2P
	ListenTCP         string   `yaml:"listen_tcp"`
	ListenQUIC        string   `yaml:"listen_quic"`
	EnableUPnP        bool     `yaml:"enable_upnp"`
	EnableHolePunch   bool     `yaml:"enable_holepunch"`
	EnableRelayClient bool     `yaml:"enable_relay_client"`
	BootstrapPeers    []string `yaml:"bootstrap_peers"`
	RelayAddr         string   `yaml:"relay_addr"`

	// Embedded Relay (in-process CircuitV2)
	EnableEmbeddedRelay bool     `yaml:"enable_embedded_relay"`
	RelayListen         string   `yaml:"relay_listen"`
	AnnounceAddrs       []string `yaml:"announce_addrs"`
}

func Load() Config {
	cfg := Config{
		AppName: "PixelVal",
		AppRoom: "main",
		WebAddr: ":8081",
		GUI:     true,

		ListenTCP:         "/ip4/0.0.0.0/tcp/4001",
		ListenQUIC:        "", // e.g. "/ip4/0.0.0.0/udp/4001/quic-v1"
		EnableUPnP:        true,
		EnableHolePunch:   true,
		EnableRelayClient: true,

		EnableEmbeddedRelay: false,
		RelayListen:         "/ip4/0.0.0.0/tcp/4003",
	}

	// merge config.yaml (optional)
	if b, err := os.ReadFile("config.yaml"); err == nil {
		_ = yaml.Unmarshal(b, &cfg)
	}

	// ENV overrides
	overrideStr := func(k, v string) string {
		if s := strings.TrimSpace(os.Getenv(k)); s != "" {
			return s
		}
		return v
	}
	overrideBool := func(k string, v bool) bool {
		s := strings.ToLower(strings.TrimSpace(os.Getenv(k)))
		if s == "" {
			return v
		}
		return s == "1" || s == "true" || s == "yes" || s == "y"
	}
	overrideCSV := func(k string, base []string) []string {
		s := strings.TrimSpace(os.Getenv(k))
		if s == "" {
			return base
		}
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}

	cfg.AppRoom = overrideStr("APP_ROOM", cfg.AppRoom)
	cfg.WebAddr = overrideStr("WEB_ADDR", cfg.WebAddr)
	cfg.GUI = overrideBool("GUI", cfg.GUI)

	cfg.ListenTCP = overrideStr("LISTEN_TCP", cfg.ListenTCP)
	cfg.ListenQUIC = overrideStr("LISTEN_QUIC", cfg.ListenQUIC)
	cfg.EnableUPnP = overrideBool("ENABLE_UPNP", cfg.EnableUPnP)
	cfg.EnableHolePunch = overrideBool("ENABLE_HOLEPUNCH", cfg.EnableHolePunch)
	cfg.EnableRelayClient = overrideBool("ENABLE_RELAY_CLIENT", cfg.EnableRelayClient)
	cfg.BootstrapPeers = overrideCSV("BOOTSTRAP_PEERS", cfg.BootstrapPeers)
	cfg.RelayAddr = overrideStr("RELAY_ADDR", cfg.RelayAddr)

	cfg.EnableEmbeddedRelay = overrideBool("ENABLE_EMBEDDED_RELAY", cfg.EnableEmbeddedRelay)
	cfg.RelayListen = overrideStr("RELAY_LISTEN", cfg.RelayListen)
	cfg.AnnounceAddrs = overrideCSV("ANNOUNCE_ADDRS", cfg.AnnounceAddrs)

	return cfg
}

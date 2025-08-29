# PixelVal

A **single-binary** P2P world: libp2p node (+optional embedded relay) with a built-in **Web UI (8081)** and **Electron GUI**.  
Start as P2P chat; grow into an **8-bit persistent simulation** where characters live on even when you’re offline (as long as at least one peer is up).

- ✅ One binary: node + (optional) relay + Web/Electron UI
- 🌐 GossipSub pubsub, mDNS, optional QUIC/UPnP/Hole-punch
- 🧠 AI-ready: data shards can be encrypted, shared, and used to keep lives running (future)
- 🗺️ Multi-map: main city in 8-bit; user-generated maps welcome

## Quick Start
```bash
go mod tidy
go run ./cmd/pixelval
# UI: http://127.0.0.1:8081  (Electron will open if GUI=1)
```

### Docker

Build and run inside a container without needing Node.js or nginx:

```bash
docker build -t pixelval .
docker run --rm -p 8081:8081 -p 4001:4001 -p 4003:4003 pixelval
```

Run multiple peers by mapping different host ports and, if needed, set
`BOOTSTRAP_PEERS` so containers can discover each other:

```bash
# assume the first container prints: /ip4/172.17.0.2/tcp/4001/p2p/<peerID>
docker run --rm -p 8082:8081 -p 4002:4001 \
  -e BOOTSTRAP_PEERS=/ip4/172.17.0.2/tcp/4001/p2p/<peerID> pixelval
```

### License
```
🔓 Open source under a "copyleft-style" license: anyone may use, sell, or distribute games built on PixelVal, but code modifications must be published back to the community.
```
package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/asticode/go-astikit"
	"github.com/asticode/go-astilectron"

	"pixelval/internal/config"
	"strconv"
)

func (g *gateway) makeKey(cm ChatMsg) string {
	return cm.ID + ":" + strconv.FormatInt(cm.Ts, 10) + ":" + cm.Text
}

func (g *gateway) seenBefore(cm ChatMsg) bool {
	key := g.makeKey(cm)
	if _, ok := g.seen.LoadOrStore(key, time.Now()); ok {
		return true
	}
	return false
}

type ChatMsg struct {
	From string `json:"from"`
	ID   string `json:"id"`
	Text string `json:"text"`
	Ts   int64  `json:"ts"`
}

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

type gateway struct {
	ctx      context.Context
	h        host.Host
	psub     *pubsub.PubSub
	topic    *pubsub.Topic
	sub      *pubsub.Subscription
	upgrader websocket.Upgrader

	room string
	nick string

	mu      sync.RWMutex
	clients map[*wsClient]bool
	cfg     config.Config

	seen sync.Map
}

func StartWeb(ctx context.Context, h host.Host, psub *pubsub.PubSub, topic *pubsub.Topic, sub *pubsub.Subscription, cfg config.Config) {
	g := &gateway{
		ctx:      ctx,
		h:        h,
		psub:     psub,
		topic:    topic,
		sub:      sub,
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		room:     cfg.AppRoom,
		nick:     "guest",
		clients:  map[*wsClient]bool{},
		cfg:      cfg,
	}

	// go g.consume(ctx, sub)
	go g.consume()

	http.HandleFunc("/", g.serveIndex)
	http.HandleFunc("/app.js", g.serveApp)
	http.HandleFunc("/style.css", g.serveCSS)
	http.HandleFunc("/ws", g.serveWS)
	http.HandleFunc("/config", g.handleConfig)

	addr := cfg.WebAddr
	log.Printf("🌐 PixelVal UI: http://127.0.0.1%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Println("http server:", err)
	}
}

func (g *gateway) serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(IndexHTML)
}
func (g *gateway) serveApp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(AppJS)
}
func (g *gateway) serveCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.Write(AppCSS)
}

func (g *gateway) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &wsClient{conn: conn, send: make(chan []byte, 32)}
	g.mu.Lock()
	g.clients[c] = true
	g.mu.Unlock()

	// writer
	go func() {
		for b := range c.send {
			_ = c.conn.WriteMessage(websocket.TextMessage, b)
		}
	}()

	// reader -> publish
	go func() {
		defer func() {
			g.mu.Lock()
			delete(g.clients, c)
			g.mu.Unlock()
			_ = conn.Close()
		}()
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg := ChatMsg{
				From: g.nick,
				ID:   g.h.ID().String(),
				Text: string(data),
				Ts:   time.Now().Unix(),
			}

			// log outbound chat for easier debugging
			log.Printf("send[%s]: %s", msg.ID, msg.Text)

			// ✨ กันซ้ำ: mark seen ทันที
			_ = g.seenBefore(msg)

			b, _ := json.Marshal(msg)

			// แสดงผลโลคัลทันที (เร็ว) + ให้ consume() ข้าม self
			g.broadcast(b)

			// กระจายออก pubsub
			_ = g.topic.Publish(context.Background(), b)
		}
	}()
}

func (g *gateway) consume() {
	for {
		g.mu.RLock()
		sub := g.sub
		ctx := g.ctx
		selfID := g.h.ID().String()
		g.mu.RUnlock()

		msg, err := sub.Next(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}
		var cm ChatMsg
		if err := json.Unmarshal(msg.Data, &cm); err != nil {
			continue
		}

		// ข้อความที่เราส่งเอง ไม่ต้อง broadcast ซ้ำ
		if cm.ID == selfID {
			continue
		}

		// กันซ้ำด้วย seen-key (ดูข้อ 2)
		if g.seenBefore(cm) {
			continue
		}

		b, _ := json.Marshal(cm)
		g.broadcast(b)
	}
}

// func (g *gateway) consume() {
// 	for {
// 		g.mu.RLock()
// 		sub := g.sub
// 		ctx := g.ctx
// 		g.mu.RUnlock()

// 		msg, err := sub.Next(ctx)
// 		if err != nil {
// 			if ctx.Err() != nil {
// 				return
// 			}
// 			continue
// 		}
// 		var cm ChatMsg
// 		if err := json.Unmarshal(msg.Data, &cm); err != nil {
// 			continue
// 		}
// 		b, _ := json.Marshal(cm)

// 		g.broadcast(b)
// 	}
// }

func (g *gateway) broadcast(b []byte) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for c := range g.clients {
		select {
		case c.send <- b:
		default:
		}
		// g.mu.RUnlock()
	}
}

func (g *gateway) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		resp := struct {
			Nick string `json:"nick"`
			Room string `json:"room"`
			ID   string `json:"id"`
		}{g.nick, g.room, g.h.ID().String()}
		_ = json.NewEncoder(w).Encode(resp)
	case http.MethodPost:
		var req struct{ Nick, Room string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		if req.Nick != "" {
			g.mu.Lock()
			g.nick = req.Nick
			g.mu.Unlock()
		}
		if req.Room != "" && req.Room != g.room {
			newTopic, err := g.psub.Join("room:" + req.Room)
			if err != nil {
				http.Error(w, "join room failed", 500)
				return
			}
			newSub, err := newTopic.Subscribe()
			if err != nil {
				_ = newTopic.Close()
				http.Error(w, "subscribe failed", 500)
				return
			}
			g.sub.Cancel()
			_ = g.topic.Close()
			g.mu.Lock()
			g.room = req.Room
			g.topic = newTopic
			g.sub = newSub
			g.mu.Unlock()
		}
		w.WriteHeader(204)
	default:
		w.WriteHeader(405)
	}
}

// Electron GUI
func OpenGUI(cfg config.Config) {
	logger := log.New(log.Writer(), "", log.LstdFlags)
	a, err := astilectron.New(logger, astilectron.Options{
		AppName:           cfg.AppName,
		BaseDirectoryPath: "data",
	})
	if err != nil {
		log.Println("astilectron:", err)
		return
	}
	defer a.Close()
	if err := a.Start(); err != nil {
		log.Println("astilectron start:", err)
		return
	}
	w, err := a.NewWindow(fmt.Sprintf("http://127.0.0.1%s", cfg.WebAddr), &astilectron.WindowOptions{
		Center: astikit.BoolPtr(true),
		Width:  astikit.IntPtr(960),
		Height: astikit.IntPtr(640),
	})
	if err != nil {
		log.Println("new window:", err)
		return
	}
	if err := w.Create(); err != nil {
		log.Println("create window:", err)
		return
	}
	a.Wait()
}

package main

import (
	"context"
	"discordrpc/internal"
	"discordrpc/internal/token"
	"log"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type WsWrapper struct {
	ws        *websocket.Conn
	seq       *int
	parentCtx context.Context
	ctx       context.Context
	cancel    *context.CancelFunc
	user      string
}

func (w *WsWrapper) init() {
	ctx, cancel := context.WithCancel(w.parentCtx)
	if w.cancel != nil {
		(*w.cancel)()
	}
	w.ctx = ctx
	w.cancel = &cancel
	var err error
	w.ws, _, err = websocket.Dial(w.ctx, "wss://gateway.discord.gg/?encoding=json&v=9", nil)
	if err != nil {
		panic(err)
	}
	w.ws.SetReadLimit(128 * 1024 * 1024) // 128 MB
	err = w.ws.Write(w.ctx, websocket.MessageText, []byte(internal.InitData(token.GetToken(w.user))))
	if err != nil {
		panic(err)
	}
	w.seq = nil
	go w.readLoop()
	go w.pingLoop()
}

func (w *WsWrapper) readLoop() {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			var v struct {
				Op int  `json:"op"`
				S  *int `json:"s"`
			}
			if err := wsjson.Read(w.ctx, w.ws, &v); err != nil {
				log.Println(err)
				w.init()
				return
			}
			if v.S != nil {
				w.seq = v.S
			}
		}
	}
}

func (w *WsWrapper) pingLoop() {
	const interval = 41250 * time.Millisecond
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-time.After(interval):
			if err := wsjson.Write(w.ctx, w.ws, map[string]any{
				"op": 1,
				"d":  w.seq,
			}); err != nil {
				log.Println(err)
				w.init()
				return
			}
		}
	}
}

func (w *WsWrapper) Write(ctx context.Context, v interface{}) error {
	return wsjson.Write(ctx, w.ws, v)
}

func NewWsWrapper(ctx context.Context) *WsWrapper {
	w := &WsWrapper{
		parentCtx: ctx,
	}
	w.init()
	return w
}

func startDiscordRpc(ctx context.Context) func(internal.Activity) {
	discordRpc := NewWsWrapper(ctx)

	return func(activity internal.Activity) {
		activity.Name = activity.State
		var err error
		activity.Assets.LargeImage, err = internal.FetchExternalAssets(activity)
		if err != nil {
			log.Println("Failed to fetch external assets: %w", err)
		}
		log.Println(activity.Details)
		log.Println(strings.Split(activity.Assets.LargeImage, "https/")[1])
		err = discordRpc.Write(ctx, map[string]any{
			"op": 3, "d": map[string]any{
				"status":     "dnd",
				"since":      0,
				"activities": []internal.Activity{activity},
				"afk":        false,
			},
		})
		if err != nil {
			log.Println(err)
			discordRpc.init()
			return
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	discordRpc := startDiscordRpc(ctx)
	localWs, _, err := websocket.Dial(ctx, "ws://127.0.0.1:1337/", nil)
	if err != nil {
		panic(err)
	}
	for {
		var activity struct {
			Activity internal.Activity `json:"activity"`
		}
		if err := wsjson.Read(ctx, localWs, &activity); err != nil {
			panic(err)
		}
		discordRpc(activity.Activity)
	}
}

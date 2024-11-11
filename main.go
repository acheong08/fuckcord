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
		log.Println(activity.Details)
		log.Println(strings.Split(activity.Assets.LargeImage, "https/")[1])
		err := discordRpc.Write(ctx, map[string]any{
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
	largeImg, err := internal.FetchExternalAssets("https://cdn.discordapp.com/app-icons/432980957394370572/c1864b38910c209afd5bf6423b672022.webp", "1297964925133783050")
	if err != nil {
		panic(err)
	}
	smImg, err := internal.FetchExternalAssets("https://static.wikia.nocookie.net/fortnite/images/6/6c/Unreal_-_Icon_-_Fortnite.png", "1297964925133783050")
	if err != nil {
		panic(err)
	}
	println(time.Now().Unix())

	activity := internal.Activity{
		ApplicationId: "1297964925133783050",
		Type:          0,
		Flags:         1,
		Instance:      true,
		State:         "Fortnite",
		Details:       "Battle Royale",
		Timestamps: struct {
			Start int64 `json:"start"`
			End   int64 `json:"end"`
		}{
			Start: time.Now().Add(-time.Hour + 23*time.Minute).UnixMilli(),
			End:   time.Now().Add(time.Hour * 24).UnixMilli(),
		},
		Name: "Fortnite",
		Assets: struct {
			LargeImage string `json:"large_image"`
			SmallImage string `json:"small_image"`
			SmallText  string `json:"small_text"`
		}{
			LargeImage: largeImg,
			SmallImage: smImg,
			SmallText:  "Unreal Rank",
		},
		Party: internal.Party{
			Id:      "ae488379-351d-4a4f-ad32-2b9b01c91657",
			Privacy: 1,
			Size:    []int{3, 5},
		},
	}
	for {
		discordRpc(activity)
		time.Sleep(1 * time.Minute)
	}
}

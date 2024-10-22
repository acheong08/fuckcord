package main

import (
	"context"
	"discordrpc/typings"
	"log"
	"os"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func startDiscordRpc(ctx context.Context, token string) func(typings.Activity) error {
	discordRpc, _, err := websocket.Dial(ctx, "wss://gateway.discord.gg/?encoding=json&v=9", nil)
	discordRpc.SetReadLimit(128 * 1024 * 1024) // 128 MB
	if err != nil {
		panic(err)
	}
	err = discordRpc.Write(ctx, websocket.MessageText, []byte(typings.InitData(token)))
	if err != nil {
		panic(err)
	}
	var seq *int = nil
	go func() {
		for {
			if err := wsjson.Write(ctx, discordRpc, map[string]any{
				"op": 1,
				"d":  seq,
			}); err != nil {
				panic(err)
			}
			time.Sleep(41250 * time.Millisecond)
		}
	}()
	type genericResponse struct {
		Op int  `json:"op"`
		S  *int `json:"s"`
	}
	go func() {
		for {
			var v genericResponse
			if err := wsjson.Read(ctx, discordRpc, &v); err != nil {
				panic(err)
			}
			if v.S != nil {
				seq = v.S
			}
		}
	}()
	return func(activity typings.Activity) error {
		activity.Name = "🔥🔥🔥🔥"
		err := activity.FetchExternalAssets(token)
		if err != nil {
			activity.Assets.LargeImage = typings.DefaultCover()
		}
		log.Println(activity.Details)
		log.Println(strings.Split(activity.Assets.LargeImage, "https/")[1])
		return wsjson.Write(ctx, discordRpc, map[string]any{
			"op": 3, "d": map[string]any{
				"status":     "dnd",
				"since":      0,
				"activities": []typings.Activity{activity},
				"afk":        false,
			},
		})
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	token := os.Getenv("DISCORD_TOKEN")
	discordRpc := startDiscordRpc(ctx, token)
	localWs, _, err := websocket.Dial(ctx, "ws://127.0.0.1:1337/", nil)
	if err != nil {
		panic(err)
	}
	for {
		var activity struct {
			Activity typings.Activity `json:"activity"`
		}
		if err := wsjson.Read(ctx, localWs, &activity); err != nil {
			panic(err)
		}
		if err := discordRpc(activity.Activity); err != nil {
			panic(err)
		}
	}
}

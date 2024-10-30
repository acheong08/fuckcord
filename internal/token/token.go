package token

import (
	"encoding/json"
	"os"
)

var (
	defaultToken string
	tokens       map[string]string
)

func init() {
	defaultToken = os.Getenv("DISCORD_TOKEN")
	if defaultToken == "" {
		panic("DISCORD_TOKEN is not set")
	}
	if defaultToken[0] != '{' {
		return
	}
	err := json.Unmarshal([]byte(defaultToken), &tokens)
	if err != nil {
		panic(err)
	}
	for _, v := range tokens {
		defaultToken = v
		break
	}
}

func GetToken(k string) string {
	if k == "" {
		return defaultToken
	}
	if token, ok := tokens[k]; ok {
		return token
	}
	return defaultToken
}

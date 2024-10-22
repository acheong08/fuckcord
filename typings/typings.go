package typings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type browserProperties struct {
	Os                string  `json:"os"`
	Browser           string  `json:"browser"`
	Device            string  `json:"device"`
	SystemLocale      string  `json:"system_locale"`
	BrowserUserAgent  string  `json:"browser_user_agent"`
	BrowserVersion    string  `json:"browser_version"`
	OsVersion         string  `json:"os_version"`
	Referrer          string  `json:"referrer"`
	ReferringDomain   string  `json:"referring_domain"`
	ReferrerCurrent   string  `json:"referrer_current"`
	ReferringDomainC  string  `json:"referring_domain_current"`
	ReleaseChannel    string  `json:"release_channel"`
	ClientBuildNumber int     `json:"client_build_number"`
	ClientEventSource *string `json:"client_event_source"`
}
type presence struct {
	Status     string     `json:"status"`
	Since      int        `json:"since"`
	Activities []Activity `json:"activities"`
	Afk        bool       `json:"afk"`
}
type initD struct {
	Token        string            `json:"token"`
	Capabilities int               `json:"capabilities"`
	Properties   browserProperties `json:"properties"`
	Presence     presence          `json:"presence"`
	Compress     bool              `json:"compress"`
	ClientState  struct {
		GuildVersions map[string]int `json:"guild_versions"`
	} `json:"client_state"`
}

type initializationData struct {
	Op int   `json:"op"`
	D  initD `json:"d"`
}

type externalAsset struct {
	Url               string `json:"url"`
	ExternalAssetPath string `json:"external_asset_path"`
}

type Activity struct {
	ApplicationId string   `json:"application_id"`
	Type          int      `json:"type"`
	Metadata      struct{} `json:"metadata"`
	Flags         int      `json:"flags"`
	State         string   `json:"state"`
	Details       string   `json:"details"`
	Instance      bool     `json:"instance"`
	Assets        struct {
		LargeImage string `json:"large_image"`
		SmallImage string `json:"small_image"`
		SmallText  string `json:"small_text"`
	} `json:"assets"`
	Timestamps struct {
		Start int64 `json:"start"`
		End   int64 `json:"end"`
	} `json:"timestamps"`
	Name string `json:"name"`
}

var externalAssetCache map[string]string

func init() {
	externalAssetCache = make(map[string]string)
}

func (a *Activity) FetchExternalAssets(token string) {
	if a.Assets.LargeImage == "" {
		a.Assets.LargeImage = "mp:external/MWL7VFP_X2f_X42rowD74F-GFt0J-E-fg_MzMOd3gPo/https/tmp.duti.dev/9lana.jpg"
		return
	}
	a.Assets.SmallImage = "mp:external/QoXN7gKtL2gQl37wtzCLCKv43y8WYtRiCbakNm2T5CU/https/i.imgur.com/8IYhOc2.png"
	if largeImage, ok := externalAssetCache[a.Assets.LargeImage]; ok {
		a.Assets.LargeImage = largeImage
		return
	}
	if strings.HasPrefix(a.Assets.SmallText, "https://www.youtube.com") {
		ytId := strings.TrimPrefix(a.Assets.SmallText, "https://www.youtube.com/watch?v=")
		a.Assets.LargeImage = fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", ytId)
	}
	reqBody := map[string][]string{
		"urls": {
			a.Assets.LargeImage,
		},
	}
	reqBodyJson, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://discord.com/api/v9/applications/%s/external-assets", a.ApplicationId), bytes.NewReader(reqBodyJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic("Failed to fetch external assets")
	}
	var externalAssets []externalAsset
	if err := json.NewDecoder(resp.Body).Decode(&externalAssets); err != nil {
		panic(err)
	}
	largeExternalAsset := "mp:" + externalAssets[0].ExternalAssetPath
	externalAssetCache[a.Assets.LargeImage] = largeExternalAsset
	a.Assets.LargeImage = largeExternalAsset
}

func InitData(token string) string {
	initData, _ := json.Marshal(initializationData{
		Op: 2,
		D: initD{
			Token:        token,
			Capabilities: 30717,
			Properties: browserProperties{
				Os:                "Linux",
				Browser:           "Firefox",
				SystemLocale:      "en-US",
				BrowserUserAgent:  "Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0",
				BrowserVersion:    "131.0",
				ReleaseChannel:    "stable",
				ClientBuildNumber: 337154,
			},
			Presence: presence{
				Status:     "unknown",
				Since:      0,
				Activities: []Activity{},
				Afk:        false,
			},
			ClientState: struct {
				GuildVersions map[string]int `json:"guild_versions"`
			}{GuildVersions: map[string]int{}},
			Compress: false,
		},
	})
	return string(initData)
}

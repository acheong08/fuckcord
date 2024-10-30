package internal

import (
	"bytes"
	"discordrpc/internal/token"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var externalAssetCache map[string]string

func init() {
	externalAssetCache = make(map[string]string)
}

func DefaultCover() string {
	return "mp:external/MWL7VFP_X2f_X42rowD74F-GFt0J-E-fg_MzMOd3gPo/https/tmp.duti.dev/9lana.jpg"
}

func FetchExternalAssets(a Activity) (string, error) {
	largeImage := a.Assets.LargeImage
	if strings.HasPrefix(a.Assets.SmallText, "https://www.youtube.com") {
		ytId := strings.TrimPrefix(a.Assets.SmallText, "https://www.youtube.com/watch?v=")
		largeImage = fmt.Sprintf("https://iv.duti.dev/vi/%s/hqdefault.jpg", ytId)
	} else if strings.HasSuffix(a.Assets.SmallText, ".m4a") {
		ytId := strings.TrimSuffix(a.Assets.SmallText, ".m4a")
		largeImage = fmt.Sprintf("https://iv.duti.dev/vi/%s/hqdefault.jpg", ytId)
	} else if largeImage == "" {
		return DefaultCover(), nil
	}
	if largeImage, ok := externalAssetCache[largeImage]; ok {
		return largeImage, nil
	}
	reqBody := map[string][]string{
		"urls": {
			largeImage,
		},
	}
	reqBodyJson, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://discord.com/api/v9/applications/%s/external-assets", a.ApplicationId), bytes.NewReader(reqBodyJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token.GetToken(""))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return DefaultCover(), err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return DefaultCover(), fmt.Errorf("failed to fetch external assets: %d", resp.StatusCode)
	}
	var externalAssets []externalAsset
	if err := json.NewDecoder(resp.Body).Decode(&externalAssets); err != nil {
		return DefaultCover(), err
	}
	largeExternalAsset := "mp:" + externalAssets[0].ExternalAssetPath
	externalAssetCache[largeImage] = largeExternalAsset
	return largeExternalAsset, nil
}

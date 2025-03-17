package coze

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coze-dev/coze-go"
	"github.com/songquanpeng/one-api/relay/adaptor/coze/constant/event"
)

func event2StopReason(e *string) string {
	if e == nil || *e == event.Message {
		return ""
	}
	return "stop"
}

func getOAuthToken(config string) (string, error) {
	var oauthConfig coze.OAuthConfig
	err := json.Unmarshal([]byte(config), &oauthConfig)
	if err != nil {
		return "", fmt.Errorf("failed to load OAuth config: %v", err)
	}

	oauth, err := coze.LoadOAuthAppFromConfig(&oauthConfig)
	if err != nil {
		return "", fmt.Errorf("failed to load OAuth config: %v", err)
	}

	jwtClient, ok := oauth.(*coze.JWTOAuthClient)
	if !ok {
		return "", fmt.Errorf("invalid OAuth client type: expected JWT client")
	}

	resp, err := jwtClient.GetAccessToken(context.TODO(), nil)

	if err != nil {
		return "", fmt.Errorf("failed to get access token: %v", err)
	}

	return resp.AccessToken, nil

}

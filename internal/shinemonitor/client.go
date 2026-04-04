package shinemonitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/config"
)

type Client struct {
	Config     *config.Config
	HTTPClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		Config:     cfg,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// MakeRequest handles authenticated requests
func (c *Client) MakeRequest(actionParams string) (map[string]interface{}, error) {
	token, secret, err := c.getToken()
	if err != nil {
		return nil, err
	}

	salt := generateSalt()

	// reqaction = str(salt) + secret + token + action_params
	reqaction := salt + secret + token + actionParams
	sign := signSHA1(reqaction)

	reqURL := fmt.Sprintf("http://web.shinemonitor.com/public/?sign=%s&salt=%s&token=%s%s", sign, salt, token, actionParams)

	resp, err := c.HTTPClient.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if errCode, ok := result["err"].(float64); ok && errCode != 0 {
		return result, fmt.Errorf("upstream error code: %v", errCode)
	}

	return result, nil
}

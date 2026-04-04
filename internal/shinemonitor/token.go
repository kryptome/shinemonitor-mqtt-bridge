package shinemonitor

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

type tokenResponse struct {
	Dat struct {
		Token  string `json:"token"`
		Secret string `json:"secret"`
		Expire int    `json:"expire"`
	} `json:"dat"`
	Err int `json:"err"`
}

func (c *Client) requestToken() (string, string, error) {
	salt := generateSalt()
	powSha1 := signSHA1(c.Config.Password)
	
	// Note: CompanyKey may need URL encoding if it contains special chars, but typically not required strictly string.
	// For exact python parity, user said: URL-encoding may be required for portal.
	action := fmt.Sprintf("&action=auth&usr=%s&company-key=%s", url.QueryEscape(c.Config.Username), url.QueryEscape(c.Config.CompanyKey))
	pwdaction := salt + powSha1 + action
	sign := signSHA1(pwdaction)

	reqURL := fmt.Sprintf("http://web.shinemonitor.com/public/?sign=%s&salt=%s%s", sign, salt, action)
	
	resp, err := c.HTTPClient.Get(reqURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	
	var res tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", "", err
	}
	
	if res.Err != 0 {
		return "", "", fmt.Errorf("auth error: %d", res.Err)
	}

	expireTime := time.Now().Add(time.Duration(res.Dat.Expire) * time.Second)
	
	// Create token file
	timeStr := expireTime.Format("2006-01-02 15:04:05.999999")
	err = os.WriteFile("token", []byte(fmt.Sprintf("%s\n%s\n%s", res.Dat.Token, res.Dat.Secret, timeStr)), 0644)
	return res.Dat.Token, res.Dat.Secret, err
}

func (c *Client) getToken() (string, string, error) {
	data, err := os.ReadFile("token")
	if err != nil {
		return c.requestToken()
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 3 {
		return c.requestToken()
	}
	
	expireMs, err := time.Parse("2006-01-02 15:04:05.999999", lines[2])
	if err != nil || time.Now().After(expireMs) {
		return c.requestToken()
	}
	
	return lines[0], lines[1], nil
}

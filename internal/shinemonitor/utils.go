package shinemonitor

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"
)

func generateSalt() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
}

func signSHA1(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func parseTimestamp(ts string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006-01",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, ts); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse timestamp: %s", ts)
}

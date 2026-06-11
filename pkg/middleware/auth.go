package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const GoogleClientID = "217072116416-ippggfugojfj5fiqk6oe0oiujjueak3m.apps.googleusercontent.com"

type googleTokenInfo struct {
	Sub   string `json:"sub"`
	Aud   string `json:"aud"`
	Error string `json:"error"`
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		info, err := validateGoogleToken(token)
		if err != nil || info.Sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		if info.Aud != GoogleClientID {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token audience mismatch"})
			return
		}

		c.Set("userID", info.Sub)
		c.Next()
	}
}

func validateGoogleToken(idToken string) (*googleTokenInfo, error) {
	resp, err := http.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK || info.Error != "" {
		return nil, fmt.Errorf("invalid token: %s", info.Error)
	}
	return &info, nil
}

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

const ctxClaimsKey = "auth.claims"

type AuthOptions struct {
	JWTKey     string
	APIURL     string
	APITimeout time.Duration
}

func Auth(opts AuthOptions) gin.HandlerFunc {
	httpClient := &http.Client{Timeout: opts.APITimeout}

	return func(c *gin.Context) {
		token, err := bearerToken(c.GetHeader("Authorization"))
		if err != nil {
			abortUnauthorized(c, err)
			return
		}

		var claims map[string]any
		switch {
		case opts.APIURL != "":
			claims, err = verifyRemote(c.Request.Context(), httpClient, opts.APIURL, token)
		case opts.JWTKey != "":
			claims, err = verifyJWT(token, opts.JWTKey)
		default:
			err = errors.New("auth misconfigured: no JWT key or API URL")
		}
		if err != nil {
			log.Debug().Err(err).Msg("auth rejected")
			abortUnauthorized(c, err)
			return
		}

		c.Set(ctxClaimsKey, claims)
		c.Next()
	}
}

func Claims(c *gin.Context) (map[string]any, bool) {
	v, ok := c.Get(ctxClaimsKey)
	if !ok {
		return nil, false
	}
	claims, ok := v.(map[string]any)
	return claims, ok
}

func bearerToken(header string) (string, error) {
	if header == "" {
		return "", errors.New("missing Authorization header")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("Authorization header must be Bearer")
	}
	tok := strings.TrimSpace(header[len(prefix):])
	if tok == "" {
		return "", errors.New("empty bearer token")
	}
	return tok, nil
}

func verifyJWT(tokenStr, key string) (map[string]any, error) {
	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(key), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func verifyRemote(ctx context.Context, client *http.Client, url, token string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("auth API rejected token: " + resp.Status)
	}

	var claims map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func abortUnauthorized(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
}

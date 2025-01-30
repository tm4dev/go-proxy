package main

import (
	"encoding/base64"
	"errors"
	"strings"
)

type ProxyAuth struct {
	Username   string
	Password   string
	Parameters map[string]string
}

func NewProxyAuth() *ProxyAuth {
	return &ProxyAuth{
		Username:   "",
		Password:   "",
		Parameters: make(map[string]string),
	}
}

func parseProxyAuth(auth string) (*ProxyAuth, error) {
	if !strings.HasPrefix(auth, "Basic ") {
		return nil, errors.New("invalid proxy authentication")
	}

	auth = strings.TrimPrefix(auth, "Basic ")

	authDecoded := make([]byte, base64.StdEncoding.DecodedLen(len(auth)))
	n, err := base64.StdEncoding.Decode(authDecoded, []byte(auth))
	if err != nil {
		return nil, err
	}

	authDecoded = authDecoded[:n]

	username, password, ok := strings.Cut(string(authDecoded), ":")
	if !ok {
		return nil, errors.New("invalid proxy authentication")
	}

	parameters := make(map[string]string)
	parts := strings.Split(username, "-")
	if len(parts) > 1 {
		username = parts[0]
		for i := 1; i < len(parts)-1; i += 2 {
			if i+1 < len(parts) {
				parameters[parts[i]] = parts[i+1]
			}
		}
	}

	return &ProxyAuth{
		Username:   username,
		Password:   password,
		Parameters: parameters,
	}, nil
}

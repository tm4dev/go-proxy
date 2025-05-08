package auth

import (
	"encoding/base64"
	"regexp"
	"strings"

	"github.com/vlourme/go-proxy/internal/http"
)

const (
	ParamSession  string = "session"
	ParamTimeout  string = "timeout"
	ParamLocation string = "country"
	ParamFallback string = "fallback"
)

var sessionRegex = regexp.MustCompile(`^[a-zA-Z]+$`)

// GetCredentials returns the username, password and params from the Proxy-Authorization header
func GetCredentials(req *http.Request) (string, string, string) {
	auth := string(req.Header["Proxy-Authorization"])
	if auth == "" {
		return "", "", ""
	}

	auth = strings.TrimPrefix(auth, "Basic ")
	if auth == "" {
		return "", "", ""
	}

	decoded, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return "", "", ""
	}

	username, password, _ := strings.Cut(string(decoded), ":")
	username, params, _ := strings.Cut(username, "-")
	return username, password, params
}

// GetParams returns a map of the parameters from the username
func GetParams(params string) map[string]string {
	if params == "" {
		return make(map[string]string)
	}

	result := make(map[string]string)
	parts := strings.Split(params, "-")

	for i := 0; i < len(parts)-1; i += 2 {
		if i+1 < len(parts) {
			key := parts[i]
			value := parts[i+1]
			result[key] = value
		}
	}

	if !VerifySession(result) {
		delete(result, ParamSession)
	}

	return result
}

// VerifySession verifies the session
func VerifySession(result map[string]string) bool {
	session := result[ParamSession]
	if len(session) < 6 || len(session) > 24 {
		return false
	}

	if !sessionRegex.MatchString(session) {
		return false
	}

	return true
}

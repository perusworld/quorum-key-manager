package request

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
)

// BasicAuth sets Authorization header "Basic <username>:<password>"
// If header is already set then does nothing
// if User is set on request URL then set username, password to the values of the URL
func BasicAuth(cfg *BasicAuthConfig) Preparer {
	if cfg == nil {
		return NoopPreparer
	}

	return PrepareFunc(func(req *http.Request) (*http.Request, error) {
		if req.Header.Get("Authorization") != "" {
			// If Authorization Header is already set then do not alter request
			return req, nil
		}

		u := req.URL.User
		if u == nil && cfg.Username != "" && cfg.Password != "" {
			// If user is not set and username/password are valid
			// then populates users
			u = url.UserPassword(cfg.Username, cfg.Password)
			req.URL.User = u
		}

		if u != nil {
			// If user has been set then set Authorization Header with corresponding Basic Authorization header
			username := u.Username()
			password, _ := u.Password()
			req.Header.Set("Authorization", fmt.Sprintf("Basic %v", basicAuth(username, password)))
		}

		return req, nil
	})
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

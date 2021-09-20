package client

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/vault/api"
	"golang.org/x/net/http2"
	"golang.org/x/time/rate"
)

// Config object that be converted into an api.Config later
type Config struct {
	Address       string
	CACert        string
	CAPath        string
	ClientCert    string
	ClientKey     string
	TLSServerName string
	Namespace     string
	ClientTimeout time.Duration
	RateLimit     float64
	BurstLimit    int
	MaxRetries    int
	SkipVerify    bool
	Token         string
}

func NewConfig(addr, namespace string) *Config {
	return &Config{
		Address:   addr,
		Namespace: namespace,
	}
}

// ToHashicorpConfig extracts an api.Config object from self
func (c *Config) ToHashicorpConfig() *api.Config {
	// Create Hashicorp Configuration
	config := api.DefaultConfig()
	config.Address = c.Address
	config.HttpClient = cleanhttp.DefaultClient()
	config.HttpClient.Timeout = time.Second * 60

	// Create Transport
	transport := config.HttpClient.Transport.(*http.Transport)
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	if err := http2.ConfigureTransport(transport); err != nil {
		config.Error = err
		return config
	}

	// Configure TLS
	tlsConfig := &api.TLSConfig{
		CACert:        c.CACert,
		CAPath:        c.CAPath,
		ClientCert:    c.ClientCert,
		ClientKey:     c.ClientKey,
		TLSServerName: c.TLSServerName,
		Insecure:      c.SkipVerify,
	}

	_ = config.ConfigureTLS(tlsConfig)

	config.Limiter = rate.NewLimiter(rate.Limit(c.RateLimit), c.BurstLimit)
	config.MaxRetries = c.MaxRetries
	config.Timeout = c.ClientTimeout

	// Ensure redirects are not automatically followed
	// Note that this is sane for the API client as it has its own
	// redirect handling logic (and thus also for command/meta),
	// but in e.g. http_test actual redirect handling is necessary
	config.HttpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// Returning this value causes the Go net library to not close the
		// response body and to nil out the error. Otherwise, retry clients may
		// try three times on every redirect because it sees an error from this
		// function (to prevent redirects) passing through to it.
		return http.ErrUseLastResponse
	}

	config.Backoff = retryablehttp.LinearJitterBackoff

	return config
}

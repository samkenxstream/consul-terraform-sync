package client

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/hashicorp/consul-terraform-sync/config"
	"github.com/hashicorp/consul-terraform-sync/logging"
	"github.com/hashicorp/consul-terraform-sync/retry"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcat"
)

const (
	ConsulEnterpriseSKU   = "ent"
	ConsulOSSSKU          = "oss"
	ConsulDefaultMaxRetry = 8 // to be consistent with hcat retries
	consulSubsystemName   = "consul"
)

//go:generate mockery --name=ConsulClientInterface --filename=consul_client.go --output=../mocks/client --tags=enterprise

var _ ConsulClientInterface = (*ConsulClient)(nil)

// ConsulClientInterface is an interface for a Consul Client
// If more consul client functionality is required, this interface should be extended with the following
// considerations:
// Each request to Consul is:
// - Retried
// - Logged at DEBUG-level
// - Easily mocked
type ConsulClientInterface interface {
	GetLicense(ctx context.Context, q *consulapi.QueryOptions) (string, error)
	GetSKU(ctx context.Context) (string, error)
}

// ConsulClient is a client to the Consul API
type ConsulClient struct {
	*consulapi.Client
	retry  retry.Retry
	logger logging.Logger
}

// Self represents the response body from Consul /v1/agent/self API endpoint.
// Care must always be taken to do type checks when casting, as structure could
// potentially change over time.
type Self = map[string]map[string]interface{}

// NewConsulClient constructs a consul api client
func NewConsulClient(conf *config.Config, maxRetry int) (*ConsulClient, error) {
	consulConf := conf.Consul
	transport := hcat.TransportInput{
		SSLEnabled: *consulConf.TLS.Enabled,
		SSLVerify:  *consulConf.TLS.Verify,
		SSLCert:    *consulConf.TLS.Cert,
		SSLKey:     *consulConf.TLS.Key,
		SSLCACert:  *consulConf.TLS.CACert,
		SSLCAPath:  *consulConf.TLS.CAPath,
		ServerName: *consulConf.TLS.ServerName,

		DialKeepAlive:       *consulConf.Transport.DialKeepAlive,
		DialTimeout:         *consulConf.Transport.DialTimeout,
		DisableKeepAlives:   *consulConf.Transport.DisableKeepAlives,
		IdleConnTimeout:     *consulConf.Transport.IdleConnTimeout,
		MaxIdleConns:        *consulConf.Transport.MaxIdleConns,
		MaxIdleConnsPerHost: *consulConf.Transport.MaxIdleConnsPerHost,
		TLSHandshakeTimeout: *consulConf.Transport.TLSHandshakeTimeout,
	}

	consul := hcat.ConsulInput{
		Address:      *consulConf.Address,
		Token:        *consulConf.Token,
		AuthEnabled:  *consulConf.Auth.Enabled,
		AuthUsername: *consulConf.Auth.Username,
		AuthPassword: *consulConf.Auth.Password,
		Transport:    transport,
	}

	clients := hcat.NewClientSet()
	if err := clients.AddConsul(consul); err != nil {
		return nil, err
	}

	logger := logging.Global().Named(loggingSystemName).Named(consulSubsystemName)

	r := retry.NewRetry(maxRetry, time.Now().UnixNano())
	return &ConsulClient{
		Client: clients.Consul(),
		retry:  r,
		logger: logger}, nil
}

// GetLicense queries Consul for a signed license, and returns it if available
func (c *ConsulClient) GetLicense(ctx context.Context, q *consulapi.QueryOptions) (string, error) {
	c.logger.Debug("getting license")

	desc := "consul client get license"
	var err error
	var license string

	err = c.retry.Do(ctx, func(context.Context) error {
		license, err = c.Operator().LicenseGetSigned(q)
		if err != nil {
			license = ""
			return err
		}
		return nil
	}, desc)

	return license, err
}

// GetSKU queries Consul for information about itself, it then
// parses this information to determine whether the Consul being
// queried is Enterprise or OSS
func (c *ConsulClient) GetSKU(ctx context.Context) (string, error) {
	c.logger.Debug("getting sku")

	desc := "consul client get sku"
	var err error
	var info Self

	err = c.retry.Do(ctx, func(context.Context) error {
		info, err = c.Agent().Self()
		if err != nil {
			info = nil
			return err
		}
		return nil
	}, desc)
	if err != nil {
		return "", err
	}

	sku, ok := parseSKU(info)
	if !ok {
		return "", errors.New("unable to parse sku")
	}
	return sku, nil
}

func parseSKU(info Self) (string, bool) {
	v, ok := info["Config"]["Version"].(string)
	if !ok {
		return "", ok
	}

	ver, vErr := version.NewVersion(v)
	if vErr != nil {
		return "", false
	}
	if strings.Contains(ver.Metadata(), ConsulEnterpriseSKU) {
		return ConsulEnterpriseSKU, true
	}
	return ConsulOSSSKU, true
}

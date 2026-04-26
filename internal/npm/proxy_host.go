package npm

import (
	"context"
	"fmt"
	"slices"

	"github.com/tarach/npm-docker-auto-proxy/internal/proxy"
)

type ProxyHost struct {
	ID                    int               `json:"id,omitempty"`
	DomainNames           []string          `json:"domain_names"`
	ForwardHost           string            `json:"forward_host"`
	ForwardPort           int               `json:"forward_port"`
	AccessListID          int               `json:"access_list_id"`
	CertificateID         int               `json:"certificate_id"`
	SSLForced             bool              `json:"ssl_forced"`
	CachingEnabled        bool              `json:"caching_enabled"`
	BlockExploits         bool              `json:"block_exploits"`
	AdvancedConfig        string            `json:"advanced_config"`
	Meta                  map[string]any     `json:"meta"`
	AllowWebsocketUpgrade bool              `json:"allow_websocket_upgrade"`
	HTTP2Support          bool              `json:"http2_support"`
	ForwardScheme         string            `json:"forward_scheme"`
	Enabled               bool              `json:"enabled"`
	Locations             []any             `json:"locations"`
	HSTSEnabled           bool              `json:"hsts_enabled"`
	HSTSSubdomains        bool              `json:"hsts_subdomains"`
	TrustForwardedProto   bool              `json:"trust_forwarded_proto"`
}

type ChangeSet []string

func (c *Client) ListProxyHosts(ctx context.Context) ([]ProxyHost, error) {
	var hosts []ProxyHost
	err := c.doNoBody(ctx, "GET", "/nginx/proxy-hosts", &hosts)
	return hosts, err
}

func (c *Client) FindProxyHostByDomain(ctx context.Context, domain string) (ProxyHost, bool, error) {
	hosts, err := c.ListProxyHosts(ctx)
	if err != nil {
		return ProxyHost{}, false, err
	}

	for _, host := range hosts {
		if slices.Contains(host.DomainNames, domain) {
			return host, true, nil
		}
	}

	return ProxyHost{}, false, nil
}

func (c *Client) CreateProxyHost(ctx context.Context, desired proxy.DesiredHost) (ProxyHost, error) {
	var host ProxyHost
	err := c.doJSON(ctx, "POST", "/nginx/proxy-hosts", payloadFromDesired(desired), &host)
	return host, err
}

func (c *Client) UpdateProxyHost(ctx context.Context, id int, desired proxy.DesiredHost) (ProxyHost, error) {
	var host ProxyHost
	path := fmt.Sprintf("/nginx/proxy-hosts/%d", id)
	err := c.doJSON(ctx, "PUT", path, payloadFromDesired(desired), &host)
	return host, err
}

func (c *Client) EnableProxyHost(ctx context.Context, id int) error {
	path := fmt.Sprintf("/nginx/proxy-hosts/%d/enable", id)
	return c.doJSON(ctx, "POST", path, map[string]any{}, nil)
}

func (c *Client) DisableProxyHost(ctx context.Context, id int) error {
	path := fmt.Sprintf("/nginx/proxy-hosts/%d/disable", id)
	return c.doJSON(ctx, "POST", path, map[string]any{}, nil)
}

func (c *Client) DeleteProxyHost(ctx context.Context, id int) error {
	path := fmt.Sprintf("/nginx/proxy-hosts/%d", id)
	return c.doNoBody(ctx, "DELETE", path, nil)
}

func Differences(existing ProxyHost, desired proxy.DesiredHost) ChangeSet {
	changes := ChangeSet{}

	checks := []struct {
		name    string
		changed bool
	}{
		{name: "domain_names", changed: !slices.Contains(existing.DomainNames, desired.Domain)},
		{name: "forward_host", changed: existing.ForwardHost != desired.ForwardHost},
		{name: "forward_port", changed: existing.ForwardPort != desired.ForwardPort},
		{name: "forward_scheme", changed: existing.ForwardScheme != desired.ForwardScheme},
		{name: "certificate_id", changed: existing.CertificateID != desired.CertificateID},
		{name: "ssl_forced", changed: existing.SSLForced != desired.SSLForced},
		{name: "allow_websocket_upgrade", changed: existing.AllowWebsocketUpgrade != desired.AllowWebsocketUpgrade},
		{name: "http2_support", changed: existing.HTTP2Support != desired.HTTP2Support},
		{name: "block_exploits", changed: existing.BlockExploits != desired.BlockExploits},
	}

	for _, check := range checks {
		if check.changed {
			changes = append(changes, check.name)
		}
	}

	return changes
}

func payloadFromDesired(desired proxy.DesiredHost) ProxyHost {
	return ProxyHost{
		DomainNames:           []string{desired.Domain},
		ForwardScheme:         desired.ForwardScheme,
		ForwardHost:           desired.ForwardHost,
		ForwardPort:           desired.ForwardPort,
		AccessListID:          desired.AccessListID,
		CertificateID:         desired.CertificateID,
		SSLForced:             desired.SSLForced,
		CachingEnabled:        desired.CachingEnabled,
		BlockExploits:         desired.BlockExploits,
		AllowWebsocketUpgrade: desired.AllowWebsocketUpgrade,
		HTTP2Support:          desired.HTTP2Support,
		HSTSEnabled:           desired.HSTSEnabled,
		HSTSSubdomains:        desired.HSTSSubdomains,
		Locations:             []any{},
		AdvancedConfig:        desired.AdvancedConfig,
		Enabled:               true,
		Meta: map[string]any{
			"letsencrypt_agree": false,
			"dns_challenge":    false,
		},
	}
}

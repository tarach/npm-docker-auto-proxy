package proxy

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	LabelEnabled       = "npm.proxy.enabled"
	LabelDomain        = "npm.proxy.domain"
	LabelForwardHost   = "npm.proxy.forward_host"
	LabelForwardPort   = "npm.proxy.forward_port"
	LabelScheme        = "npm.proxy.scheme"
	LabelWebsocket     = "npm.proxy.websocket"
	LabelSSL           = "npm.proxy.ssl"
	LabelForceSSL      = "npm.proxy.force_ssl"
	LabelBlockExploits = "npm.proxy.block_exploits"
	LabelHTTP2         = "npm.proxy.http2"
	LabelOnStop        = "npm.proxy.on_stop"
)

type DesiredHost struct {
	Enabled               bool
	Domain                string
	ForwardHost           string
	ForwardPort           int
	ForwardScheme         string
	AllowWebsocketUpgrade bool
	CertificateID         int
	SSLForced             bool
	HTTP2Support          bool
	BlockExploits         bool
	CachingEnabled        bool
	HSTSEnabled           bool
	HSTSSubdomains        bool
	AccessListID          int
	AdvancedConfig        string
	OnStop                StopAction
}

func FromLabels(labels map[string]string) (DesiredHost, error) {
	enabled := parseBool(labels[LabelEnabled])
	if !enabled {
		return DesiredHost{Enabled: false}, nil
	}

	domain := strings.TrimSpace(labels[LabelDomain])
	if domain == "" {
		return DesiredHost{}, fmt.Errorf("%s is required", LabelDomain)
	}

	forwardHost := strings.TrimSpace(labels[LabelForwardHost])
	if forwardHost == "" {
		return DesiredHost{}, fmt.Errorf("%s is required", LabelForwardHost)
	}

	forwardPort, err := strconv.Atoi(strings.TrimSpace(labels[LabelForwardPort]))
	if err != nil {
		return DesiredHost{}, fmt.Errorf("%s must be a number", LabelForwardPort)
	}

	scheme := strings.TrimSpace(labels[LabelScheme])
	if scheme == "" {
		scheme = "http"
	}

	host := DesiredHost{
		Enabled:               true,
		Domain:                domain,
		ForwardHost:           forwardHost,
		ForwardPort:           forwardPort,
		ForwardScheme:         scheme,
		AllowWebsocketUpgrade: parseBool(labels[LabelWebsocket]),
		CertificateID:         resolveCertificateID(labels),
		SSLForced:             parseBool(labels[LabelForceSSL]),
		HTTP2Support:          parseBoolDefault(labels[LabelHTTP2], true),
		BlockExploits:         parseBoolDefault(labels[LabelBlockExploits], true),
		CachingEnabled:        false,
		HSTSEnabled:           false,
		HSTSSubdomains:        false,
		AccessListID:          0,
		AdvancedConfig:        "",
		OnStop:                ResolveStopAction(labels[LabelOnStop]),
	}

	return host, nil
}

func parseBool(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))

	trueValues := map[string]bool{
		"1":    true,
		"true": true,
		"yes":  true,
		"on":   true,
	}

	return trueValues[normalized]
}

func parseBoolDefault(value string, defaultValue bool) bool {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return defaultValue
	}

	return parseBool(normalized)
}

func resolveCertificateID(labels map[string]string) int {
	if !parseBool(labels[LabelSSL]) {
		return 0
	}

	value := strings.TrimSpace(labels["npm.proxy.certificate_id"])
	if value == "" {
		return 0
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}

	return id
}

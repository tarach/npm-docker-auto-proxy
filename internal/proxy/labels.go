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
	LabelCertificate   = "npm.proxy.certificate"
	LabelCertificateID = "npm.proxy.certificate_id"
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
	CertificateRef        string
	CertificateID         int
	SSLEnabled            bool
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

	err = validateSSLLabels(labels)
	if err != nil {
		return DesiredHost{}, err
	}

	certificateID, err := resolveCertificateID(labels)
	if err != nil {
		return DesiredHost{}, err
	}

	host := DesiredHost{
		Enabled:               true,
		Domain:                domain,
		ForwardHost:           forwardHost,
		ForwardPort:           forwardPort,
		ForwardScheme:         scheme,
		AllowWebsocketUpgrade: parseBool(labels[LabelWebsocket]),
		CertificateRef:        strings.TrimSpace(labels[LabelCertificate]),
		CertificateID:         certificateID,
		SSLEnabled:            parseBool(labels[LabelSSL]),
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

func validateSSLLabels(labels map[string]string) error {
	sslEnabled := parseBool(labels[LabelSSL])
	forceSSL := parseBool(labels[LabelForceSSL])
	certificateRef := strings.TrimSpace(labels[LabelCertificate])
	certificateID := strings.TrimSpace(labels[LabelCertificateID])

	if forceSSL && !sslEnabled {
		return fmt.Errorf("%s=true requires %s=true", LabelForceSSL, LabelSSL)
	}

	if !sslEnabled {
		return nil
	}

	if certificateRef != "" {
		return nil
	}

	if certificateID != "" {
		return nil
	}

	return fmt.Errorf("%s=true requires %s or %s", LabelSSL, LabelCertificate, LabelCertificateID)
}

func resolveCertificateID(labels map[string]string) (int, error) {
	value := strings.TrimSpace(labels[LabelCertificateID])
	if value == "" {
		return 0, nil
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", LabelCertificateID)
	}

	if id <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", LabelCertificateID)
	}

	return id, nil
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

package proxy

import (
	"fmt"
	"strconv"
	"strings"
)

const DefaultLabelsPrefix = "npm.proxy."

type Labels struct {
	Prefix           string
	Enabled          string
	Domain           string
	ForwardHost      string
	ForwardPort      string
	Scheme           string
	Websocket        string
	SSL              string
	Certificate      string
	CertificateID    string
	ForceSSL         string
	BlockExploits    string
	HTTP2            string
	OnStop           string
}

func NewLabels(prefix string) Labels {
	if prefix == "" {
		prefix = DefaultLabelsPrefix
	}

	if !strings.HasSuffix(prefix, ".") {
		prefix += "."
	}

	return Labels{
		Prefix:        prefix,
		Enabled:       prefix + "enabled",
		Domain:        prefix + "domain",
		ForwardHost:   prefix + "forward_host",
		ForwardPort:   prefix + "forward_port",
		Scheme:        prefix + "scheme",
		Websocket:     prefix + "websocket",
		SSL:           prefix + "ssl",
		Certificate:   prefix + "certificate",
		CertificateID: prefix + "certificate_id",
		ForceSSL:      prefix + "force_ssl",
		BlockExploits: prefix + "block_exploits",
		HTTP2:         prefix + "http2",
		OnStop:        prefix + "on_stop",
	}
}

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

func FromLabels(labels map[string]string, labelNames Labels) (DesiredHost, error) {
	enabled := parseBool(labels[labelNames.Enabled])
	if !enabled {
		return DesiredHost{Enabled: false}, nil
	}

	domain := strings.TrimSpace(labels[labelNames.Domain])
	if domain == "" {
		return DesiredHost{}, fmt.Errorf("%s is required", labelNames.Domain)
	}

	forwardHost := strings.TrimSpace(labels[labelNames.ForwardHost])
	if forwardHost == "" {
		return DesiredHost{}, fmt.Errorf("%s is required", labelNames.ForwardHost)
	}

	forwardPort, err := strconv.Atoi(strings.TrimSpace(labels[labelNames.ForwardPort]))
	if err != nil {
		return DesiredHost{}, fmt.Errorf("%s must be a number", labelNames.ForwardPort)
	}

	scheme := strings.TrimSpace(labels[labelNames.Scheme])
	if scheme == "" {
		scheme = "http"
	}

	err = validateSSLLabels(labels, labelNames)
	if err != nil {
		return DesiredHost{}, err
	}

	certificateID, err := resolveCertificateID(labels, labelNames)
	if err != nil {
		return DesiredHost{}, err
	}

	host := DesiredHost{
		Enabled:               true,
		Domain:                domain,
		ForwardHost:           forwardHost,
		ForwardPort:           forwardPort,
		ForwardScheme:         scheme,
		AllowWebsocketUpgrade: parseBool(labels[labelNames.Websocket]),
		CertificateRef:        strings.TrimSpace(labels[labelNames.Certificate]),
		CertificateID:         certificateID,
		SSLEnabled:            parseBool(labels[labelNames.SSL]),
		SSLForced:             parseBool(labels[labelNames.ForceSSL]),
		HTTP2Support:          parseBoolDefault(labels[labelNames.HTTP2], true),
		BlockExploits:         parseBoolDefault(labels[labelNames.BlockExploits], true),
		CachingEnabled:        false,
		HSTSEnabled:           false,
		HSTSSubdomains:        false,
		AccessListID:          0,
		AdvancedConfig:        "",
		OnStop:                ResolveStopAction(labels[labelNames.OnStop]),
	}

	return host, nil
}

func validateSSLLabels(labels map[string]string, labelNames Labels) error {
	sslEnabled := parseBool(labels[labelNames.SSL])
	forceSSL := parseBool(labels[labelNames.ForceSSL])
	certificateRef := strings.TrimSpace(labels[labelNames.Certificate])
	certificateID := strings.TrimSpace(labels[labelNames.CertificateID])

	if forceSSL && !sslEnabled {
		return fmt.Errorf("%s=true requires %s=true", labelNames.ForceSSL, labelNames.SSL)
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

	return fmt.Errorf("%s=true requires %s or %s", labelNames.SSL, labelNames.Certificate, labelNames.CertificateID)
}

func resolveCertificateID(labels map[string]string, labelNames Labels) (int, error) {
	value := strings.TrimSpace(labels[labelNames.CertificateID])
	if value == "" {
		return 0, nil
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", labelNames.CertificateID)
	}

	if id <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", labelNames.CertificateID)
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

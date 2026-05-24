package npm

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type Certificate struct {
	ID          int      `json:"id"`
	Provider    string   `json:"provider"`
	NiceName    string   `json:"nice_name"`
	DomainNames []string `json:"domain_names"`
	ExpiresOn   string   `json:"expires_on"`
}

func (c *Client) ListCertificates(ctx context.Context) ([]Certificate, error) {
	var certificates []Certificate

	err := c.doNoBody(ctx, "GET", "/nginx/certificates", &certificates)

	return certificates, err
}

func (c *Client) ResolveCertificateID(ctx context.Context, ref string) (int, error) {
	value := strings.TrimSpace(ref)
	if value == "" {
		return 0, fmt.Errorf("certificate reference is required")
	}

	id, err := strconv.Atoi(value)
	if err == nil && id > 0 {
		return id, nil
	}

	certificates, err := c.ListCertificates(ctx)
	if err != nil {
		return 0, err
	}

	for _, certificate := range certificates {
		if certificateMatches(certificate, value) {
			return certificate.ID, nil
		}
	}

	return 0, fmt.Errorf("certificate not found: %s", value)
}

func certificateMatches(certificate Certificate, value string) bool {
	if strings.EqualFold(certificate.NiceName, value) {
		return true
	}

	if slices.Contains(certificate.DomainNames, value) {
		return true
	}

	return false
}

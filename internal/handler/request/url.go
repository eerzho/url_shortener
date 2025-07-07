package request

import (
	"errors"
	"net/url"
	"strings"
	"url_shortener/internal/constant"
)

// CreateUrl represents the request payload for creating a short URL
type CreateUrl struct {
	LongUrl string `json:"long_url" validate:"required,url,min=10,max=2048"`
}

// Validate performs custom validation on the CreateUrl request
func (c *CreateUrl) Validate() error {
	// Trim whitespace
	c.LongUrl = strings.TrimSpace(c.LongUrl)

	// Check if URL is empty after trimming
	if c.LongUrl == "" {
		return errors.New("long_url is required")
	}

	// Check length constraints
	if len(c.LongUrl) < constant.MinURLLength {
		return errors.New("long_url is too short")
	}

	if len(c.LongUrl) > constant.MaxURLLength {
		return errors.New("long_url is too long")
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(c.LongUrl)
	if err != nil {
		return errors.New("long_url is not a valid URL")
	}

	// Check if scheme is present and valid
	if parsedURL.Scheme == "" {
		return errors.New("long_url must include a scheme (http:// or https://)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("long_url must use http or https scheme")
	}

	// Check if host is present
	if parsedURL.Host == "" {
		return errors.New("long_url must include a valid host")
	}

	// Check for localhost or private IP ranges in production
	if isLocalhost(parsedURL.Host) {
		return errors.New("long_url cannot point to localhost or private networks")
	}

	// Normalize the URL
	c.LongUrl = parsedURL.String()

	return nil
}

// isLocalhost checks if the host is localhost or a private IP
func isLocalhost(host string) bool {
	// Remove port if present
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// Check for localhost variants
	localhostVariants := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"0.0.0.0",
	}

	for _, variant := range localhostVariants {
		if strings.EqualFold(host, variant) {
			return true
		}
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.",
		"192.168.",
		"172.16.", "172.17.", "172.18.", "172.19.", "172.20.",
		"172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
	}

	for _, privateRange := range privateRanges {
		if strings.HasPrefix(host, privateRange) {
			return true
		}
	}

	return false
}

// GetShortCodeRequest represents the request for getting URL info by short code
type GetShortCodeRequest struct {
	ShortCode string `json:"short_code" validate:"required,min=1,max=20"`
}

// Validate performs custom validation on the GetShortCodeRequest
func (g *GetShortCodeRequest) Validate() error {
	// Trim whitespace
	g.ShortCode = strings.TrimSpace(g.ShortCode)

	// Check if short code is empty after trimming
	if g.ShortCode == "" {
		return errors.New("short_code is required")
	}

	// Check length constraints
	if len(g.ShortCode) > constant.DefaultShortCodeLength*2 {
		return errors.New("short_code is too long")
	}

	// Check for valid characters
	for _, char := range g.ShortCode {
		if !isValidShortCodeChar(char) {
			return errors.New("short_code contains invalid characters")
		}
	}

	return nil
}

// isValidShortCodeChar checks if a character is valid for short codes
func isValidShortCodeChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}

// RedirectRequest represents the request for URL redirection
type RedirectRequest struct {
	ShortCode string `json:"short_code" validate:"required,min=1,max=20"`
}

// Validate performs custom validation on the RedirectRequest
func (r *RedirectRequest) Validate() error {
	// Trim whitespace
	r.ShortCode = strings.TrimSpace(r.ShortCode)

	// Check if short code is empty after trimming
	if r.ShortCode == "" {
		return errors.New("short_code is required")
	}

	// Check length constraints
	if len(r.ShortCode) > constant.DefaultShortCodeLength*2 {
		return errors.New("short_code is too long")
	}

	// Check for valid characters
	for _, char := range r.ShortCode {
		if !isValidShortCodeChar(char) {
			return errors.New("short_code contains invalid characters")
		}
	}

	return nil
}

// StatsRequest represents the request for getting statistics
type StatsRequest struct {
	Limit  int    `json:"limit" validate:"min=1,max=100"`
	Offset int    `json:"offset" validate:"min=0"`
	Period string `json:"period" validate:"omitempty,oneof=day week month year"`
}

// Validate performs custom validation on the StatsRequest
func (s *StatsRequest) Validate() error {
	// Set default values
	if s.Limit == 0 {
		s.Limit = 10
	}

	if s.Offset < 0 {
		s.Offset = 0
	}

	// Validate limit
	if s.Limit > 100 {
		return errors.New("limit cannot exceed 100")
	}

	// Validate period
	if s.Period != "" {
		validPeriods := map[string]bool{
			"day":   true,
			"week":  true,
			"month": true,
			"year":  true,
		}

		if !validPeriods[s.Period] {
			return errors.New("period must be one of: day, week, month, year")
		}
	}

	return nil
}

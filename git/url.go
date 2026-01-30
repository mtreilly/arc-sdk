// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package git

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ParsedURL contains parsed git URL information.
type ParsedURL struct {
	URL      string
	Platform string
	Owner    string
	Name     string
	Host     string
}

var (
	// sshURLRegex matches git@host:owner/repo.git or git@host:owner/repo
	sshURLRegex = regexp.MustCompile(`^git@([^:]+):([^/]+)/([^/.]+?)(?:\.git)?$`)

	// httpURLRegex matches https://host/owner/repo or https://host/owner/repo.git
	httpURLRegex = regexp.MustCompile(`^https?://([^/]+)/([^/]+)/([^/.]+?)(?:\.git)?/?$`)
)

// ParseGitURL parses a git URL and extracts platform, owner, and repo name.
func ParseGitURL(rawURL string) (*ParsedURL, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("empty URL")
	}

	// Try SSH format first
	if matches := sshURLRegex.FindStringSubmatch(rawURL); matches != nil {
		return &ParsedURL{
			URL:      rawURL,
			Host:     matches[1],
			Platform: detectPlatform(matches[1]),
			Owner:    matches[2],
			Name:     matches[3],
		}, nil
	}

	// Try HTTP format
	if matches := httpURLRegex.FindStringSubmatch(rawURL); matches != nil {
		return &ParsedURL{
			URL:      rawURL,
			Host:     matches[1],
			Platform: detectPlatform(matches[1]),
			Owner:    matches[2],
			Name:     matches[3],
		}, nil
	}

	// Try parsing as URL for more complex cases
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("URL must contain owner and repo: %s", rawURL)
	}

	name := parts[1]
	name = strings.TrimSuffix(name, ".git")

	return &ParsedURL{
		URL:      rawURL,
		Host:     u.Host,
		Platform: detectPlatform(u.Host),
		Owner:    parts[0],
		Name:     name,
	}, nil
}

func detectPlatform(host string) string {
	host = strings.ToLower(host)
	switch {
	case strings.Contains(host, "github"):
		return "github"
	case strings.Contains(host, "gitlab"):
		return "gitlab"
	case strings.Contains(host, "bitbucket"):
		return "bitbucket"
	case strings.Contains(host, "codeberg"):
		return "codeberg"
	case strings.Contains(host, "sourcehut") || strings.Contains(host, "sr.ht"):
		return "sourcehut"
	default:
		return "git"
	}
}

// FullName returns "owner/name" format.
func (p *ParsedURL) FullName() string {
	return p.Owner + "/" + p.Name
}

// CloneURL returns the HTTPS clone URL.
func (p *ParsedURL) CloneURL() string {
	return fmt.Sprintf("https://%s/%s/%s.git", p.Host, p.Owner, p.Name)
}

// WebURL returns the web URL for browsing.
func (p *ParsedURL) WebURL() string {
	return fmt.Sprintf("https://%s/%s/%s", p.Host, p.Owner, p.Name)
}

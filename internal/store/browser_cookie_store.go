package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrBrowserCookieEncryptionRequired = errors.New("browser cookie encryption key required")

// BrowserCookieScope is the tenant/user/agent boundary for synced browser cookies.
type BrowserCookieScope struct {
	TenantID uuid.UUID
	UserID   string
	AgentID  string
}

// BrowserCookie stores one selected browser cookie. Value is plaintext only at API
// and browser-application boundaries; stores encrypt it at rest.
type BrowserCookie struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	TenantID  uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	AgentID   string     `json:"agent_id" db:"agent_id"`
	Domain    string     `json:"domain" db:"domain"`
	Name      string     `json:"name" db:"name"`
	Path      string     `json:"path" db:"path"`
	Value     string     `json:"-" db:"-"`
	Secure    bool       `json:"secure" db:"secure"`
	HTTPOnly  bool       `json:"http_only" db:"http_only"`
	SameSite  string     `json:"same_site,omitempty" db:"same_site"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Source    string     `json:"source,omitempty" db:"source"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

type BrowserCookieFilter struct {
	Domain string
	Name   string
	Path   string
}

type BrowserCookieStore interface {
	Upsert(ctx context.Context, scope BrowserCookieScope, cookies []BrowserCookie) (int, error)
	List(ctx context.Context, scope BrowserCookieScope, filter BrowserCookieFilter) ([]BrowserCookie, error)
	Delete(ctx context.Context, scope BrowserCookieScope, filter BrowserCookieFilter) (int, error)
}

func BrowserCookieScopeFromContext(ctx context.Context, agentID string) BrowserCookieScope {
	tid := TenantIDFromContext(ctx)
	if tid == uuid.Nil {
		tid = MasterTenantID
	}
	return BrowserCookieScope{
		TenantID: tid,
		UserID:   CredentialUserIDFromContext(ctx),
		AgentID:  strings.TrimSpace(agentID),
	}
}

func (s BrowserCookieScope) Validate() error {
	switch {
	case s.TenantID == uuid.Nil:
		return errors.New("tenant_id required")
	case strings.TrimSpace(s.UserID) == "":
		return errors.New("user_id required")
	case strings.TrimSpace(s.AgentID) == "":
		return errors.New("agent_id required")
	default:
		return nil
	}
}

func NormalizeBrowserCookie(c BrowserCookie) BrowserCookie {
	c.Domain = strings.ToLower(strings.TrimSpace(c.Domain))
	c.Name = strings.TrimSpace(c.Name)
	c.Path = strings.TrimSpace(c.Path)
	c.SameSite = strings.TrimSpace(c.SameSite)
	c.Source = strings.TrimSpace(c.Source)
	if c.Path == "" {
		c.Path = "/"
	}
	return c
}

func ValidateBrowserCookie(c BrowserCookie) error {
	c = NormalizeBrowserCookie(c)
	switch {
	case c.Domain == "":
		return errors.New("domain required")
	case c.Name == "":
		return errors.New("name required")
	case c.Path == "":
		return errors.New("path required")
	default:
		return nil
	}
}

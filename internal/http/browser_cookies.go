package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nextlevelbuilder/goclaw/internal/i18n"
	"github.com/nextlevelbuilder/goclaw/internal/permissions"
	"github.com/nextlevelbuilder/goclaw/internal/store"
)

const (
	maxBrowserCookieSyncBodyBytes = 1 << 20
	maxBrowserCookieSyncItems     = 200
	maxBrowserCookieValueBytes    = 16 << 10
)

var (
	errBrowserCookieInvalidURL     = errors.New("invalid cookie url")
	errBrowserCookieValueTooLarge  = errors.New("cookie value too large")
	errBrowserCookieTooManyCookies = errors.New("too many cookies")
)

// BrowserCookiesHandler stores selected browser cookies for server-side browser sessions.
type BrowserCookiesHandler struct {
	cookies store.BrowserCookieStore
}

func NewBrowserCookiesHandler(cookies store.BrowserCookieStore) *BrowserCookiesHandler {
	return &BrowserCookiesHandler{cookies: cookies}
}

func (h *BrowserCookiesHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/browser/cookies/sync", requireAuth(permissions.RoleOperator, h.handleSync))
	mux.HandleFunc("GET /v1/browser/cookies", requireAuth(permissions.RoleOperator, h.handleList))
	mux.HandleFunc("DELETE /v1/browser/cookies", requireAuth(permissions.RoleOperator, h.handleDelete))
}

type browserCookieSyncRequest struct {
	AgentID string                     `json:"agent_id"`
	Agent   string                     `json:"agent,omitempty"`
	Source  string                     `json:"source,omitempty"`
	Cookies []browserCookieSyncPayload `json:"cookies"`
}

type browserCookieSyncPayload struct {
	Domain              string     `json:"domain"`
	Name                string     `json:"name"`
	Path                string     `json:"path"`
	Value               string     `json:"value"`
	URL                 string     `json:"url,omitempty"`
	Secure              bool       `json:"secure"`
	HTTPOnly            bool       `json:"httpOnly"`
	HTTPOnlySnake       bool       `json:"http_only"`
	SameSite            string     `json:"sameSite"`
	SameSiteSnake       string     `json:"same_site"`
	ExpiresAt           *time.Time `json:"expiresAt"`
	ExpiresAtSnake      *time.Time `json:"expires_at"`
	ExpirationDate      *float64   `json:"expirationDate"`
	ExpirationDateSnake *float64   `json:"expiration_date"`
}

type browserCookieMetadata struct {
	Domain    string     `json:"domain"`
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	Secure    bool       `json:"secure"`
	HTTPOnly  bool       `json:"httpOnly"`
	SameSite  string     `json:"sameSite,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	Source    string     `json:"source,omitempty"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (h *BrowserCookiesHandler) handleSync(w http.ResponseWriter, r *http.Request) {
	locale := store.LocaleFromContext(r.Context())
	var req browserCookieSyncRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBrowserCookieSyncBodyBytes)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": i18n.T(locale, i18n.MsgInvalidJSON)})
		return
	}

	agentID := firstBrowserCookieNonEmpty(req.AgentID, req.Agent)
	scope, ok := h.scopeFromRequest(w, r, agentID)
	if !ok {
		return
	}
	if len(req.Cookies) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": i18n.T(locale, i18n.MsgRequired, "cookies")})
		return
	}
	if len(req.Cookies) > maxBrowserCookieSyncItems {
		h.writeValidationError(w, locale, errBrowserCookieTooManyCookies)
		return
	}

	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = "chrome-extension"
	}
	cookies := make([]store.BrowserCookie, 0, len(req.Cookies))
	for _, item := range req.Cookies {
		c, err := item.toStoreCookie(source)
		if err != nil {
			h.writeValidationError(w, locale, err)
			return
		}
		cookies = append(cookies, c)
	}

	count, err := h.cookies.Upsert(r.Context(), scope, cookies)
	if err != nil {
		h.writeStoreError(w, locale, "browser_cookie_sync.upsert", err)
		return
	}
	slog.Info("browser_cookie_sync.synced",
		"tenant_id", scope.TenantID,
		"user_id", scope.UserID,
		"agent_id", scope.AgentID,
		"count", count,
		"source", source,
	)
	writeJSON(w, http.StatusOK, map[string]any{"synced": count})
}

func (h *BrowserCookiesHandler) handleList(w http.ResponseWriter, r *http.Request) {
	locale := store.LocaleFromContext(r.Context())
	scope, ok := h.scopeFromRequest(w, r, r.URL.Query().Get("agent_id"))
	if !ok {
		return
	}
	cookies, err := h.cookies.List(r.Context(), scope, browserCookieFilterFromQuery(r))
	if err != nil {
		h.writeStoreError(w, locale, "browser_cookie_sync.list", err)
		return
	}
	items := make([]browserCookieMetadata, 0, len(cookies))
	for _, c := range cookies {
		items = append(items, browserCookieMetadata{
			Domain:    c.Domain,
			Name:      c.Name,
			Path:      c.Path,
			Secure:    c.Secure,
			HTTPOnly:  c.HTTPOnly,
			SameSite:  c.SameSite,
			ExpiresAt: c.ExpiresAt,
			Source:    c.Source,
			UpdatedAt: c.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *BrowserCookiesHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	locale := store.LocaleFromContext(r.Context())
	scope, ok := h.scopeFromRequest(w, r, r.URL.Query().Get("agent_id"))
	if !ok {
		return
	}
	deleted, err := h.cookies.Delete(r.Context(), scope, browserCookieFilterFromQuery(r))
	if err != nil {
		h.writeStoreError(w, locale, "browser_cookie_sync.delete", err)
		return
	}
	slog.Info("browser_cookie_sync.deleted",
		"tenant_id", scope.TenantID,
		"user_id", scope.UserID,
		"agent_id", scope.AgentID,
		"count", deleted,
	)
	writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
}

func (h *BrowserCookiesHandler) scopeFromRequest(w http.ResponseWriter, r *http.Request, agentID string) (store.BrowserCookieScope, bool) {
	locale := store.LocaleFromContext(r.Context())
	scope := store.BrowserCookieScopeFromContext(r.Context(), agentID)
	if err := scope.Validate(); err != nil {
		slog.Warn("security.browser_cookie_sync.scope_denied", "error", err, "path", r.URL.Path)
		h.writeValidationError(w, locale, err)
		return store.BrowserCookieScope{}, false
	}
	return scope, true
}

func (h *BrowserCookiesHandler) writeValidationError(w http.ResponseWriter, locale string, err error) {
	msg := i18n.T(locale, i18n.MsgInvalidRequest, "browser cookies")
	switch {
	case errors.Is(err, store.ErrBrowserCookieTenantRequired):
		msg = i18n.T(locale, i18n.MsgRequired, "tenant_id")
	case errors.Is(err, store.ErrBrowserCookieUserRequired):
		msg = i18n.T(locale, i18n.MsgRequired, "user_id")
	case errors.Is(err, store.ErrBrowserCookieAgentRequired):
		msg = i18n.T(locale, i18n.MsgRequired, "agent_id")
	case errors.Is(err, store.ErrBrowserCookieDomainRequired):
		msg = i18n.T(locale, i18n.MsgRequired, "domain")
	case errors.Is(err, store.ErrBrowserCookieNameRequired):
		msg = i18n.T(locale, i18n.MsgRequired, "name")
	case errors.Is(err, store.ErrBrowserCookiePathRequired):
		msg = i18n.T(locale, i18n.MsgRequired, "path")
	case errors.Is(err, errBrowserCookieInvalidURL):
		msg = i18n.T(locale, i18n.MsgInvalidCookieURL)
	case errors.Is(err, errBrowserCookieValueTooLarge):
		msg = i18n.T(locale, i18n.MsgBrowserCookieValueTooLarge)
	case errors.Is(err, errBrowserCookieTooManyCookies):
		msg = i18n.T(locale, i18n.MsgBrowserCookieTooMany)
	}
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func (h *BrowserCookiesHandler) writeStoreError(w http.ResponseWriter, locale, op string, err error) {
	if errors.Is(err, store.ErrBrowserCookieEncryptionRequired) {
		slog.Warn("security.browser_cookie_sync.encryption_required", "op", op)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": i18n.T(locale, i18n.MsgBrowserCookieEncryptionRequired)})
		return
	}
	slog.Error(op, "error", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": i18n.T(locale, i18n.MsgInternalError, "browser cookies")})
}

func (p browserCookieSyncPayload) toStoreCookie(source string) (store.BrowserCookie, error) {
	domain := strings.TrimSpace(p.Domain)
	if domain == "" && p.URL != "" {
		u, err := url.Parse(p.URL)
		if err != nil {
			return store.BrowserCookie{}, errBrowserCookieInvalidURL
		}
		domain = u.Hostname()
	}
	if len(p.Value) > maxBrowserCookieValueBytes {
		return store.BrowserCookie{}, errBrowserCookieValueTooLarge
	}
	c := store.NormalizeBrowserCookie(store.BrowserCookie{
		Domain:    domain,
		Name:      p.Name,
		Path:      p.Path,
		Value:     p.Value,
		Secure:    p.Secure,
		HTTPOnly:  p.HTTPOnly || p.HTTPOnlySnake,
		SameSite:  firstBrowserCookieNonEmpty(p.SameSite, p.SameSiteSnake),
		ExpiresAt: firstTime(p.ExpiresAt, p.ExpiresAtSnake, timeFromUnixSeconds(p.ExpirationDate), timeFromUnixSeconds(p.ExpirationDateSnake)),
		Source:    source,
	})
	if err := store.ValidateBrowserCookie(c); err != nil {
		return store.BrowserCookie{}, err
	}
	return c, nil
}

func browserCookieFilterFromQuery(r *http.Request) store.BrowserCookieFilter {
	q := r.URL.Query()
	return store.BrowserCookieFilter{
		Domain: q.Get("domain"),
		Name:   q.Get("name"),
		Path:   q.Get("path"),
	}
}

func firstBrowserCookieNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func firstTime(values ...*time.Time) *time.Time {
	for _, v := range values {
		if v != nil {
			t := v.UTC()
			return &t
		}
	}
	return nil
}

func timeFromUnixSeconds(v *float64) *time.Time {
	if v == nil {
		return nil
	}
	sec := int64(*v)
	nsec := int64((*v - float64(sec)) * 1e9)
	t := time.Unix(sec, nsec).UTC()
	return &t
}

package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type contextKey string

const (
	authEnabledContextKey contextKey = "auth_enabled"
	actorUserIDContextKey contextKey = "actor_user_id"
	actorRolesContextKey  contextKey = "actor_roles"
	requestIPContextKey   contextKey = "request_ip"
	requestUAContextKey   contextKey = "request_user_agent"
	tenantIDContextKey    contextKey = "tenant_id"
)

func AuthEnabled(ctx context.Context) bool {
	value := ctx.Value(authEnabledContextKey)
	enabled, ok := value.(bool)
	return ok && enabled
}

func SetAuthContext(enabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), authEnabledContextKey, enabled)
			ctx = withActorFromHeaders(ctx, r)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ActorUserIDFromContext(ctx context.Context) *string {
	value := ctx.Value(actorUserIDContextKey)
	userID, ok := value.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return nil
	}
	userID = strings.TrimSpace(userID)
	return &userID
}

func ActorRolesFromContext(ctx context.Context) []string {
	value := ctx.Value(actorRolesContextKey)
	roles, ok := value.([]string)
	if !ok {
		return nil
	}

	filtered := make([]string, 0, len(roles))
	for _, role := range roles {
		value := strings.TrimSpace(role)
		if value != "" {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func TenantIDFromContext(ctx context.Context) *string {
	value := ctx.Value(tenantIDContextKey)
	tenantID, ok := value.(string)
	if !ok || strings.TrimSpace(tenantID) == "" {
		return nil
	}
	tenantID = strings.TrimSpace(tenantID)
	return &tenantID
}

func HasRole(ctx context.Context, expectedRoles ...string) bool {
	if !AuthEnabled(ctx) {
		return true
	}

	roles := ActorRolesFromContext(ctx)
	if len(roles) == 0 {
		return len(expectedRoles) == 0
	}

	needed := map[string]struct{}{}
	for _, role := range expectedRoles {
		role = strings.ToLower(strings.TrimSpace(role))
		if role != "" {
			needed[role] = struct{}{}
		}
	}

	if len(needed) == 0 {
		return true
	}

	for _, role := range roles {
		if _, ok := needed[strings.ToLower(role)]; ok {
			return true
		}
	}

	return false
}

func RequestIPAddressFromContext(ctx context.Context) *string {
	value := ctx.Value(requestIPContextKey)
	ip, ok := value.(string)
	if !ok {
		return nil
	}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return nil
	}
	return &ip
}

func RequestUserAgentFromContext(ctx context.Context) *string {
	value := ctx.Value(requestUAContextKey)
	ua, ok := value.(string)
	if !ok {
		return nil
	}
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return nil
	}
	return &ua
}

func withActorFromHeaders(ctx context.Context, r *http.Request) context.Context {
	actorID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if actorID == "" {
		actorID = strings.TrimSpace(r.Header.Get("X-User-Id"))
	}
	if actorID != "" {
		ctx = context.WithValue(ctx, actorUserIDContextKey, actorID)
	}

	tenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	if tenantID != "" {
		ctx = context.WithValue(ctx, tenantIDContextKey, tenantID)
	}

	if rolesHeader := strings.TrimSpace(r.Header.Get("X-User-Roles")); rolesHeader != "" {
		roles := []string{}
		for _, role := range strings.FieldsFunc(rolesHeader, func(r rune) bool {
			return r == ',' || r == ';' || r == ' '
		}) {
			value := strings.ToLower(strings.TrimSpace(role))
			if value != "" {
				roles = append(roles, value)
			}
		}
		if len(roles) > 0 {
			ctx = context.WithValue(ctx, actorRolesContextKey, roles)
		}
	}

	if requestIP := extractRequestIPAddress(r); requestIP != "" {
		ctx = context.WithValue(ctx, requestIPContextKey, requestIP)
	}
	if userAgent := strings.TrimSpace(r.Header.Get("User-Agent")); userAgent != "" {
		ctx = context.WithValue(ctx, requestUAContextKey, userAgent)
	}

	return ctx
}

func extractRequestIPAddress(r *http.Request) string {
	rawIP := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if rawIP == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			rawIP = r.RemoteAddr
		} else {
			rawIP = host
		}
	}

	if rawIP == "" {
		return ""
	}
	if strings.Contains(rawIP, ",") {
		return strings.TrimSpace(strings.Split(rawIP, ",")[0])
	}
	return rawIP
}

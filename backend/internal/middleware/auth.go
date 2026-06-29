package middleware

import (
\t"net/http"

\t"github.com/example/file-platform/backend/internal/config"
\t"github.com/gin-gonic/gin"
)

const userContextKey = "userID"

func Auth(cfg config.Config) gin.HandlerFunc {
\treturn func(c *gin.Context) {
\t\tif c.Request.Method == http.MethodOptions {
\t\t\tc.Next()
\t\t\treturn
\t\t}
\t\tif c.Request.URL.Path == "/health" || c.Request.URL.Path == "/ready" {
\t\t\tc.Next()
\t\t\treturn
\t\t}
\t\tif !cfg.AuthEnabled {
\t\t\tc.Set(userContextKey, "anonymous")
\t\t\tc.Next()
\t\t\treturn
\t\t}
\t\th := c.GetHeader("X-API-Key")
\t\tif h == "" || h != cfg.APIKey {
\t\t\tc.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
\t\t\treturn
\t\t}
\t\tuser := c.GetHeader("X-User-Id")
\t\tif user == "" {
\t\t\tuser = "unknown"
\t\t}
\t\tc.Set(userContextKey, user)
\t\tc.Next()
\t}
}

func UserID(c *gin.Context) string {
\tv, ok := c.Get(userContextKey)
\tif !ok {
\t\treturn "anonymous"
\t}
\tif user, ok := v.(string); ok {
\t\treturn user
\t}
\treturn "anonymous"
}


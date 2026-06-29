package middleware

import (
\t"net/http"

\t"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
\treturn func(c *gin.Context) {
\t\tc.Writer.Header().Set("Access-Control-Allow-Origin", "*")
\t\tc.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key, X-User-Id")
\t\tc.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
\t\tif c.Request.Method == http.MethodOptions {
\t\t\tc.Status(http.StatusNoContent)
\t\t\treturn
\t\t}
\t\tc.Next()
\t}
}


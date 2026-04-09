package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"bet/backend/internal/auth"
)

const claimsContextKey = "auth_claims"

func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := parseClaimsFromAuthorizationHeader(secret, c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

func RequireRoles(secret string, roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		claims, ok := parseClaimsFromAuthorizationHeader(secret, c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if _, exists := allowed[claims.Role]; !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

func ClaimsFromContext(c *gin.Context) (auth.Claims, bool) {
	v, ok := c.Get(claimsContextKey)
	if !ok {
		return auth.Claims{}, false
	}

	claims, ok := v.(auth.Claims)
	if !ok {
		return auth.Claims{}, false
	}

	return claims, true
}

func parseClaimsFromAuthorizationHeader(secret, header string) (auth.Claims, bool) {
	if !strings.HasPrefix(header, "Bearer ") {
		return auth.Claims{}, false
	}

	token := strings.TrimPrefix(header, "Bearer ")
	claims, err := auth.ParseToken(secret, token)
	if err != nil {
		return auth.Claims{}, false
	}

	return claims, true
}

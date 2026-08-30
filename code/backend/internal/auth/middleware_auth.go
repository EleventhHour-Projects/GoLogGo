
package auth

import (
	"strings"

	"github.com/nottechdm/notnet/pkg/notnet"
)

type authedHandler func(*notnet.Request, *notnet.Response, *CustomClaims) error

func MiddlewareAuth(handler authedHandler) notnet.HandlerFunc {
	return func(req *notnet.Request, res *notnet.Response) error {

		authHeader := req.HTTPRequest.Header.Get("Authorization")

		if authHeader == "" {
			return res.JSON(401, map[string]string{
				"error": "Missing Authorization header",
			})
		}

		const bearerPrefix = "Bearer "

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			return res.JSON(401, map[string]string{
				"error": "Invalid Authorization header",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		if tokenString == "" {
			return res.JSON(401, map[string]string{
				"error": "Missing Authorization header",
			})
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			return res.JSON(401, map[string]string{
				"error": "Unauthorized",
			})
		}

		return handler(req, res, claims)
	}
}


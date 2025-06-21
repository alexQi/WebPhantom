package middlewares

import (
	corsMiddleware "github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
)

func CorsNew() iris.Handler {
	corsHandler := corsMiddleware.New(corsMiddleware.Options{
		AllowedOrigins:   []string{"*"}, //允许通过的主机名称
		AllowCredentials: true,
		AllowedHeaders:   []string{"Origin", "Content-Type", "Cookie", "X-CSRF-TOKEN", "Accept", "Authorization", "X-XSRF-TOKEN"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		Debug:            false,
	})
	return corsHandler
}

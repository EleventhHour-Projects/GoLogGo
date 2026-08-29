package api

import (
	"github.com/nottechdm/notnet/pkg/notnet"
)

func RegisterRoutes(app *notnet.Engine) {
	app.Use(notnet.Logger(), notnet.Recovery())
	app.POST("/log", LogHandler)
}

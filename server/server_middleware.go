package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// ApplyCorsRestrictions godoc
func (srv *server) ApplyCorsRestrictions(r *gin.Engine) {
	if !srv.Config.Server.API.Cors {
		return
	}
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "X-Requested-With", "Content-Length", "Content-Type", "Accept", "X-Api-Key", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"}
	r.Use(cors.New(corsConfig)) // Allow requests from anywhere
}

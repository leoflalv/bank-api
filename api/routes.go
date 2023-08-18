package api

import "github.com/gin-gonic/gin"

func SetupRoutes(server *Server) {
	router := gin.Default()

	// Accounts
	router.GET("/account/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)
	router.POST("/account", server.createAccount)
	router.PUT("/account", server.updateAccount)
	router.DELETE("/account/:id", server.deleteAccount)

	server.router = router
}

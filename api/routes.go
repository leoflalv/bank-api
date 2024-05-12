package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func SetupRoutes(server *Server) {
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	// Users
	router.GET("/user/:username", server.getUser)
	router.POST("/user", server.createUser)
	router.POST("/user/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenManager))

	// Accounts
	authRoutes.GET("/account/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccounts)
	authRoutes.POST("/account", server.createAccount)
	authRoutes.PUT("/account", server.updateAccount)
	authRoutes.DELETE("/account/:id", server.deleteAccount)

	// Transfers
	authRoutes.POST("/transaction", server.createTransaction)

	server.router = router
}

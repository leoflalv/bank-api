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

	// Accounts
	router.GET("/account/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)
	router.POST("/account", server.createAccount)
	router.PUT("/account", server.updateAccount)
	router.DELETE("/account/:id", server.deleteAccount)

	// Transfers
	router.POST("/transfer", server.createTransfer)

	server.router = router
}

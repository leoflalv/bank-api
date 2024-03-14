package api

import (
	"fmt"
	db "github/leoflalv/bank-api/db/sqlc"
	"github/leoflalv/bank-api/token"
	"github/leoflalv/bank-api/util"

	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests.
type Server struct {
	config       util.Config
	store        db.Store
	tokenManager token.Manager
	router       *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenManager, err := token.NewPasetoManager(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token manager: %w", err)
	}

	server := &Server{store: store, tokenManager: tokenManager, config: config}

	return server, nil
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/thirochas/simplebank-golang-api/internal/repository"
	"github.com/thirochas/simplebank-golang-api/internal/token"
	"github.com/thirochas/simplebank-golang-api/internal/util"
	"net/http"
)

type Server struct {
	store      repository.Store
	tokenMaker token.Maker
	config     util.Config
	router     *gin.Engine
}

func NewServer(store repository.Store, config util.Config) *Server {
	maker, err := getMaker(config)
	if err != nil {
		panic(fmt.Sprintf("error trying to load token maker: %v", err))
	}

	server := &Server{
		store:      store,
		tokenMaker: maker,
		config:     config,
	}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	router.GET("/ping", server.ping)

	router.POST("/users/login", server.createToken)

	router.POST("/users", server.createUser)
	router.GET("/users/:username", server.getUserByUsername)

	authRoutes := router.Group("/").Use(authenticationMiddleware(server.tokenMaker))
	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.PUT("/accounts", server.updateAccountById)
	authRoutes.GET("/accounts/:id", server.getAccountById)
	authRoutes.GET("/accounts", server.getAccounts)
	authRoutes.DELETE("/accounts/:id", server.deleteAccountById)

	authRoutes.POST("/transfers", server.createTransfer)

	server.router = router

	return server
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func (s *Server) ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "pong")
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func getMaker(config util.Config) (token.Maker, error) {
	if config.TokenType == "paseto" {
		return token.NewPasetoMaker(config.SecretKey)
	}

	return token.NewJWTMaker(config.SecretKey)
}

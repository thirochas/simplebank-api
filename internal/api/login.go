package api

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/thirochas/simplebank-golang-api/internal/util"
	"net/http"
	"time"
)

const (
	TokenDuration             = time.Minute * 15
	InvalidUsernameOrPassword = "invalid username or password"
)

type createTokenRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type createTokenResponse struct {
	AccessToken string    `json:"access_token"`
	ExpireAt    time.Time `json:"expired_at"`
}

func (s *Server) createToken(ctx *gin.Context) {
	var req createTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	invalidUsernameOrPasswordError := errors.New(InvalidUsernameOrPassword)

	user, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse(invalidUsernameOrPasswordError))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(invalidUsernameOrPasswordError))
		return
	}

	token, payload, err := s.tokenMaker.CreateToken(user.Username, TokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := createTokenResponse{
		AccessToken: token,
		ExpireAt:    payload.ExpiredAt,
	}

	ctx.JSON(http.StatusOK, response)
}

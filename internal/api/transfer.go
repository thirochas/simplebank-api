package api

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/thirochas/simplebank-golang-api/internal/repository"
	"github.com/thirochas/simplebank-golang-api/internal/token"
	"net/http"
)

const (
	AccountFrom = iota
	AccountTo
)

type createTransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required"`
	ToAccountID   int64  `json:"to_account_id" binding:"required"`
	Currency      string `json:"currency" binding:"required,currency"`
	Amount        int64  `json:"amount" binding:"required,min=1"`
}

func (s *Server) createTransfer(ctx *gin.Context) {
	var req createTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !s.isValid(ctx, req.FromAccountID, req.Currency, AccountFrom) || !s.isValid(ctx, req.ToAccountID, req.Currency, AccountTo) {
		return
	}

	arg := repository.CreateTransferParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	transfer, err := s.store.CreateTransfer(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, transfer)
}

func (s *Server) isValid(ctx *gin.Context, accountID int64, currency string, accountType int) bool {
	account, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if accountType == AccountFrom {
		payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
		if payload.Username != account.Owner {
			err := errors.New("authenticated account different from account founded")
			ctx.JSON(http.StatusUnauthorized, errorResponse(err))
			return false
		}
	} else {
		payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
		if payload.Username == account.Owner {
			err := errors.New("authenticated account cannot transfer to itself")
			ctx.JSON(http.StatusUnauthorized, errorResponse(err))
			return false
		}
	}

	if account.Currency != currency {
		currencyError := errors.New(
			fmt.Sprintf("currency %s is not valid with account currency %s of account-id %d",
				currency, account.Currency, account.ID))
		ctx.JSON(http.StatusBadRequest, errorResponse(currencyError))
		return false
	}

	return true
}

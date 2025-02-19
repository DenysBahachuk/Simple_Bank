package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/DenysBahachuk/Simple_Bank/db/sqlc"
	"github.com/DenysBahachuk/Simple_Bank/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
}

// CreateAccount godoc
//
//	@Summary	Creates an account
//	@Schemes
//	@Description	Creates an account
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		createAccountRequest	true	"Account payload"
//	@Success		200		{object}	db.Account
//	@Failure		400		{string}	error	"Bad request"
//	@Failure		403		{string}	error	"Forbidden"
//	@Failure		500		{string}	error	"Internal server error"
//	@Security		ApiKeyAuth
//	@Router			/accounts [post]
func (s *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	userPayload := ctx.MustGet(authPayloadKey).(*token.Payload)

	args := db.CreateAccountParams{
		Owner:    userPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := s.store.CreateAccount(ctx, args)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// GetAccount godoc
//
//	@Summary	Fetches an account
//	@Schemes
//	@Description	Fetches an account by ID
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Account ID"
//	@Success		200	{object}	db.Account
//	@Failure		400	{string}	error	"Bad request"
//	@Failure		404	{string}	error	"Account not found"
//	@Failure		500	{string}	error	"Internal server error"
//	@Security		ApiKeyAuth
//	@Router			/accounts/{id} [get]
func (s *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest

	err := ctx.ShouldBindUri(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := s.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	userPayload := ctx.MustGet(authPayloadKey).(*token.Payload)
	if account.Owner != userPayload.Username {
		err := fmt.Errorf("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type listAccountsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// ListAccounts godoc
//
//	@Summary		Fetches all accounts
//	@Description	Fetches all accounts
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			page_id		query		int	false	"Page ID"
//	@Param			page_size	query		int	false	"Page Size"
//	@Success		200			{object}	[]db.Account
//	@Failure		400			{string}	error	"Bad request"
//	@Failure		500			{string}	error	"Internal server error"
//	@Security		ApiKeyAuth
//	@Router			/accounts [get]
func (s *Server) listAccounts(ctx *gin.Context) {
	var req listAccountsRequest

	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	userPayload := ctx.MustGet(authPayloadKey).(*token.Payload)

	args := db.ListAccountsParams{
		Owner:  userPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := s.store.ListAccounts(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}

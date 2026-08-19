package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/interfaces/http/dto"
	"github.com/apolzek/pastebomb/internal/interfaces/http/httperr"
)

// AuthHandler exposes the sign-in endpoint.
type AuthHandler struct{ auth *application.AuthService }

func NewAuthHandler(auth *application.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Login handles POST /login.
func (h *AuthHandler) Login(ctx *gin.Context) {
	var request dto.LoginRequest
	if err := ctx.ShouldBind(&request); err != nil {
		httperr.Invalid(ctx, err)
		return
	}

	token, err := h.auth.Login(ctx.Request.Context(), application.LoginCommand{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		httperr.Write(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Envelope{
		Message: "login success",
		Data: dto.TokenView{
			Token:     token.Value,
			ExpiresAt: token.ExpiresAt,
			TokenType: "Bearer",
		},
	})
}

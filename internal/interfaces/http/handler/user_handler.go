package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/domain/user"
	"github.com/apolzek/pastebomb/internal/interfaces/http/dto"
	"github.com/apolzek/pastebomb/internal/interfaces/http/httperr"
	"github.com/apolzek/pastebomb/internal/interfaces/http/reqctx"
)

// UserHandler exposes the account endpoints.
type UserHandler struct{ users *application.UserService }

func NewUserHandler(users *application.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// Register handles POST /u/.
func (h *UserHandler) Register(ctx *gin.Context) {
	var request dto.RegisterUserRequest
	if err := ctx.ShouldBind(&request); err != nil {
		httperr.Invalid(ctx, err)
		return
	}

	account, err := h.users.Register(ctx.Request.Context(), application.RegisterCommand(request))
	if err != nil {
		httperr.Write(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.Envelope{
		Message: "user created successfully",
		Data:    dto.NewUserView(account),
	})
}

// Me handles GET /u/me.
func (h *UserHandler) Me(ctx *gin.Context) {
	identity := reqctx.Identity(ctx)

	account, err := h.users.Profile(ctx.Request.Context(), identity.UserID)
	if err != nil {
		httperr.Write(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Envelope{
		Message: "data transmitted",
		Data:    dto.NewUserView(account),
	})
}

// UpdateMe handles PATCH /u/me.
func (h *UserHandler) UpdateMe(ctx *gin.Context) {
	identity := reqctx.Identity(ctx)

	var request dto.UpdateProfileRequest
	if err := ctx.ShouldBind(&request); err != nil {
		httperr.Invalid(ctx, err)
		return
	}

	account, err := h.users.UpdateProfile(ctx.Request.Context(), identity.UserID, application.UpdateProfileCommand(request))
	if err != nil {
		httperr.Write(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Envelope{
		Message: "user updated successfully",
		Data:    dto.NewUserView(account),
	})
}

// DeactivateMe handles DELETE /u/me. It has its own handler because the
// path-parameter version could never parse the literal "me" as an id.
func (h *UserHandler) DeactivateMe(ctx *gin.Context) {
	identity := reqctx.Identity(ctx)
	h.deactivate(ctx, identity, identity.UserID)
}

// DeactivateByID handles DELETE /u/:id.
func (h *UserHandler) DeactivateByID(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("handle"), 10, 64)
	if err != nil || id <= 0 {
		httperr.Write(ctx, shared.Invalid("user.id_invalid", "user id must be a positive integer"))
		return
	}
	h.deactivate(ctx, reqctx.Identity(ctx), user.ID(id))
}

func (h *UserHandler) deactivate(ctx *gin.Context, identity application.Identity, target user.ID) {
	if err := h.users.Deactivate(ctx.Request.Context(), identity, target); err != nil {
		httperr.Write(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.Envelope{Message: "user deleted successfully"})
}

// PublicProfile handles GET /u/:handle. It is unauthenticated, so it renders
// the reduced view that omits the email address and date of birth.
func (h *UserHandler) PublicProfile(ctx *gin.Context) {
	account, err := h.users.PublicProfile(ctx.Request.Context(), ctx.Param("handle"))
	if err != nil {
		httperr.Write(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Envelope{
		Message: "data transmitted",
		Data:    dto.NewPublicUserView(account),
	})
}

// List handles GET /u/all, restricted to administrators.
func (h *UserHandler) List(ctx *gin.Context) {
	page := pageFrom(ctx)

	accounts, err := h.users.List(ctx.Request.Context(), page)
	if err != nil {
		httperr.Write(ctx, err)
		return
	}
	writeUserPage(ctx, accounts, page)
}

// ListActive handles GET /u/ and GET /u/paginate, restricted to administrators.
func (h *UserHandler) ListActive(ctx *gin.Context) {
	page := pageFrom(ctx)

	accounts, err := h.users.ListActive(ctx.Request.Context(), page)
	if err != nil {
		httperr.Write(ctx, err)
		return
	}
	writeUserPage(ctx, accounts, page)
}

func writeUserPage(ctx *gin.Context, accounts []*user.User, page shared.Page) {
	views := dto.NewUserViews(accounts)
	ctx.JSON(http.StatusOK, dto.Envelope{
		Message: "data transmitted",
		Data:    views,
		Meta:    &dto.Meta{Page: page.Number(), PerPage: page.Size(), Count: len(views)},
	})
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/domain/paste"
	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/interfaces/http/dto"
	"github.com/apolzek/pastebomb/internal/interfaces/http/httperr"
	"github.com/apolzek/pastebomb/internal/interfaces/http/reqctx"
)

// secretHeader carries the access secret of a protected paste. A header keeps
// it out of URLs, and therefore out of access logs and Referer headers.
const secretHeader = "X-Post-Secret"

// PasteHandler exposes the paste endpoints.
type PasteHandler struct{ pastes *application.PasteService }

func NewPasteHandler(pastes *application.PasteService) *PasteHandler {
	return &PasteHandler{pastes: pastes}
}

// CreateAnonymous handles POST /u/anonymous/post. Anonymous pastes are always
// public: an unlisted paste with no owner could never be listed again.
func (h *PasteHandler) CreateAnonymous(ctx *gin.Context) {
	h.create(ctx, 0, true)
}

// CreateOwned handles POST /u/me/post.
func (h *PasteHandler) CreateOwned(ctx *gin.Context) {
	identity := reqctx.Identity(ctx)
	h.create(ctx, paste.OwnerID(identity.UserID), false)
}

func (h *PasteHandler) create(ctx *gin.Context, owner paste.OwnerID, defaultPublic bool) {
	var request dto.CreatePasteRequest
	if err := ctx.ShouldBind(&request); err != nil {
		httperr.Invalid(ctx, err)
		return
	}

	isPublic := request.PublicOr(defaultPublic)
	if owner.IsAnonymous() {
		isPublic = true
	}

	document, err := h.pastes.Create(ctx.Request.Context(), application.CreatePasteCommand{
		Owner:     owner,
		Title:     request.Title,
		Content:   request.Content,
		Category:  request.Category,
		IsPublic:  isPublic,
		ExpiresAt: request.ExpirationDate,
		Secret:    request.SecretAccess,
	})
	if err != nil {
		httperr.Write(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, dto.Envelope{
		Message: "data created successfully",
		Data:    dto.NewPasteView(document),
	})
}

// Read handles GET /u/anonymous/:slug.
func (h *PasteHandler) Read(ctx *gin.Context) {
	document, err := h.pastes.Read(ctx.Request.Context(), ctx.Param("slug"), ctx.GetHeader(secretHeader))
	if err != nil {
		httperr.Write(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Envelope{
		Message: "data transmitted",
		Data:    dto.NewPasteView(document),
	})
}

// ReadRaw handles GET /u/anonymous/:slug/raw, serving the body as plain text
// so a paste can be fetched with curl.
//
// The response is forced to text/plain with nosniff and an attachment
// disposition, so a paste containing HTML cannot execute in a browser.
func (h *PasteHandler) ReadRaw(ctx *gin.Context) {
	document, err := h.pastes.Read(ctx.Request.Context(), ctx.Param("slug"), ctx.GetHeader(secretHeader))
	if err != nil {
		httperr.WriteText(ctx, err)
		return
	}

	ctx.Header("Content-Disposition", "attachment; filename=\""+document.Slug().String()+".txt\"")
	ctx.Data(http.StatusOK, "text/plain; charset=utf-8", document.Content().Bytes())
}

// ListPublic handles GET /u/anonymous/posts. It returns metadata only: the
// index must not hand out the body of a secret-protected paste.
func (h *PasteHandler) ListPublic(ctx *gin.Context) {
	page := pageFrom(ctx)

	documents, err := h.pastes.ListPublic(ctx.Request.Context(), page)
	if err != nil {
		httperr.Write(ctx, err)
		return
	}

	summaries := dto.NewPasteSummaries(documents)
	ctx.JSON(http.StatusOK, dto.Envelope{
		Message: "data transmitted",
		Data:    summaries,
		Meta:    &dto.Meta{Page: page.Number(), PerPage: page.Size(), Count: len(summaries)},
	})
}

// ListMine handles GET /u/me/posts, returning the caller's own pastes without
// their bodies.
func (h *PasteHandler) ListMine(ctx *gin.Context) {
	documents, page, err := h.ownedPastes(ctx)
	if err != nil {
		httperr.Write(ctx, err)
		return
	}

	summaries := dto.NewPasteSummaries(documents)
	ctx.JSON(http.StatusOK, dto.Envelope{
		Message: "data transmitted",
		Data:    summaries,
		Meta:    &dto.Meta{Page: page.Number(), PerPage: page.Size(), Count: len(summaries)},
	})
}

// ListMineWithContent handles GET /u/me/posts/content. Bodies are included
// because the caller already owns every paste in the result.
func (h *PasteHandler) ListMineWithContent(ctx *gin.Context) {
	documents, page, err := h.ownedPastes(ctx)
	if err != nil {
		httperr.Write(ctx, err)
		return
	}

	views := dto.NewPasteViews(documents)
	ctx.JSON(http.StatusOK, dto.Envelope{
		Message: "data transmitted",
		Data:    views,
		Meta:    &dto.Meta{Page: page.Number(), PerPage: page.Size(), Count: len(views)},
	})
}

func (h *PasteHandler) ownedPastes(ctx *gin.Context) ([]*paste.Paste, shared.Page, error) {
	identity := reqctx.Identity(ctx)
	page := pageFrom(ctx)

	documents, err := h.pastes.ListOwned(ctx.Request.Context(), paste.OwnerID(identity.UserID), page)
	return documents, page, err
}

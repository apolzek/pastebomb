// Package handler adapts HTTP requests to application use cases. Handlers do
// three things and nothing else: parse input, call one use case, render the
// result. Every business rule lives behind that call.
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/apolzek/pastebomb/internal/domain/shared"
)

// pageFrom reads the pagination query parameters. Unparsable values fall back
// to the defaults rather than failing the request, and shared.NewPage clamps
// the result so no caller can ask for an unbounded scan.
func pageFrom(ctx *gin.Context) shared.Page {
	number, _ := strconv.Atoi(ctx.Query("page"))
	size, _ := strconv.Atoi(ctx.Query("per_page"))
	if size == 0 {
		// perPage is the spelling the original API used; keep accepting it.
		size, _ = strconv.Atoi(ctx.Query("perPage"))
	}
	return shared.NewPage(number, size)
}

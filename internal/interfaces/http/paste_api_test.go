package http_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func anonymousPaste() map[string]any {
	return map[string]any{
		"title":   "deploy notes",
		"content": "kubectl rollout status deploy/api",
	}
}

func TestCreateAndReadAnAnonymousPaste(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	slug := h.createPaste(t, anonymousPaste())

	recorder := h.do(t, request{method: http.MethodGet, path: "/u/anonymous/" + slug})
	require.Equal(t, http.StatusOK, recorder.Code)

	data := dataObject(t, recorder)
	assert.Equal(t, "deploy notes", data["title"])
	assert.Equal(t, "kubectl rollout status deploy/api", data["content"])
	assert.Equal(t, false, data["has_secret"])
}

func TestReadRawServesPlainTextThatCannotBeSniffedAsHTML(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	body := anonymousPaste()
	body["content"] = "<script>alert(1)</script>"
	slug := h.createPaste(t, body)

	recorder := h.do(t, request{method: http.MethodGet, path: "/u/anonymous/" + slug + "/raw"})
	require.Equal(t, http.StatusOK, recorder.Code)

	assert.Equal(t, "<script>alert(1)</script>", recorder.Body.String())
	assert.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
	// Without these two headers a browser can decide to execute the paste.
	assert.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	assert.Contains(t, recorder.Header().Get("Content-Disposition"), "attachment")
}

// The slug is the only thing guarding an unlisted paste. If a sequential id
// also resolved, every paste in the database could be walked.
func TestPastesCannotBeFetchedByNumericID(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.createPaste(t, anonymousPaste())

	for _, id := range []string{"1", "2", "0", "-1"} {
		recorder := h.do(t, request{method: http.MethodGet, path: "/u/anonymous/" + id})
		assert.Equal(t, http.StatusNotFound, recorder.Code, "id %q must not resolve", id)
	}
}

func TestASecretProtectedPasteNeedsTheHeaderToBeRead(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	body := anonymousPaste()
	body["secret_access"] = "open sesame"
	slug := h.createPaste(t, body)

	withoutSecret := h.do(t, request{method: http.MethodGet, path: "/u/anonymous/" + slug})
	assert.Equal(t, http.StatusUnauthorized, withoutSecret.Code)
	assert.NotContains(t, withoutSecret.Body.String(), "kubectl")

	wrongSecret := h.do(t, request{
		method:  http.MethodGet,
		path:    "/u/anonymous/" + slug,
		headers: map[string]string{"X-Post-Secret": "guess"},
	})
	assert.Equal(t, http.StatusUnauthorized, wrongSecret.Code)

	correctSecret := h.do(t, request{
		method:  http.MethodGet,
		path:    "/u/anonymous/" + slug,
		headers: map[string]string{"X-Post-Secret": "open sesame"},
	})
	require.Equal(t, http.StatusOK, correctSecret.Code)
	assert.Equal(t, "kubectl rollout status deploy/api", dataObject(t, correctSecret)["content"])
}

func TestRawEndpointEnforcesTheSecretToo(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	body := anonymousPaste()
	body["secret_access"] = "open sesame"
	slug := h.createPaste(t, body)

	recorder := h.do(t, request{method: http.MethodGet, path: "/u/anonymous/" + slug + "/raw"})
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "kubectl")
}

// The public index used to return every field, including the body of pastes
// that are supposed to be unlocked with a secret first.
func TestThePublicIndexNeverExposesPasteBodies(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	protected := anonymousPaste()
	protected["secret_access"] = "open sesame"
	protected["content"] = "TOP-SECRET-PAYLOAD"
	h.createPaste(t, protected)
	h.createPaste(t, anonymousPaste())

	recorder := h.do(t, request{method: http.MethodGet, path: "/u/anonymous/posts"})
	require.Equal(t, http.StatusOK, recorder.Code)

	assert.NotContains(t, recorder.Body.String(), "TOP-SECRET-PAYLOAD")
	assert.NotContains(t, recorder.Body.String(), "kubectl")

	entries := dataArray(t, recorder)
	require.Len(t, entries, 2)
	for _, entry := range entries {
		summary := entry.(map[string]any)
		_, hasContent := summary["content"]
		assert.False(t, hasContent, "the index must not carry a content field")
		assert.NotEmpty(t, summary["url_id"])
	}
}

func TestThePublicIndexIsPaginatedAndBounded(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	for i := range 7 {
		body := anonymousPaste()
		body["title"] = fmt.Sprintf("paste %d", i)
		h.createPaste(t, body)
	}

	firstPage := h.do(t, request{method: http.MethodGet, path: "/u/anonymous/posts?page=1&per_page=3"})
	require.Equal(t, http.StatusOK, firstPage.Code)
	assert.Len(t, dataArray(t, firstPage), 3)

	// An absurd page size is clamped rather than honoured.
	huge := h.do(t, request{method: http.MethodGet, path: "/u/anonymous/posts?per_page=100000"})
	require.Equal(t, http.StatusOK, huge.Code)
	meta := decodeEnvelope(t, huge)["meta"].(map[string]any)
	assert.EqualValues(t, 100, meta["per_page"])
}

func TestCreateRejectsInvalidPastes(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	cases := map[string]map[string]any{
		"missing title":      {"content": "x"},
		"missing content":    {"title": "x"},
		"blank title":        {"title": "   ", "content": "x"},
		"past expiration":    {"title": "x", "content": "y", "expiration_date": "2020-01-01T00:00:00Z"},
		"garbled expiration": {"title": "x", "content": "y", "expiration_date": "tomorrow"},
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			recorder := h.do(t, request{method: http.MethodPost, path: "/u/anonymous/post", body: body})
			assert.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
		})
	}
}

func TestAnExpiredPasteBecomesUnreachable(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	body := anonymousPaste()
	body["expiration_date"] = time.Now().Add(900 * time.Millisecond).UTC().Format(time.RFC3339Nano)
	slug := h.createPaste(t, body)

	require.Equal(t, http.StatusOK, h.do(t, request{method: http.MethodGet, path: "/u/anonymous/" + slug}).Code)

	time.Sleep(1100 * time.Millisecond)

	// Expiry reports 404, not 410, so a stale handle reveals nothing about
	// whether it ever existed.
	assert.Equal(t, http.StatusNotFound, h.do(t, request{method: http.MethodGet, path: "/u/anonymous/" + slug}).Code)
}

func TestOwnedPastesRequireAuthenticationAndStayPrivateToTheirOwner(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	owner := h.registerUser(t, "owner", "owner@example.com")
	stranger := h.registerUser(t, "stranger", "stranger@example.com")

	anonymous := h.do(t, request{method: http.MethodPost, path: "/u/me/post", body: anonymousPaste()})
	assert.Equal(t, http.StatusUnauthorized, anonymous.Code)

	created := h.do(t, request{
		method:  http.MethodPost,
		path:    "/u/me/post",
		body:    map[string]any{"title": "private notes", "content": "ssh key rotation", "is_public": false},
		headers: bearer(h.tokenFor(t, owner)),
	})
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())

	mine := h.do(t, request{method: http.MethodGet, path: "/u/me/posts", headers: bearer(h.tokenFor(t, owner))})
	require.Equal(t, http.StatusOK, mine.Code)
	assert.Len(t, dataArray(t, mine), 1)

	theirs := h.do(t, request{method: http.MethodGet, path: "/u/me/posts", headers: bearer(h.tokenFor(t, stranger))})
	require.Equal(t, http.StatusOK, theirs.Code)
	assert.Empty(t, dataArray(t, theirs))

	// An unlisted paste must not appear in the public index either.
	index := h.do(t, request{method: http.MethodGet, path: "/u/anonymous/posts"})
	require.Equal(t, http.StatusOK, index.Code)
	assert.Empty(t, dataArray(t, index))
}

func TestOwnPastesEndpointReturnsBodies(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	owner := h.registerUser(t, "owner", "owner@example.com")
	token := h.tokenFor(t, owner)

	created := h.do(t, request{
		method:  http.MethodPost,
		path:    "/u/me/post",
		body:    map[string]any{"title": "private notes", "content": "ssh key rotation"},
		headers: bearer(token),
	})
	require.Equal(t, http.StatusCreated, created.Code)

	withContent := h.do(t, request{method: http.MethodGet, path: "/u/me/posts/content", headers: bearer(token)})
	require.Equal(t, http.StatusOK, withContent.Code)

	entries := dataArray(t, withContent)
	require.Len(t, entries, 1)
	assert.Equal(t, "ssh key rotation", entries[0].(map[string]any)["content"])
}

// An anonymous paste has no owner, so an unlisted one could never be listed
// again by anybody. The endpoint forces every anonymous paste public.
func TestAnonymousPastesAreAlwaysPublic(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	body := anonymousPaste()
	body["is_public"] = false

	recorder := h.do(t, request{method: http.MethodPost, path: "/u/anonymous/post", body: body})
	require.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, true, dataObject(t, recorder)["is_public"])
}

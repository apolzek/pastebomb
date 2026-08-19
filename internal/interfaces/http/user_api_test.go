package http_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginReturnsAUsableToken(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.registerUser(t, "ada", "ada@example.com")

	recorder := h.do(t, request{
		method: http.MethodPost,
		path:   "/login",
		body:   map[string]any{"email": "ada@example.com", "password": "correct horse battery"},
	})
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	data := dataObject(t, recorder)
	token, ok := data["token"].(string)
	require.True(t, ok)
	assert.Equal(t, "Bearer", data["token_type"])
	assert.NotEmpty(t, data["expires_at"])

	me := h.do(t, request{method: http.MethodGet, path: "/u/me", headers: bearer(token)})
	require.Equal(t, http.StatusOK, me.Code)
	assert.Equal(t, "ada@example.com", dataObject(t, me)["email"])
}

func TestLoginFailsIdenticallyForUnknownAccountsAndWrongPasswords(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.registerUser(t, "ada", "ada@example.com")

	unknown := h.do(t, request{
		method: http.MethodPost,
		path:   "/login",
		body:   map[string]any{"email": "nobody@example.com", "password": "correct horse battery"},
	})
	wrong := h.do(t, request{
		method: http.MethodPost,
		path:   "/login",
		body:   map[string]any{"email": "ada@example.com", "password": "wrong password"},
	})

	require.Equal(t, http.StatusUnauthorized, unknown.Code)
	require.Equal(t, http.StatusUnauthorized, wrong.Code)

	// Same status, same code, same message: the response tells an attacker
	// nothing about whether the address is registered. Only the correlation
	// id differs, which is per-request by construction.
	unknownBody, wrongBody := decodeEnvelope(t, unknown), decodeEnvelope(t, wrong)
	delete(unknownBody, "request_id")
	delete(wrongBody, "request_id")
	assert.Equal(t, unknownBody, wrongBody)
}

func TestProtectedRoutesRejectMissingAndMalformedTokens(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	cases := map[string]map[string]string{
		"no header":        nil,
		"empty bearer":     {"Authorization": "Bearer "},
		"wrong scheme":     {"Authorization": "Basic YWRtaW46YWRtaW4="},
		"garbage token":    {"Authorization": "Bearer not-a-token"},
		"bearer substring": {"Authorization": "NotBearer abc"},
	}

	for name, headers := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			recorder := h.do(t, request{method: http.MethodGet, path: "/u/me", headers: headers})
			assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		})
	}
}

// The old admin gate compared a plain "owner" header to a hardcoded name, so
// anyone who guessed it read every account in the database.
func TestTheOwnerHeaderNoLongerGrantsAdminAccess(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	account := h.registerUser(t, "ada", "ada@example.com")

	adminRoutes := []string{"/u/all", "/u", "/u/paginate"}

	for _, route := range adminRoutes {
		spoofed := h.do(t, request{
			method:  http.MethodGet,
			path:    route,
			headers: map[string]string{"owner": "apolzek"},
		})
		assert.Equal(t, http.StatusUnauthorized, spoofed.Code, "route %s accepted a spoofed header", route)

		// A valid token without admin rights is not enough either.
		asUser := h.do(t, request{method: http.MethodGet, path: route, headers: bearer(h.tokenFor(t, account))})
		assert.Equal(t, http.StatusForbidden, asUser.Code, "route %s accepted a non-admin user", route)
	}
}

func TestAdminRoutesAcceptATokenCarryingTheAdminClaim(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	account := h.registerUser(t, "root", "root@example.com")
	h.users.PromoteToAdmin(account.ID())

	promoted, err := h.users.ByID(t.Context(), account.ID())
	require.NoError(t, err)
	require.True(t, promoted.IsAdmin())

	recorder := h.do(t, request{method: http.MethodGet, path: "/u/all", headers: bearer(h.tokenFor(t, promoted))})
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	assert.Len(t, dataArray(t, recorder), 1)
}

// The public profile route is unauthenticated, so it must not hand out the
// personal data the previous version returned.
func TestThePublicProfileHidesEmailAndDateOfBirth(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.registerUser(t, "ada", "ada@example.com")

	for _, handle := range []string{"ada", "1"} {
		recorder := h.do(t, request{method: http.MethodGet, path: "/u/" + handle})
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

		assert.NotContains(t, recorder.Body.String(), "ada@example.com")
		assert.NotContains(t, recorder.Body.String(), "1990-01-01")

		data := dataObject(t, recorder)
		assert.Equal(t, "ada", data["username"])
		_, hasEmail := data["email"]
		assert.False(t, hasEmail)
	}
}

func TestRegisterCreatesAnAccountAndRejectsDuplicates(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	payload := map[string]any{
		"name":      "Ada Lovelace",
		"username":  "ada",
		"email":     "ada@example.com",
		"born_date": "1815-12-10",
		"password":  "correct horse battery",
	}

	created := h.do(t, request{method: http.MethodPost, path: "/u/", body: payload})
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	// The response must never echo a password or a hash back.
	assert.NotContains(t, created.Body.String(), "correct horse battery")
	assert.NotContains(t, created.Body.String(), "password")

	duplicate := h.do(t, request{method: http.MethodPost, path: "/u/", body: payload})
	assert.Equal(t, http.StatusConflict, duplicate.Code)
}

func TestRegisterRefusesToGrantAdminRightsFromThePayload(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	recorder := h.do(t, request{method: http.MethodPost, path: "/u/", body: map[string]any{
		"name":      "Eve",
		"username":  "eve",
		"email":     "eve@example.com",
		"born_date": "1990-01-01",
		"password":  "correct horse battery",
		"is_admin":  true,
		"is_active": true,
	}})
	require.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, false, dataObject(t, recorder)["is_admin"])

	// And the resulting token must not open the administrative routes.
	login := h.do(t, request{method: http.MethodPost, path: "/login", body: map[string]any{
		"email": "eve@example.com", "password": "correct horse battery",
	}})
	require.Equal(t, http.StatusOK, login.Code)

	token := dataObject(t, login)["token"].(string)
	assert.Equal(t, http.StatusForbidden, h.do(t, request{method: http.MethodGet, path: "/u/all", headers: bearer(token)}).Code)
}

func TestUpdateProfileOnlyTouchesTheAuthenticatedAccount(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	account := h.registerUser(t, "ada", "ada@example.com")

	recorder := h.do(t, request{
		method:  http.MethodPatch,
		path:    "/u/me",
		body:    map[string]any{"name": "Ada King", "username": "ada.king", "email": "ada.king@example.com", "born_date": "1815-12-10"},
		headers: bearer(h.tokenFor(t, account)),
	})
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	assert.Equal(t, "ada.king", dataObject(t, recorder)["username"])
}

// DELETE /u/me used to route into the handler that parses an :id parameter,
// which meant it always failed on the literal "me".
func TestDeleteOnSelfWorksThroughBothRoutes(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	byAlias := h.registerUser(t, "alias", "alias@example.com")
	recorder := h.do(t, request{method: http.MethodDelete, path: "/u/me", headers: bearer(h.tokenFor(t, byAlias))})
	assert.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	byID := h.registerUser(t, "byid", "byid@example.com")
	recorder = h.do(t, request{
		method:  http.MethodDelete,
		path:    "/u/2",
		headers: bearer(h.tokenFor(t, byID)),
	})
	assert.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
}

func TestDeleteRefusesToDeactivateSomeoneElse(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.registerUser(t, "victim", "victim@example.com")
	attacker := h.registerUser(t, "attacker", "attacker@example.com")

	recorder := h.do(t, request{
		method:  http.MethodDelete,
		path:    "/u/1",
		headers: bearer(h.tokenFor(t, attacker)),
	})
	assert.Equal(t, http.StatusForbidden, recorder.Code)

	victim, err := h.users.ByID(t.Context(), 1)
	require.NoError(t, err)
	assert.True(t, victim.IsActive())
}

func TestHealthAndReadinessProbes(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	assert.Equal(t, http.StatusOK, h.do(t, request{method: http.MethodGet, path: "/health"}).Code)
	assert.Equal(t, http.StatusOK, h.do(t, request{method: http.MethodGet, path: "/ready"}).Code)
}

func TestEveryResponseCarriesTheProtectiveHeadersAndARequestID(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	recorder := h.do(t, request{method: http.MethodGet, path: "/health"})

	assert.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", recorder.Header().Get("X-Frame-Options"))
	assert.Equal(t, "no-referrer", recorder.Header().Get("Referrer-Policy"))
	assert.NotEmpty(t, recorder.Header().Get("X-Request-ID"))
}

// A client-supplied request id would let a caller forge or collide log lines.
func TestTheRequestIDIsGeneratedNotEchoed(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	recorder := h.do(t, request{
		method:  http.MethodGet,
		path:    "/health",
		headers: map[string]string{"X-Request-ID": "forged-by-the-client"},
	})

	assert.NotEqual(t, "forged-by-the-client", recorder.Header().Get("X-Request-ID"))
	assert.NotEmpty(t, recorder.Header().Get("X-Request-ID"))
}

// Internal failures must never reach the client as a driver message or a
// stack trace.
func TestInternalErrorsAreReportedGenerically(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	recorder := h.do(t, request{method: http.MethodGet, path: "/u/does-not-exist"})

	require.Equal(t, http.StatusNotFound, recorder.Code)
	body := decodeEnvelope(t, recorder)
	assert.Equal(t, "user.not_found", body["code"])
	assert.NotContains(t, recorder.Body.String(), "gorm")
	assert.NotContains(t, recorder.Body.String(), "sql")
}

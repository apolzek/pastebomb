package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/domain/paste"
	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/domain/user"
	"github.com/apolzek/pastebomb/internal/infrastructure/config"
	"github.com/apolzek/pastebomb/internal/infrastructure/security"
	httpiface "github.com/apolzek/pastebomb/internal/interfaces/http"
	"github.com/apolzek/pastebomb/internal/testsupport/fake"
	"github.com/apolzek/pastebomb/internal/testsupport/memory"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	// The request logger is exercised by these tests; discarding its output
	// keeps a failing assertion readable instead of buried in access lines.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	m.Run()
}

// harness wires the real router over in-memory repositories, so these tests
// exercise the genuine middleware stack, routing table and error mapping.
type harness struct {
	engine *gin.Engine
	users  *memory.UserRepository
	tokens *security.JWTProvider
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	clock := shared.SystemClock{}
	hasher := fake.NewHasher()
	users := memory.NewUserRepository()
	pastes := memory.NewPasteRepository()
	recorder := memory.NewAuditRecorder()
	tokens := security.NewJWTProvider([]byte("a-test-secret-that-is-long-enough-32"), "pastebomb", time.Hour)

	engine, limiters := httpiface.NewRouter(httpiface.Dependencies{
		Config: config.Config{
			HTTP: config.HTTP{MaxBodyBytes: 2 << 20},
			// Generous budgets: these tests assert routing and authorization,
			// and the limiter has its own suite.
			RateLimit: config.RateLimit{
				LoginRPS: 1000, LoginBurst: 1000,
				AnonRPS: 1000, AnonBurst: 1000,
				ReadRPS: 1000, ReadBurst: 1000,
			},
			CORS: config.CORS{AllowedOrigins: []string{"http://localhost:8080"}},
		},
		AuthService: application.NewAuthService(users, hasher, tokens, recorder),
		UserService: application.NewUserService(users, hasher, clock, recorder),
		PasteService: application.NewPasteService(
			pastes,
			paste.NewFactory(security.NewRandomSlugGenerator(paste.SlugLength), hasher, clock),
			paste.NewAccessGuard(hasher, clock),
			clock,
			recorder,
		),
		TokenVerifier: tokens,
		CheckDatabase: func(context.Context) error { return nil },
	})
	t.Cleanup(limiters.Stop)

	return &harness{engine: engine, users: users, tokens: tokens}
}

type request struct {
	method  string
	path    string
	body    any
	headers map[string]string
}

func (h *harness) do(t *testing.T, req request) *httptest.ResponseRecorder {
	t.Helper()

	var payload *bytes.Buffer
	if req.body != nil {
		encoded, err := json.Marshal(req.body)
		require.NoError(t, err)
		payload = bytes.NewBuffer(encoded)
	} else {
		payload = bytes.NewBuffer(nil)
	}

	httpRequest := httptest.NewRequest(req.method, req.path, payload)
	httpRequest.Header.Set("Content-Type", "application/json")
	for key, value := range req.headers {
		httpRequest.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	h.engine.ServeHTTP(recorder, httpRequest)
	return recorder
}

// registerUser creates an account directly through the repository, bypassing
// the rate-limited public endpoint.
func (h *harness) registerUser(t *testing.T, username, email string) *user.User {
	t.Helper()

	account, err := user.Register(user.Registration{
		Name:     "Test User",
		Username: username,
		Email:    email,
		BornDate: "1990-01-01",
		Password: "correct horse battery",
	}, fake.NewHasher(), shared.SystemClock{})
	require.NoError(t, err)
	require.NoError(t, h.users.Create(context.Background(), account))

	return account
}

func (h *harness) tokenFor(t *testing.T, account *user.User) string {
	t.Helper()

	token, err := h.tokens.Issue(application.Identity{
		UserID: account.ID(),
		Email:  account.Email().String(),
		Name:   account.Name().String(),
		Admin:  account.IsAdmin(),
	})
	require.NoError(t, err)
	return token.Value
}

func bearer(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// decodeEnvelope reads the standard success body.
func decodeEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body), "body was %s", recorder.Body.String())
	return body
}

func dataObject(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	data, ok := decodeEnvelope(t, recorder)["data"].(map[string]any)
	require.True(t, ok, "expected an object under data, body was %s", recorder.Body.String())
	return data
}

func dataArray(t *testing.T, recorder *httptest.ResponseRecorder) []any {
	t.Helper()

	body := decodeEnvelope(t, recorder)
	if body["data"] == nil {
		return nil
	}
	data, ok := body["data"].([]any)
	require.True(t, ok, "expected an array under data, body was %s", recorder.Body.String())
	return data
}

// createPaste posts an anonymous paste and returns its handle.
func (h *harness) createPaste(t *testing.T, body map[string]any) string {
	t.Helper()

	recorder := h.do(t, request{method: http.MethodPost, path: "/u/anonymous/post", body: body})
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	return dataObject(t, recorder)["url_id"].(string)
}

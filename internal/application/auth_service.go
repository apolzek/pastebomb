package application

import (
	"context"
	"errors"

	"github.com/apolzek/pastebomb/internal/domain/shared"
	"github.com/apolzek/pastebomb/internal/domain/user"
)

// LoginCommand carries a sign-in attempt.
type LoginCommand struct {
	Email    string
	Password string
}

// AuthService turns credentials into tokens.
type AuthService struct {
	users  user.Repository
	hasher shared.Hasher
	tokens TokenIssuer
	audit  AuditRecorder
}

func NewAuthService(users user.Repository, hasher shared.Hasher, tokens TokenIssuer, audit AuditRecorder) *AuthService {
	return &AuthService{users: users, hasher: hasher, tokens: tokens, audit: audit}
}

// Login verifies the credentials and issues a token.
//
// Every failure path returns user.ErrInvalidCredentials and still runs a hash
// comparison shape that does not depend on whether the account exists, so the
// endpoint cannot be used to enumerate registered addresses.
func (s *AuthService) Login(ctx context.Context, cmd LoginCommand) (Token, error) {
	email, err := user.NewEmail(cmd.Email)
	if err != nil {
		s.recordFailure(ctx, cmd.Email, "malformed email")
		return Token{}, user.ErrInvalidCredentials
	}

	account, err := s.users.ByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			s.recordFailure(ctx, email.String(), "unknown account")
			return Token{}, user.ErrInvalidCredentials
		}
		return Token{}, err
	}

	if err := account.Authenticate(cmd.Password, s.hasher); err != nil {
		s.recordFailure(ctx, email.String(), "wrong password or inactive account")
		return Token{}, user.ErrInvalidCredentials
	}

	token, err := s.tokens.Issue(Identity{
		UserID: account.ID(),
		Email:  account.Email().String(),
		Name:   account.Name().String(),
		Admin:  account.IsAdmin(),
	})
	if err != nil {
		return Token{}, shared.Internal("auth.token_failed", "could not issue token").WithCause(err)
	}

	s.audit.Record(ctx, AuditEvent{
		Level:  AuditInfo,
		Actor:  account.Username().String(),
		Action: "login.succeeded",
		Details: map[string]any{
			"user_id":    int64(account.ID()),
			"expires_at": token.ExpiresAt,
		},
	})

	return token, nil
}

func (s *AuthService) recordFailure(ctx context.Context, actor, reason string) {
	s.audit.Record(ctx, AuditEvent{
		Level:   AuditWarning,
		Actor:   actor,
		Action:  "login.failed",
		Details: map[string]any{"reason": reason},
	})
}

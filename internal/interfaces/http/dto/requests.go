// Package dto defines the wire contract of the HTTP API. Binding tags only
// guard the shape of a payload; the real rules live in the domain, so a
// request that slips past binding is still rejected by an aggregate.
package dto

// LoginRequest is the sign-in payload.
type LoginRequest struct {
	Email    string `json:"email"    form:"email"    binding:"required,max=255"`
	Password string `json:"password" form:"password" binding:"required,max=72"`
}

// RegisterUserRequest is the sign-up payload.
type RegisterUserRequest struct {
	Name     string `json:"name"      form:"name"      binding:"required,max=255"`
	Username string `json:"username"  form:"username"  binding:"required,min=3,max=100"`
	Email    string `json:"email"     form:"email"     binding:"required,max=255"`
	BornDate string `json:"born_date" form:"born_date" binding:"required,len=10"`
	Password string `json:"password"  form:"password"  binding:"required,min=8,max=72"`
}

// UpdateProfileRequest replaces the mutable profile fields.
type UpdateProfileRequest struct {
	Name     string `json:"name"      form:"name"      binding:"required,max=255"`
	Username string `json:"username"  form:"username"  binding:"required,min=3,max=100"`
	Email    string `json:"email"     form:"email"     binding:"required,max=255"`
	BornDate string `json:"born_date" form:"born_date" binding:"required,len=10"`
}

// CreatePasteRequest is the paste submission payload. The JSON names predate
// the current vocabulary and are kept so existing clients keep working.
type CreatePasteRequest struct {
	Title    string `json:"title"    form:"title"    binding:"required,max=255"`
	Content  string `json:"content"  form:"content"  binding:"required,max=1048576"`
	Category string `json:"category" form:"category" binding:"max=255"`
	// IsPublic is a pointer so an omitted field keeps the endpoint default
	// instead of silently meaning "private".
	IsPublic       *bool  `json:"is_public"       form:"is_public"`
	ExpirationDate string `json:"expiration_date" form:"expiration_date" binding:"max=40"`
	SecretAccess   string `json:"secret_access"   form:"secret_access"   binding:"max=72"`
}

// PublicOr resolves IsPublic against the endpoint's default.
func (r CreatePasteRequest) PublicOr(fallback bool) bool {
	if r.IsPublic == nil {
		return fallback
	}
	return *r.IsPublic
}

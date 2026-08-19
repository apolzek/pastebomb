package persistence

import (
	"encoding/base64"
	"time"

	"github.com/apolzek/pastebomb/internal/domain/paste"
	"github.com/apolzek/pastebomb/internal/domain/user"
)

// userRecord is the row shape of the users table.
type userRecord struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name"`
	Username  string    `gorm:"column:username"`
	Email     string    `gorm:"column:email"`
	BornDate  time.Time `gorm:"column:born_date"`
	Password  string    `gorm:"column:password"`
	IsActive  bool      `gorm:"column:is_active"`
	IsAdmin   bool      `gorm:"column:is_admin"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (userRecord) TableName() string { return "users" }

func userRecordFrom(snapshot user.Snapshot) userRecord {
	return userRecord{
		ID:        snapshot.ID,
		Name:      snapshot.Name,
		Username:  snapshot.Username,
		Email:     snapshot.Email,
		BornDate:  snapshot.BornDate,
		Password:  snapshot.PasswordHash,
		IsActive:  snapshot.Active,
		IsAdmin:   snapshot.Admin,
		CreatedAt: snapshot.CreatedAt,
		UpdatedAt: snapshot.UpdatedAt,
	}
}

func (r userRecord) toAggregate() *user.User {
	return user.Rehydrate(user.Snapshot{
		ID:           r.ID,
		Name:         r.Name,
		Username:     r.Username,
		Email:        r.Email,
		BornDate:     r.BornDate,
		PasswordHash: r.Password,
		Active:       r.IsActive,
		Admin:        r.IsAdmin,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	})
}

// pasteRecord is the row shape of the posts table. The column names predate
// the "paste" vocabulary and are kept so existing data stays readable.
type pasteRecord struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Title          string     `gorm:"column:title"`
	Content        string     `gorm:"column:content"`
	Category       string     `gorm:"column:category"`
	UserID         *int64     `gorm:"column:user_id"`
	Slug           string     `gorm:"column:url_id"`
	IsPublic       bool       `gorm:"column:is_public"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	ExpirationDate *time.Time `gorm:"column:expiration_date"`
	SecretAccess   string     `gorm:"column:secret_access"`
}

func (pasteRecord) TableName() string { return "posts" }

func pasteRecordFrom(snapshot paste.Snapshot) pasteRecord {
	var owner *int64
	if snapshot.OwnerID != 0 {
		id := snapshot.OwnerID
		owner = &id
	}
	return pasteRecord{
		ID:       snapshot.ID,
		Title:    snapshot.Title,
		Content:  encodeContent(snapshot.Content),
		Category: snapshot.Category,
		UserID:   owner,
		Slug:     snapshot.Slug,
		IsPublic: snapshot.IsPublic,
		// The database stamps its own timestamps on insert; sending ours keeps
		// the aggregate and the row aligned when they are read back.
		CreatedAt:      snapshot.CreatedAt,
		UpdatedAt:      snapshot.UpdatedAt,
		ExpirationDate: snapshot.ExpiresAt,
		SecretAccess:   snapshot.SecretHash,
	}
}

func (r pasteRecord) toAggregate() *paste.Paste {
	var owner int64
	if r.UserID != nil {
		owner = *r.UserID
	}
	return paste.Rehydrate(paste.Snapshot{
		ID:         r.ID,
		Slug:       r.Slug,
		OwnerID:    owner,
		Title:      r.Title,
		Content:    decodeContent(r.Content),
		Category:   r.Category,
		IsPublic:   r.IsPublic,
		ExpiresAt:  r.ExpirationDate,
		SecretHash: r.SecretAccess,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	})
}

// encodeContent stores the body base64-encoded. This is a storage convention
// inherited from the original schema, not a security measure: it only keeps
// arbitrary bytes safe to round-trip through a TEXT column.
func encodeContent(plain string) string {
	return base64.StdEncoding.EncodeToString([]byte(plain))
}

// decodeContent reverses encodeContent, falling back to the raw value for rows
// written before the encoding was introduced.
func decodeContent(stored string) string {
	decoded, err := base64.StdEncoding.DecodeString(stored)
	if err != nil {
		return stored
	}
	return string(decoded)
}

package repository

import (
	"errors"
	"time"

	"gin-goinc-api/database"
	"gin-goinc-api/model"
	"gin-goinc-api/responses"

	"gorm.io/gorm"
)

type PostRepository struct{}

func NewPostRepository() *PostRepository {
	return &PostRepository{}
}

func (repo *PostRepository) GetAllPublicPosts() ([]responses.PostResponse, error) {
	var posts []responses.PostResponse
	err := database.DB.Table("posts").
		Where("is_public = ?", 1).
		Where("expiration_date IS NULL OR expiration_date > ?", time.Now()).
		Find(&posts).Error
	return posts, err
}

func (repo *PostRepository) CreateNewUserPost(post *model.Post) error {
	return database.DB.Table("posts").Create(post).Error
}

func (repo *PostRepository) CreatePublicPost(post *model.Post) error {
	return database.DB.Table("posts").Create(post).Error
}

func (repo *PostRepository) ListUserPosts(userID interface{}) ([]responses.ListPostResponse, error) {
	var posts []responses.ListPostResponse
	err := database.DB.Table("posts").
		Where("user_id = ?", userID).
		Where("expiration_date IS NULL OR expiration_date > ?", time.Now()).
		Select("id, title, url_id").
		Find(&posts).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return posts, err
}

func (r *PostRepository) GetUserPosts(userID uint) ([]responses.PostResponse, error) {
	var posts []responses.PostResponse
	if err := database.DB.Table("posts").
		Where("user_id = ?", userID).
		Where("expiration_date IS NULL OR expiration_date > ?", time.Now()).
		Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// GetActivePostByIdOrURL fetches a non-expired post and, separately, the
// bcrypt hash of its secret. The hash is returned through a second value
// instead of bolted onto PostResponse so it never serializes to JSON.
func (r *PostRepository) GetActivePostByIdOrURL(id string) (*responses.PostResponse, *string, error) {
	var post responses.PostResponse
	err := database.DB.Table("posts").
		Where("(id = ? OR url_id = ?)", id, id).
		Where("expiration_date IS NULL OR expiration_date > ?", time.Now()).
		First(&post).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	var secret struct {
		SecretAccess *string
	}
	if err := database.DB.Table("posts").
		Select("secret_access").
		Where("(id = ? OR url_id = ?)", id, id).
		First(&secret).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return &post, nil, err
	}

	return &post, secret.SecretAccess, nil
}

// PurgeExpiredPosts removes posts whose expiration date has already passed.
// Safe to call on a schedule; bounded delete with no fan-out.
func (r *PostRepository) PurgeExpiredPosts() (int64, error) {
	res := database.DB.Table("posts").
		Where("expiration_date IS NOT NULL AND expiration_date <= ?", time.Now()).
		Delete(&model.Post{})
	return res.RowsAffected, res.Error
}

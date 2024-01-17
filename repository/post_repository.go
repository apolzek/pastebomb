package repository

import (
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
	var postResponseStore []responses.PostResponse
	err := database.DB.Table("posts").Where("is_public = ?", 1).Find(&postResponseStore).Error
	return postResponseStore, err
}

func (repo *PostRepository) CreateNewUserPost(post *model.Post) error {
	return database.DB.Table("posts").Create(post).Error
}

func (repo *PostRepository) CreatePublicPost(post *model.Post) error {
	return database.DB.Table("posts").Create(post).Error
}

func (repo *PostRepository) ListUserPosts(userID interface{}) ([]responses.ListPostResponse, error) {
	var posts []responses.ListPostResponse
	err := database.DB.Table("posts").Where("user_id = ?", userID).Select("id, title, url_id").Find(&posts).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return posts, err
}

func (r *PostRepository) GetUserPosts(userID uint) ([]responses.PostResponse, error) {
	var postResponseStore []responses.PostResponse
	if err := database.DB.Table("posts").Where("user_id = ?", userID).Find(&postResponseStore).Error; err != nil {
		return nil, err
	}
	return postResponseStore, nil
}

func (r *PostRepository) GetPostByIdOrURL(id string) (*responses.PostResponse, error) {
	var post responses.PostResponse
	err := database.DB.Table("posts").Where("id = ? OR url_id = ?", id, id).Find(&post).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &post, err
}

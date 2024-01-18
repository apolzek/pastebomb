package repository

import (
	"gin-goinc-api/database"
	"gin-goinc-api/responses"
)

type PostRepository interface {
	GetAllPublicPosts() ([]responses.PostResponse, error)
	// CreateNewUserPost(post *dto.PostRequest, userID int8) error
	// CreatePublicPost(post *dto.PostRequest) error
	// ListUserPosts(userID int8) ([]response.ListPostResponse, error)
	// GetUserPosts(userID int8) ([]response.PostResponse, error)
	// GetPostByIdOrURL(id string) (*response.PostResponse, error)
}

type PostRepositoryImpl struct{}

func NewPostRepository() *PostRepositoryImpl {
	return &PostRepositoryImpl{}
}

func (repo *PostRepositoryImpl) GetAllPublicPosts() ([]responses.PostResponse, error) {
	var postResponseStore []responses.PostResponse
	err := database.DB.Table("posts").Where("is_public = ?", 1).Find(&postResponseStore).Error
	return postResponseStore, err
}

// func (repo *PostRepositoryImpl) CreateNewUserPost(post *dto.PostRequest, userID int8) error {
// 	// Implemente a lógica para criar um novo post de usuário no banco de dados
// }

// func (repo *PostRepositoryImpl) CreatePublicPost(post *dto.PostRequest) error {
// 	// Implemente a lógica para criar um novo post público no banco de dados
// }

// func (repo *PostRepositoryImpl) ListUserPosts(userID int8) ([]response.ListPostResponse, error) {
// 	// Implemente a lógica para obter a lista de posts de um usuário no banco de dados
// }

// func (repo *PostRepositoryImpl) GetUserPosts(userID int8) ([]response.PostResponse, error) {
// 	// Implemente a lógica para obter todos os posts de um usuário no banco de dados
// }

// func (repo *PostRepositoryImpl) GetPostByIdOrURL(id string) (*response.PostResponse, error) {
// 	// Implemente a lógica para obter um post pelo ID ou URL no banco de dados
// }

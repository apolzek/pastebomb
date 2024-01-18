package service

import (
	"gin-goinc-api/repository"
	"gin-goinc-api/responses"
)

type PostService struct {
	postRepository repository.PostRepository
}

func NewPostService(postRepo repository.PostRepository) *PostService {
	return &PostService{postRepository: postRepo}
}

func (service *PostService) GetAllPublicPosts() ([]responses.PostResponse, error) {
	posts, err := service.postRepository.GetAllPublicPosts()
	if err != nil {
		// Lidar com o erro, logar, retornar uma resposta apropriada, etc.
		return nil, err
	}

	// Aqui você pode realizar qualquer lógica adicional no serviço, se necessário

	return posts, nil
}

// func (service *PostService) CreateNewUserPost(post *dto.PostRequest, userID int8) error {
// 	// Implemente a lógica de serviço para criar um novo post de usuário
// }

// func (service *PostService) CreatePublicPost(post *dto.PostRequest) error {
// 	// Implemente a lógica de serviço para criar um novo post público
// }

// func (service *PostService) ListUserPosts(userID int8) ([]response.ListPostResponse, error) {
// 	// Implemente a lógica de serviço para obter a lista de posts de um usuário
// }

// func (service *PostService) GetUserPosts(userID int8) ([]response.PostResponse, error) {
// 	// Implemente a lógica de serviço para obter todos os posts de um usuário
// }

// func (service *PostService) GetPostByIdOrURL(id string) (*response.PostResponse, error) {
// 	// Implemente a lógica de serviço para obter um post pelo ID ou URL
// }

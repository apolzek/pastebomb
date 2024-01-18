package post_controller

import (
	"gin-goinc-api/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	postService *service.PostService
}

func NewPostController(postService *service.PostService) *PostController {
	return &PostController{postService: postService}
}

func (controller *PostController) GetAllPublicPosts(ctx *gin.Context) {
	posts, err := controller.postService.GetAllPublicPosts()
	if err != nil {
		// Lidar com o erro, retornar uma resposta de erro apropriada
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Retornar a resposta de sucesso
	ctx.JSON(http.StatusOK, gin.H{"data": posts})
}

// func (controller *PostController) CreateNewUserPost(ctx *gin.Context) {
// 	// Implemente a lógica para criar um novo post de usuário
// }

// func (controller *PostController) CreatePublicPost(ctx *gin.Context) {
// 	// Implemente a lógica para criar um novo post público
// }

// func (controller *PostController) ListUserPosts(ctx *gin.Context) {
// 	// Implemente a lógica para obter a lista de posts de um usuário
// }

// func (controller *PostController) GetUserPosts(ctx *gin.Context) {
// 	// Implemente a lógica para obter todos os posts de um usuário
// }

// func (controller *PostController) GetPostByIdOrURL(ctx *gin.Context) {
// 	// Implemente a lógica para obter um post pelo ID ou URL
// }

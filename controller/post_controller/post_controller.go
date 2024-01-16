package post_controller

import (
	"encoding/base64"
	"fmt"
	"gin-goinc-api/database"
	"gin-goinc-api/middleware"
	"gin-goinc-api/model"
	"gin-goinc-api/repository"
	"gin-goinc-api/requests"
	"gin-goinc-api/responses"
	"gin-goinc-api/utils"
	"net/http"

	// "log"

	"github.com/gin-gonic/gin"
)

func GetUserPosts(ctx *gin.Context) {

	userID, _ := ctx.Get("user_id")

	var post_response_store []responses.PostResponse

	err := database.DB.Table("posts").Where("user_id = ?", userID).Find(&post_response_store).Error

	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{
			"message": "Internal Server Error",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"data": post_response_store,
	})
}

func ListUserPosts(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	postRepository := repository.NewPostRepository()

	posts, err := postRepository.ListUserPosts(userID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": posts})
}

func GetAllPublicPosts(ctx *gin.Context) {
	postRepository := repository.NewPostRepository()

	postResponseStore, err := postRepository.GetAllPublicPosts()

	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"message": "Internal Server Error"})
		return
	}

	ctx.JSON(200, gin.H{"data": postResponseStore})
}

func CreateNewUserPost(ctx *gin.Context) {
	post_request := new(requests.PostRequest)

	userID := middleware.ExtractUserIDFromContext(ctx)

	fmt.Println(userID)
	if errReq := ctx.ShouldBind(&post_request); errReq != nil {
		ctx.JSON(400, gin.H{
			"message": errReq.Error(),
		})
		return
	}

	var user model.User
	userDeleted := database.DB.Table("users").Where("id = ? AND is_active = ?", uint(userID), 1).First(&user).Error
	if userDeleted != nil || userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		ctx.Abort() // Interrompe o processamento adicional do middleware ou da rota
		return
	}

	random_url := utils.GenerateRandomString(10)

	post := new(model.Post)
	post.Title = &post_request.Title
	contentBase64 := base64.StdEncoding.EncodeToString([]byte(post_request.Content))
	post.Content = &contentBase64
	post.Category = &post_request.Category
	post.UserID = &userID
	post.UrlID = &random_url
	post.IsPublic = &post_request.IsPublic
	post.ExpirationDate = &post_request.ExpirationDate
	post.SecretAccess = &post_request.SecretAccess

	ErrDB := database.DB.Table("posts").Create(&post).Error
	if ErrDB != nil {
		ctx.JSON(500, gin.H{
			"message": "can't create data",
		})
		return
	}

	post_response_store := responses.PostResponseStore{
		ID:       post.ID,
		Title:    post.Title,
		Content:  post.Content,
		IsPublic: post.IsPublic,
		UrlID:    post.UrlID,
	}

	ctx.JSON(200, gin.H{
		"message": "Data created successfully",
		"data":    post_response_store,
	})
}

func CreatePublicPost(ctx *gin.Context) {
	postRequest := new(requests.PostRequest)

	if errReq := ctx.ShouldBind(&postRequest); errReq != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": errReq.Error()})
		return
	}

	staticIsPublic := int8(1)
	randomURL := utils.GenerateRandomString(10)

	post := &model.Post{
		Title:    &postRequest.Title,
		Content:  utils.EncodeContent(postRequest.Content),
		Category: &postRequest.Category,
		IsPublic: &staticIsPublic,
		UrlID:    &randomURL,
	}

	postRepository := repository.NewPostRepository()
	if err := postRepository.CreatePublicPost(post); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	postResponseStore := responses.PostResponseStore{
		ID:       post.ID,
		Title:    post.Title,
		Content:  post.Content,
		Category: post.Category,
		UrlID:    post.UrlID,
		IsPublic: post.IsPublic,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Data created successfully",
		"data":    postResponseStore,
	})
}
func GetById(ctx *gin.Context) {
	id := ctx.Param("id")
	post := new(responses.PostResponse)

	errDb := database.DB.Table("posts").Where("id = ? ", id).Or("url_id = ?", id).Find(&post).Error

	if errDb != nil {
		ctx.JSON(500, gin.H{
			"messange": "Inernal server error",
		})
		return
	}

	if post.ID == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"messange": "Data post not found",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "data transmitted",
		"data":    post,
	})
}

func GetPublicPostById(ctx *gin.Context) {
	id := ctx.Param("id")
	post := new(responses.PostResponse)

	errDb := database.DB.Table("posts").
		Where("id = ? OR url_id = ?", id, id).
		Where("is_public = ?", 1).
		Find(&post).
		Error

	if errDb != nil {
		ctx.JSON(500, gin.H{
			"messange": "Inernal server error",
		})
		return
	}

	if post.ID == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"messange": "Data post not found",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "data transmitted",
		"data":    post,
	})
}

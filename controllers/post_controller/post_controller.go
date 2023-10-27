package post_controller

import (
	"gin-goinc-api/database"
	"gin-goinc-api/models"
	"gin-goinc-api/requests"
	"gin-goinc-api/responses"
	"gin-goinc-api/utils"
	"net/http"

	// "log"

	"github.com/gin-gonic/gin"
)

func GetAllPost(ctx *gin.Context) {

	// var post_response_store []responses.post_response_store
	var post_response_store []responses.PostResponse
	// users := new([]models.User)

	err := database.DB.Table("posts").Find(&post_response_store).Error

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

func Store(ctx *gin.Context) {
	post_request := new(requests.PostRequest)

	if errReq := ctx.ShouldBind(&post_request); errReq != nil {
		ctx.JSON(400, gin.H{
			"message": errReq.Error(),
		})
		return
	}

	post := new(models.Post)
	post.Title = &post_request.Title
	post.Content = &post_request.Content
	post.Author = &post_request.Author
	post.URL = &post_request.URL

	ErrDB := database.DB.Table("posts").Create(&post).Error
	if ErrDB != nil {
		ctx.JSON(500, gin.H{
			"message": "can't create data",
		})
		return
	}

	random_url := utils.GenerateRandomString(10)
	post_response_store := responses.PostResponseStore{
		ID:      post.ID,
		Title:   post.Title,
		Content: post.Content,
		Author:  post.Author,
		URL:     random_url,
	}

	ctx.JSON(200, gin.H{
		"message": "Data created successfully",
		"data":    post_response_store,
	})
}
func GetById(ctx *gin.Context) {
	id := ctx.Param("id")
	post := new(responses.PostResponse)

	errDb := database.DB.Table("posts").Where("id = ?", id).Find(&post).Error

	if errDb != nil || post.ID == nil {
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

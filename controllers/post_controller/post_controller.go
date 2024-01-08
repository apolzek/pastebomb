package post_controller

import (
	"encoding/base64"
	"fmt"
	"gin-goinc-api/database"
	"gin-goinc-api/middleware"
	"gin-goinc-api/models"
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

func GetAllPublicPosts(ctx *gin.Context) {

	// var post_response_store []responses.post_response_store
	var post_response_store []responses.PostResponse
	// users := new([]models.User)

	err := database.DB.Table("posts").Where("is_public = ?", 1).Find(&post_response_store).Error

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

	userID := middleware.ExtractUserIDFromContext(ctx)

	fmt.Println(userID)
	if errReq := ctx.ShouldBind(&post_request); errReq != nil {
		ctx.JSON(400, gin.H{
			"message": errReq.Error(),
		})
		return
	}

	random_url := utils.GenerateRandomString(10)

	post := new(models.Post)
	post.Title = &post_request.Title
	contentBase64 := base64.StdEncoding.EncodeToString([]byte(post_request.Content))
	post.Content = &contentBase64
	post.IsPublic = &post_request.IsPublic
	post.UserId = &userID
	post.UrlID = &random_url

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

func StorePublic(ctx *gin.Context) {
	post_request := new(requests.PostRequest)

	if errReq := ctx.ShouldBind(&post_request); errReq != nil {
		ctx.JSON(400, gin.H{
			"message": errReq.Error(),
		})
		return
	}

	staticIsPublic := int8(1)
	random_url := utils.GenerateRandomString(10)

	post := new(models.Post)
	post.Title = &post_request.Title
	contentBase64 := base64.StdEncoding.EncodeToString([]byte(post_request.Content))
	post.Content = &contentBase64
	post.IsPublic = &staticIsPublic
	post.UrlID = &random_url

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
		UrlID:    post.UrlID,
		IsPublic: post.IsPublic,
	}

	ctx.JSON(200, gin.H{
		"message": "Data created successfully",
		"data":    post_response_store,
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

package post_controller

import (
	"gin-goinc-api/database"
	"gin-goinc-api/models"
	"gin-goinc-api/requests"
	"gin-goinc-api/responses"

	// "log"

	"github.com/gin-gonic/gin"
)

func GetAllPost(ctx *gin.Context) {

	// var postResponse []responses.postResponse
	var postResponse []responses.PostResponse
	// users := new([]models.User)

	err := database.DB.Table("posts").Find(&postResponse).Error

	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{
			"message": "Internal Server Error",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"data": postResponse,
	})
}

func Store(ctx *gin.Context) {
	postReq := new(requests.PostRequest)

	if errReq := ctx.ShouldBind(&postReq); errReq != nil {
		ctx.JSON(400, gin.H{
			"message": errReq.Error(),
		})
		return
	}

	// userEmailExist := new(models.User)
	// database.DB.Table("users").Where("email = ?", postReq.Email).First(&userEmailExist)

	// if userEmailExist.Email != nil {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{
	// 		"message": "Email already used",
	// 	})
	// 	return
	// }

	post := new(models.Post)
	post.Title = &postReq.Title
	post.Content = &postReq.Content
	post.DateTime = &postReq.DateTime
	post.Author = &postReq.Author
	post.URL = &postReq.URL

	ErrDB := database.DB.Table("posts").Create(&post).Error
	if ErrDB != nil {
		ctx.JSON(500, gin.H{
			"message": "can't create data",
		})
		return
	}

	postResponse := responses.PostResponse{
		ID:      post.ID,
		Title:   post.Title,
		Content: post.Content,
		//DateTime: post.DateTime,
		Author: post.Author,
		URL:    post.URL,
	}

	ctx.JSON(200, gin.H{
		"message": "Data created successfully",
		"data":    postResponse,
	})
}

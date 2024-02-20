package auth_controller

import (
	"gin-goinc-api/database"
	"gin-goinc-api/model"
	"gin-goinc-api/requests"
	"gin-goinc-api/security"
	utils "gin-goinc-api/security"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Login(ctx *gin.Context) {

	loginReq := new(requests.LoginRequest)

	if errReq := ctx.ShouldBind(&loginReq); errReq != nil {
		ctx.AbortWithStatusJSON(400, gin.H{
			"message": errReq.Error(),
		})
		return
	}

	user := new(model.User)
	errUser := database.DB.Table("users").Where("email = ?", loginReq.Email).Where("is_active = ?", 1).First(&user).Error

	log.Println(errUser)

	if errUser != nil {
		ctx.AbortWithStatusJSON(404, gin.H{
			"message": "user not found",
		})
		return
	}
	if loginReq.Password != *user.Password {
		ctx.AbortWithStatusJSON(404, gin.H{
			"message": "Invalid email or password. Please try again.",
		})
		return
	}

	claims := jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}

	token, errToken := utils.GenerateToken(&claims)

	if errToken != nil {
		ctx.AbortWithStatusJSON(500, gin.H{

			"message": "failed to generate token",
		})
		return
	}

	security.LogAuditEvent("INFO", *user.Name, "successfully authenticated", "token issued and returned to user")

	ctx.JSON(200, gin.H{
		"message": "login success",
		"token":   token,
	})
}

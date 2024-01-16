package user_controller

import (
	"gin-goinc-api/database"
	"gin-goinc-api/middleware"
	"gin-goinc-api/model"
	"gin-goinc-api/repository"
	"gin-goinc-api/requests"
	"gin-goinc-api/responses"
	"strconv"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllUsers(ctx *gin.Context) {
	userRepository := repository.NewUserRepository()

	userResponse, err := userRepository.GetAllUsers()

	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"message": "Internal Server Error"})
		return
	}

	ctx.JSON(200, gin.H{"data": userResponse})
}

func GetAllActiveUsers(ctx *gin.Context) {
	var userResponse []responses.UserResponse
	userRepository := repository.NewUserRepository()

	userResponse, err := userRepository.GetAllActiveUsers()

	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{
			"message": "Internal Server Error",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"data": userResponse,
	})
}

func GetAllActiveUsersPaginate(ctx *gin.Context) {
	page := ctx.Query("page")
	perPage := ctx.Query("perPage")

	if page == "" {
		page = "1"
	}

	if perPage == "" {
		perPage = "10"
	}

	pageInt, _ := strconv.Atoi(page)
	perPageInt, _ := strconv.Atoi(perPage)

	if pageInt < 1 {
		pageInt = 1
	}

	userRepository := repository.NewUserRepository()

	userResponse, err := userRepository.GetAllActiveUsersPaginate(pageInt, perPageInt)

	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"message": "Internal Server Error"})
		return
	}

	if userResponse == nil {
		ctx.JSON(404, gin.H{"message": "No active users found"})
		return
	}

	ctx.JSON(200, gin.H{
		"data":     userResponse,
		"page":     pageInt,
		"per_page": perPageInt,
	})
}

func GetUserByIDorUsername(ctx *gin.Context) {
	id := ctx.Param("IDorUsername")
	userRepository := repository.NewUserRepository()

	user, errDb := userRepository.GetUserByIDorUsername(id)

	if errDb != nil {
		ctx.JSON(500, gin.H{"message": "Internal Server Error"})
		return
	}

	if user.ID == nil {
		ctx.JSON(404, gin.H{"message": "User not found"})
		return
	}

	ctx.JSON(200, gin.H{"message": "Data transmitted", "data": user})
}

func GetUserData(ctx *gin.Context) {

	userID, _ := ctx.Get("user_int")
	userRepository := repository.NewUserRepository()

	user, errDb := userRepository.GetUserData(userID)

	if errDb != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	if user.ID == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Data transmitted", "data": user})
}

func CreateNewUser(ctx *gin.Context) {
	user_request := new(requests.UserRequest)

	if errReq := ctx.ShouldBind(&user_request); errReq != nil {
		ctx.JSON(400, gin.H{
			"message": errReq.Error(),
		})
		return
	}

	// check email already exists
	userEmailExist := new(model.User)
	database.DB.Table("users").Where("email = ?", user_request.Email).First(&userEmailExist)
	if userEmailExist.Email != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Email already used",
		})
		return
	}

	// check username already exists
	userUsernameExist := new(model.User)
	database.DB.Table("users").Where("username = ?", user_request.Username).First(&userUsernameExist)
	if userUsernameExist.Email != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Username already used",
		})
		return
	}

	user := new(model.User)
	user.Name = &user_request.Name
	user.Username = &user_request.Username
	user.Email = &user_request.Email
	user.BornDate = &user_request.BornDate
	user.Password = &user_request.Password
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = 1

	ErrDB := database.DB.Table("users").Create(&user).Error
	if ErrDB != nil {
		ctx.JSON(500, gin.H{
			"message": "can't create data",
		})
		return
	}

	userResponse := responses.UserResponseStore{
		ID:       user.ID,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		BornDate: user.BornDate,
	}

	ctx.JSON(200, gin.H{
		"message": "Data created successfully",
		"data":    userResponse,
	})
}

func UpdateUserById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// Se houver um erro na conversão, retorna um status HTTP 400 (Bad Request)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "ID inválido",
		})
		return
	}
	tokenID := middleware.ExtractUserIDFromContext(ctx)

	if int8(id) != tokenID {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		ctx.Abort() // Interrompe o processamento adicional do middleware ou da rota
		return
	}

	user := new(model.User)
	user_request := new(requests.UserRequest)
	userEmailExist := new(model.User)

	if errReq := ctx.ShouldBind(&user_request); errReq != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": errReq.Error(),
		})
		return
	}

	errDB := database.DB.Table("users").Where("id = ?", id).Find(&user).Error
	if errDB != nil {
		ctx.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}

	if user.ID == nil {
		ctx.JSON(404, gin.H{
			"message": "Data not found",
		})
		return
	}

	//email exist
	errUserEmailExist := database.DB.Table("users").Where("email = ?", user_request.Email).Find(&userEmailExist).Error
	if errUserEmailExist != nil {
		ctx.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}

	if userEmailExist.Email != nil && *user.ID != *userEmailExist.ID {
		ctx.JSON(400, gin.H{
			"message": "Email already used",
		})
		return
	}

	user.Name = &user_request.Name
	user.Username = &user_request.Username
	user.Email = &user_request.Email
	user.BornDate = &user_request.BornDate

	errUpdate := database.DB.Table("users").Where("id = ?", id).Updates(&user).Error
	if errUpdate != nil {
		ctx.JSON(500, gin.H{
			"message": errUpdate.Error(), // tratar
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "data updated successfully",
		// "data":    user,
	})
}

func UpdateUserData(ctx *gin.Context) {
	userID, _ := ctx.Get("user_int")
	userRepository := repository.NewUserRepository()

	user, errDb := userRepository.GetUserData(userID)
	if errDb != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	if user.ID == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	userRequest := new(requests.UserRequest)
	if errReq := ctx.ShouldBind(&userRequest); errReq != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": errReq.Error()})
		return
	}

	if err := userRepository.UpdateUserData(int(*user.ID), userRequest); err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "Data not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
}

func DeactivateUserByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	userRepository := repository.NewUserRepository()

	tokenID := middleware.ExtractUserIDFromContext(ctx)
	if int8(id) != tokenID {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		ctx.Abort()
		return
	}

	user, err := userRepository.GetUserByIDorUsername(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Data not found"})
		return
	}

	if err := userRepository.DeactivateUserByID(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Data deleted successfully"})
}

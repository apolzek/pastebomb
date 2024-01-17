package user_controller

import (
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
	userRequest := new(requests.UserRequest)

	if errReq := ctx.ShouldBind(&userRequest); errReq != nil {
		ctx.JSON(400, gin.H{"message": errReq.Error()})
		return
	}

	userRepo := repository.NewUserRepository()

	// check email already exists
	if userRepo.CheckEmailExists(userRequest.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Email already used"})
		return
	}

	// check username already exists
	if userRepo.CheckUsernameExists(userRequest.Username) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Username already used"})
		return
	}

	user := &model.User{
		Name:      &userRequest.Name,
		Username:  &userRequest.Username,
		Email:     &userRequest.Email,
		BornDate:  &userRequest.BornDate,
		Password:  &userRequest.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  1,
	}

	if err := userRepo.CreateUser(user); err != nil {
		ctx.JSON(500, gin.H{"message": "Can't create data"})
		return
	}

	userResponse := responses.UserResponseStore{
		ID:       user.ID,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		BornDate: user.BornDate,
	}

	ctx.JSON(200, gin.H{"message": "Data created successfully", "data": userResponse})
}

func UpdateUserById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	tokenID := middleware.ExtractUserIDFromContext(ctx)
	if int8(id) != tokenID {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		ctx.Abort()
		return
	}

	userRepo := repository.NewUserRepository()

	userRequest := new(requests.UserRequest)
	if errReq := ctx.ShouldBind(&userRequest); errReq != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": errReq.Error()})
		return
	}

	existingUser, err := userRepo.GetUserByEmail(string(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	if existingUser.ID == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Data not found"})
		return
	}

	if userRepo.CheckUserEmailExists(userRequest.Email, *existingUser.ID) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Email already used"})
		return
	}

	existingUser.Name = &userRequest.Name
	existingUser.Username = &userRequest.Username
	existingUser.Email = &userRequest.Email
	existingUser.BornDate = &userRequest.BornDate

	err = userRepo.UpdateUser(id, existingUser)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
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

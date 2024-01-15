package user_controller

import (
	"fmt"
	"gin-goinc-api/database"
	"gin-goinc-api/middleware"
	"gin-goinc-api/models"
	"gin-goinc-api/requests"
	"gin-goinc-api/responses"
	"strconv"
	"time"

	// "log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllActiveUsers(ctx *gin.Context) {
	var userResponse []responses.UserResponse

	err := database.DB.Table("users").Where("is_active = ?", 1).Find(&userResponse).Error

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

func GetAllUsers(ctx *gin.Context) {
	var userResponse []responses.UserResponse

	err := database.DB.Table("users").Find(&userResponse).Error

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

	if page == "" {
		page = ""
	}

	perPage := ctx.Query("perPage")

	if perPage == "" {
		perPage = "10"
	}

	perPageInt, _ := strconv.Atoi(perPage)
	pageInt, _ := strconv.Atoi(page)

	if pageInt < 1 {
		pageInt = 1
	}

	var userResponse []responses.UserResponse
	// users := new([]models.User)

	err := database.DB.Table("users").Where("is_active = ?", 1).Offset((pageInt - 1) * perPageInt).Limit(perPageInt).Find(&userResponse).Error

	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{
			"message": "Internal Server Error",
		})
		return
	}
	ctx.JSON(200, gin.H{
		"data":     userResponse,
		"page":     pageInt,
		"per_page": perPageInt,
	})
}

func GetUserById(ctx *gin.Context) {

	id := ctx.Param("id")
	user := new(responses.UserResponse)

	errDb := database.DB.Table("users").Where("id = ?", id).Find(&user).Error

	if errDb != nil {
		ctx.JSON(500, gin.H{
			"messange": "Inernal server error",
		})
		return
	}

	if user.ID == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"messange": "Data user not found",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "data transmitted",
		"data":    user,
	})
}

func GetUserData(ctx *gin.Context) {

	userID, _ := ctx.Get("user_int")
	fmt.Println("ta ai o cara:", userID)
	// id := ctx.Param("id")
	user := new(responses.UserResponse)

	errDb := database.DB.Table("users").Where("id = ?", userID).Find(&user).Error

	if errDb != nil {
		ctx.JSON(500, gin.H{
			"messange": "Inernal server error",
		})
		return
	}

	if user.ID == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"messange": "Data user not found",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "data transmitted",
		"data":    user,
	})
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
	userEmailExist := new(models.User)
	database.DB.Table("users").Where("email = ?", user_request.Email).First(&userEmailExist)
	if userEmailExist.Email != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Email already used",
		})
		return
	}

	// check username already exists
	userUsernameExist := new(models.User)
	database.DB.Table("users").Where("username = ?", user_request.Username).First(&userUsernameExist)
	if userUsernameExist.Email != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Username already used",
		})
		return
	}

	user := new(models.User)
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

	user := new(models.User)
	user_request := new(requests.UserRequest)
	userEmailExist := new(models.User)

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
	fmt.Println("ta ai o cara:", userID)
	// id := ctx.Param("id")
	userResponde := new(responses.UserResponse)

	errDb := database.DB.Table("users").Where("id = ?", userID).Find(&userResponde).Error

	if errDb != nil {
		ctx.JSON(500, gin.H{
			"messange": "Inernal server error",
		})
		return
	}

	tokenID := middleware.ExtractUserIDFromContext(ctx)

	var userIDconvert int
	switch v := userID.(type) {
	case float64:
		userIDconvert = int(v)
	case int:
		userIDconvert = v
	default:
		ctx.JSON(500, gin.H{
			"message": "Tipo de ID não suportado",
		})
		return
	}

	if int8(userIDconvert) != tokenID {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		ctx.Abort() // Interrompe o processamento adicional do middleware ou da rota
		return
	}

	user := new(models.User)
	user_request := new(requests.UserRequest)
	userEmailExist := new(models.User)

	if errReq := ctx.ShouldBind(&user_request); errReq != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": errReq.Error(),
		})
		return
	}

	errDB := database.DB.Table("users").Where("id = ?", userIDconvert).Find(&user).Error
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

	errUpdate := database.DB.Table("users").Where("id = ?", userIDconvert).Updates(&user).Error
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

func DeleteUserById(ctx *gin.Context) {
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

	user := new(models.User)

	errFind := database.DB.Table("users").Where("id = ?", id).Find(&user).Error
	if errFind != nil {
		ctx.JSON(500, gin.H{
			"message": "internal server error",
		})
		return
	}

	// log.Println("user", user)

	if user.ID == nil {
		ctx.JSON(404, gin.H{
			"message": "data not found",
		})
		return
	}

	// ErrDB := database.DB.Table("users").Unscoped().Where("id = ?", id).Delete(&models.User{}).Error

	// if ErrDB != nil {
	// 	ctx.JSON(500, gin.H{
	// 		"message": "internal server error",
	// 		"error":   ErrDB.Error(),
	// 	})
	// 	return
	// }
	errUpdate := database.DB.Table("users").Where("id = ?", id).Update("is_active", 0).Error
	if errUpdate != nil {
		ctx.JSON(500, gin.H{
			"message": errUpdate.Error(), // tratar
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "data deleted successfully",
	})
}

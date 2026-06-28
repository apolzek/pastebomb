package post_controller

import (
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"gin-goinc-api/database"
	"gin-goinc-api/middleware"
	"gin-goinc-api/model"
	"gin-goinc-api/repository"
	"gin-goinc-api/requests"
	"gin-goinc-api/responses"
	"gin-goinc-api/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetUserPosts(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	var post_response_store []responses.PostResponse
	err := database.DB.Table("posts").
		Where("user_id = ?", userID).
		Where("expiration_date IS NULL OR expiration_date > ?", time.Now()).
		Find(&post_response_store).Error
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"message": "Internal Server Error"})
		return
	}
	ctx.JSON(200, gin.H{"data": post_response_store})
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

	userID := middleware.UserIDFromContext(ctx)
	if userID == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if errReq := ctx.ShouldBind(&post_request); errReq != nil {
		ctx.JSON(400, gin.H{"message": errReq.Error()})
		return
	}

	var user model.User
	if err := database.DB.Table("users").Where("id = ? AND is_active = ?", userID, 1).First(&user).Error; err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	expiration, expErr := parseExpiration(post_request.ExpirationDate)
	if expErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": expErr.Error()})
		return
	}

	secret, secretErr := hashSecretAccess(post_request.SecretAccess)
	if secretErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "can't create data"})
		return
	}

	randomURL := utils.GenerateRandomString(10)
	contentBase64 := base64.StdEncoding.EncodeToString([]byte(post_request.Content))

	post := &model.Post{
		Title:          &post_request.Title,
		Content:        &contentBase64,
		Category:       &post_request.Category,
		UserID:         &userID,
		UrlID:          &randomURL,
		IsPublic:       &post_request.IsPublic,
		ExpirationDate: expiration,
		SecretAccess:   secret,
	}

	if err := database.DB.Table("posts").Create(post).Error; err != nil {
		ctx.JSON(500, gin.H{"message": "can't create data"})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "Data created successfully",
		"data": responses.PostResponseStore{
			ID:       post.ID,
			Title:    post.Title,
			Content:  post.Content,
			IsPublic: post.IsPublic,
			UrlID:    post.UrlID,
		},
	})
}

func CreatePublicPost(ctx *gin.Context) {
	postRequest := new(requests.PostRequest)
	if errReq := ctx.ShouldBind(&postRequest); errReq != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": errReq.Error()})
		return
	}

	expiration, expErr := parseExpiration(postRequest.ExpirationDate)
	if expErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": expErr.Error()})
		return
	}

	secret, secretErr := hashSecretAccess(postRequest.SecretAccess)
	if secretErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	staticIsPublic := int8(1)
	randomURL := utils.GenerateRandomString(10)

	post := &model.Post{
		Title:          &postRequest.Title,
		Content:        utils.EncodeContent(postRequest.Content),
		Category:       &postRequest.Category,
		IsPublic:       &staticIsPublic,
		UrlID:          &randomURL,
		ExpirationDate: expiration,
		SecretAccess:   secret,
	}

	postRepository := repository.NewPostRepository()
	if err := postRepository.CreatePublicPost(post); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Data created successfully",
		"data": responses.PostResponseStore{
			ID:       post.ID,
			Title:    post.Title,
			Content:  post.Content,
			Category: post.Category,
			UrlID:    post.UrlID,
			IsPublic: post.IsPublic,
		},
	})
}

func GetPublicPostById(ctx *gin.Context) {
	id := ctx.Param("id")
	postRepository := repository.NewPostRepository()

	post, secretHash, err := postRepository.GetActivePostByIdOrURL(id)
	if err != nil {
		ctx.JSON(500, gin.H{"message": "Internal Server Error"})
		return
	}

	if post == nil || post.ID == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Data post not found"})
		return
	}

	if !checkSecretAccess(secretHash, ctx.GetHeader("X-Post-Secret")) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "secret required"})
		return
	}

	ctx.JSON(200, gin.H{"message": "Data transmitted", "data": post})
}

// GetRawPostByID returns the decoded paste body as text/plain so users can
// `curl https://.../u/anonymous/<id>/raw` like a classic pastebin.
func GetRawPostByID(ctx *gin.Context) {
	id := ctx.Param("id")
	postRepository := repository.NewPostRepository()

	post, secretHash, err := postRepository.GetActivePostByIdOrURL(id)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "internal server error")
		return
	}
	if post == nil || post.ID == nil {
		ctx.String(http.StatusNotFound, "not found")
		return
	}
	if !checkSecretAccess(secretHash, ctx.GetHeader("X-Post-Secret")) {
		ctx.String(http.StatusUnauthorized, "secret required")
		return
	}

	if post.Content == nil {
		ctx.String(http.StatusOK, "")
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(*post.Content)
	if err != nil {
		// Old rows stored raw text — fall through.
		decoded = []byte(*post.Content)
	}
	ctx.Data(http.StatusOK, "text/plain; charset=utf-8", decoded)
}

// parseExpiration accepts RFC3339 timestamps or empty (= no expiration).
// Rejects dates in the past so callers can't pre-expire a paste.
func parseExpiration(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, errors.New("expiration_date must be RFC3339 (e.g. 2026-12-31T23:59:59Z)")
	}
	if t.Before(time.Now()) {
		return nil, errors.New("expiration_date must be in the future")
	}
	return &t, nil
}

// hashSecretAccess bcrypts the user-supplied secret so a DB read does not
// leak the plaintext. Returns (nil, nil) when no secret was set.
func hashSecretAccess(raw string) (*string, error) {
	if raw == "" {
		return nil, nil
	}
	h, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	s := string(h)
	return &s, nil
}

// checkSecretAccess returns true when either the paste has no secret, or
// the supplied header matches the stored hash.
func checkSecretAccess(hash *string, supplied string) bool {
	if hash == nil || *hash == "" {
		return true
	}
	return bcrypt.CompareHashAndPassword([]byte(*hash), []byte(supplied)) == nil
}

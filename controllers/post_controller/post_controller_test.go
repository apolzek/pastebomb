package post_controller

import (
	"gin-goinc-api/configs/db_config"
	"gin-goinc-api/database"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetUserPosts(t *testing.T) {
	db_config.DB_DRIVER = "test"
	database.ConnectDatabase()

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.GET("/u/me/posts", GetUserPosts)

	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/u/me/posts", nil)
	r.ServeHTTP(w, req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

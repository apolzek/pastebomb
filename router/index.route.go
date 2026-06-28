package router

import (
	"gin-goinc-api/controller/auth_controller"
	"gin-goinc-api/controller/default_controller"
	"gin-goinc-api/controller/post_controller"
	"gin-goinc-api/controller/user_controller"
	"gin-goinc-api/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoute(app *gin.Engine) {
	route := app.Group("")

	userRoute := route.Group("u")

	// Administrative
	userRoute.GET("/all", middleware.AdminMiddleware, user_controller.GetAllUsers)
	userRoute.GET("/", middleware.AdminMiddleware, user_controller.GetAllActiveUsers)
	userRoute.GET("/paginate", middleware.AdminMiddleware, user_controller.GetAllActiveUsersPaginate)

	// Authentication — rate-limited to slow brute force.
	route.POST("/login", middleware.LoginLimiter.Handler(), auth_controller.Login)

	// Private
	userRoute.GET("/me", middleware.AuthMiddleware, user_controller.GetUserData)
	userRoute.PATCH("/me", middleware.AuthMiddleware, user_controller.UpdateUserData)
	userRoute.DELETE("/:id", middleware.AuthMiddleware, user_controller.DeactivateUserByID)
	userRoute.DELETE("/me", middleware.AuthMiddleware, user_controller.DeactivateUserByID)
	userRoute.GET("/me/posts/content", middleware.AuthMiddleware, post_controller.GetUserPosts)
	userRoute.GET("/me/posts", middleware.AuthMiddleware, post_controller.ListUserPosts)
	userRoute.POST("/me/post", middleware.AuthMiddleware, post_controller.CreateNewUserPost)

	// Public — anonymous post creation is rate-limited per IP.
	userRoute.POST("/", middleware.AnonLimiter.Handler(), user_controller.CreateNewUser)
	userRoute.GET("/anonymous/posts", post_controller.GetAllPublicPosts)
	userRoute.POST("/anonymous/post", middleware.AnonLimiter.Handler(), post_controller.CreatePublicPost)
	userRoute.GET("/anonymous/:id", post_controller.GetPublicPostById)
	userRoute.GET("/anonymous/:id/raw", post_controller.GetRawPostByID)
	userRoute.GET("/:IDorUsername", user_controller.GetUserByIDorUsername)

	route.GET("/health", default_controller.GetAllBook)

	v1Route(route)
}

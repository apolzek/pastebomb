package router

import (
	"gin-goinc-api/controller/auth_controller"
	"gin-goinc-api/controller/default_controller"
	"gin-goinc-api/controller/post_controller"
	"gin-goinc-api/controller/user_controller"
	"gin-goinc-api/middleware"
	"gin-goinc-api/repository"
	"gin-goinc-api/service"

	"github.com/gin-gonic/gin"
)

func InitRoute(app *gin.Engine) {
	route := app.Group("")
	// route.Static(app_config.STATIC_ROUTE, app_config.STATIC_DIR)

	//route user
	userRoute := route.Group("u")

	// Administrative
	userRoute.GET("/all", middleware.AdminMiddleware, user_controller.GetAllUsers)
	userRoute.GET("/", middleware.AdminMiddleware, user_controller.GetAllActiveUsers)
	userRoute.GET("/paginate", middleware.AdminMiddleware, user_controller.GetAllActiveUsersPaginate)

	// Authentication
	route.POST("/login", auth_controller.Login)

	// Private
	userRoute.GET("/me", middleware.AuthMiddlelware, user_controller.GetUserData)
	userRoute.PATCH("/me", middleware.AuthMiddlelware, user_controller.UpdateUserData)
	userRoute.DELETE("/:id", middleware.AuthMiddlelware, user_controller.DeactivateUserByID)
	userRoute.DELETE("/me", middleware.AuthMiddlelware, user_controller.DeactivateUserByID)
	// userRoute.GET("/me/posts/content", middleware.AuthMiddlelware, post_controller.GetUserPosts)
	// userRoute.GET("/me/posts", middleware.AuthMiddlelware, post_controller.ListUserPosts)
	// userRoute.POST("/me/post", middleware.AuthMiddlelware, post_controller.CreateNewUserPost)

	// Public
	userRoute.POST("/", user_controller.CreateNewUser)

	postRepo := repository.NewPostRepository()
	postService := service.NewPostService(postRepo)
	postController := post_controller.NewPostController(postService)
	userRoute.GET("/anonymous/posts", postController.GetAllPublicPosts)

	// userRoute.POST("/anonymous/post", post_controller.CreatePublicPost)
	// userRoute.GET("/anonymous/:id", post_controller.GetPublicPostById)
	userRoute.GET("/:IDorUsername", user_controller.GetUserByIDorUsername)

	//route default
	route.GET("/health", default_controller.GetAllBook)

	v1Route(route)
}

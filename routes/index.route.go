package routes

import (
	"gin-goinc-api/controllers/auth_controller"
	"gin-goinc-api/controllers/default_controller"
	"gin-goinc-api/controllers/post_controller"
	"gin-goinc-api/controllers/user_controller"
	"gin-goinc-api/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoute(app *gin.Engine) {
	route := app.Group("")
	// route.Static(app_config.STATIC_ROUTE, app_config.STATIC_DIR)

	//auth route
	route.POST("/login", auth_controller.Login)

	//route user
	userRoute := route.Group("u")

	// Public
	userRoute.POST("/", user_controller.CreateNewUser)
	userRoute.GET("/anonymous/posts", post_controller.GetAllPublicPosts)
	userRoute.POST("/anonymous/post", post_controller.StorePublicPost)
	userRoute.GET("/anonymous/:id", post_controller.GetPublicPostById)
	userRoute.GET("/:username", user_controller.GetUserById)

	// Private
	userRoute.GET("/me", middleware.AuthMiddlelware, user_controller.GetUserData)
	// userRoute.PATCH("/me", middleware.AuthMiddlelware, user_controller.UpdateUserById)
	userRoute.PATCH("/me", middleware.AuthMiddlelware, user_controller.UpdateUserData)
	userRoute.DELETE("/:id", middleware.AuthMiddlelware, user_controller.DeleteUserById)
	userRoute.GET("/me/posts", middleware.AuthMiddlelware, post_controller.GetUserPosts)
	userRoute.POST("/me/post", middleware.AuthMiddlelware, post_controller.StoreUserPost)

	// Administrative
	userRoute.GET("/", middleware.AdminMiddleware, user_controller.GetAllActiveUsers)
	userRoute.GET("/paginate", middleware.AdminMiddleware, user_controller.GetAllActiveUsersPaginate)
	userRoute.GET("/all", middleware.AdminMiddleware, user_controller.GetAllUsers)

	//route default
	route.GET("/health", default_controller.GetAllBook)

	v1Route(route)
}

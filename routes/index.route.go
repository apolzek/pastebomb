package routes

import (
	"gin-goinc-api/configs/app_config"
	"gin-goinc-api/controllers/auth_controller"
	"gin-goinc-api/controllers/default_controller"
	"gin-goinc-api/controllers/post_controller"
	"gin-goinc-api/controllers/user_controller"
	"gin-goinc-api/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoute(app *gin.Engine) {
	route := app.Group("")
	route.Static(app_config.STATIC_ROUTE, app_config.STATIC_DIR)

	//auth route
	route.POST("/login", auth_controller.Login)

	//route user
	userRoute := route.Group("user")
	userRoute.GET("/", user_controller.GetAllUser)
	userRoute.GET("/paginate", user_controller.GetUserPaginate)
	userRoute.POST("/", user_controller.Store)
	userRoute.GET("/:id", user_controller.GetById)
	userRoute.PATCH("/:id", user_controller.UpdateById)
	userRoute.DELETE("/:id", user_controller.DeleteById)

	//route default
	route.GET("/health", default_controller.GetAllBook)

	postRoute := route.Group("post")
	postRoute.GET("/", middleware.AuthMiddlelware, post_controller.GetAllPrivatePosts)
	postRoute.GET("/public", post_controller.GetAllPublicPosts)

	postRoute.POST("/", middleware.AuthMiddlelware, post_controller.Store)
	postRoute.POST("/public", post_controller.StorePublic)

	postRoute.GET("/:id", middleware.AuthMiddlelware, post_controller.GetById)
	postRoute.GET("/public/:id", post_controller.GetPublicPostById)

	v1Route(route)

	// Middlewares
	// middleware.AuthMiddlelware
	// middleware.AdminMiddleware

}

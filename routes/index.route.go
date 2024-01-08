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
	userRoute.GET("/", middleware.AdminMiddleware, user_controller.GetAllUser)
	userRoute.GET("/paginate", middleware.AdminMiddleware, user_controller.GetUserPaginate)
	userRoute.GET("/:id", middleware.AuthMiddlelware, user_controller.GetById)
	userRoute.PATCH("/:id", middleware.AuthMiddlelware, user_controller.UpdateById)
	userRoute.DELETE("/:id", middleware.AuthMiddlelware, user_controller.DeleteById)
	userRoute.POST("/", user_controller.Store)
	userRoute.GET("/me/posts", middleware.AuthMiddlelware, post_controller.GetUserPosts)
	userRoute.POST("/me/post", middleware.AuthMiddlelware, post_controller.Store)
	userRoute.GET("/anonymous/posts", post_controller.GetAllPublicPosts)
	userRoute.POST("/anonymous/post", post_controller.StorePublic)
	userRoute.GET("/anonymous/:id", post_controller.GetPublicPostById)

	//route default
	route.GET("/health", default_controller.GetAllBook)

	// postRoute := route.Group("post")

	// postRoute.POST("/", middleware.AuthMiddlelware, post_controller.Store)

	// postRoute.GET("/:id", middleware.AuthMiddlelware, post_controller.GetById)

	v1Route(route)

	// Middlewares
	// middleware.AuthMiddlelware
	// middleware.AdminMiddleware

}

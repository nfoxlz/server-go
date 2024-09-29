// main.go
package main

import (
	"server/controllers"
	"server/web"

	"github.com/kataras/iris/v12"
	// "github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/mvc"
)

func main() {
	app := iris.New() //.Default() //.New()

	// app.Logger().SetLevel("debug")
	// app.Use(logger.New(
	// 	logger.Config{
	// 		// 是否记录状态码,默认false
	// 		Status: true,
	// 		// 是否记录远程IP地址,默认false
	// 		IP: true,
	// 		// 是否呈现HTTP谓词,默认false
	// 		Method: true,
	// 		// 是否记录请求路径,默认true
	// 		Path: true,
	// 		// 是否开启查询追加,默认false
	// 		Query: true,
	// 	}))

	app.Use(web.AuthorizationInterceptor())

	mvc.New(app.Party("/api/Account")).Handle(new(controllers.AccountController))
	mvc.New(app.Party("/api/Frame")).Handle(new(controllers.FrameController))
	mvc.New(app.Party("/api/Data")).Handle(new(controllers.DataController))

	// var logger util.Logger
	// logger.Level = util.Debug
	// logger.LogInformation("Hello")
	// logger.

	app.Listen(":8080")
}

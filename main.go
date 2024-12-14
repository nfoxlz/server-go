// main.go
package main

import (
	"server/config"
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

	app.HandleDir("/public", iris.Dir(config.PublicPath)) //"D:/Projects/CompeteMIS/release/public"

	app.Get("/*filepath", func(ctx iris.Context) {
		// 这里可以添加额外的逻辑，如果需要的话
		// 例如，记录访问日志、验证用户权限等

		// 但通常情况下，静态文件处理已经足够，不需要额外的逻辑
		// Iris会自动根据请求的路径在指定的目录中查找文件并返回给客户端
	})

	// var logger util.Logger
	// logger.Level = util.Debug
	// logger.LogInformation("Hello")
	// logger.

	app.Listen(":8080")
}

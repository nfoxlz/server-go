// main.go
package main

import (
	"server/controllers"
	"server/web"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
)

func main() {
	app := iris.New() //.Default() //.New()
	app.Use(web.AuthorizationInterceptor())

	mvc.New(app.Party("/api/Account")).Handle(new(controllers.AccountController))
	mvc.New(app.Party("/api/Frame")).Handle(new(controllers.FrameController))
	mvc.New(app.Party("/api/Data")).Handle(new(controllers.DataController))

	app.Listen(":8080")
}

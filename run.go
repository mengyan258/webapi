package webapi

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	"github.com/farseer-go/fs/configure"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/webapi/controller"
	"github.com/farseer-go/webapi/minimal"
	"net/http"
)

func Run() {
	controller.Run()
	minimal.Run()
	flog.Info(http.ListenAndServe(":8888", nil))
}

func RegisterController(c controller.IController) {
	controller.Register(c)
}

func RegisterAction(route string, actionFunc any, params ...string) {
	minimal.Register(route, actionFunc, params...)
}

// Run2 webapi.Run() default run on config:FS.
// webapi.Run("localhost")
// webapi.Run(":8089")
// webapi.Run("127.0.0.1:8089")
func Run2(params ...string) {
	param := ""
	if len(params) > 0 && params[0] != "" {
		param = params[0]
	}
	if param == "" {
		param = configure.GetString("WebApi.Url")
	}

	// 启用CORS
	if config.enableCORS {
		web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
			// 可选参数"GET", "POST", "PUT", "DELETE", "OPTIONS" (*为所有)
			// 其中Options跨域复杂请求预检
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			// 指的是允许的Header的种类
			AllowHeaders: []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
			// 公开的HTTP标头列表
			ExposeHeaders: []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
			// 如果设置，则允许共享身份验证凭据，例如cookie
			AllowCredentials: true,
			// 指定可访问域名AllowOrigins
			AllowOrigins: []string{"*"},
		}))
	}

	//param = http.ClearHttpPrefix(param)
	web.BConfig.CopyRequestBody = true
	web.BConfig.WebConfig.AutoRender = false
	web.BeeApp.Run(param)
}

// SetApiPrefix 设置api前缀
func SetApiPrefix(prefix string) {
	config.apiPrefix = prefix
}

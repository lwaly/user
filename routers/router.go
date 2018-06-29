// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"user/controllers"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
)

func init() {
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/User",
			beego.NSRouter("/Register", &controllers.UserController{}, "post:Register"),
			beego.NSRouter("/Login", &controllers.UserController{}, "post:Login"),
			beego.NSRouter("/Logout", &controllers.UserController{}, "post:Logout"),
			beego.NSRouter("/ChangePasswd", &controllers.UserController{}, "post:ChangePasswd"),
			beego.NSRouter("/RetrievePassword", &controllers.UserController{}, "post:RetrievePassword"),
			beego.NSRouter("/DelRole", &controllers.UserController{}, "post:DelRole"),
			beego.NSRouter("/UpdateRole", &controllers.UserController{}, "post:UpdateRole"),
			beego.NSRouter("/AddRole", &controllers.UserController{}, "post:AddRole"),
			beego.NSRouter("/GetUserInfo", &controllers.UserController{}, "post:GetUserInfo"),
			beego.NSRouter("/GetUserListInfo", &controllers.UserController{}, "post:GetUserListInfo"),
			beego.NSRouter("/ChangeHeadSculpture", &controllers.UserController{}, "post:ChangeHeadSculpture"),
			beego.NSRouter("/GetHeadSculpture", &controllers.UserController{}, "post:GetHeadSculpture"),
		),
	)
	beego.AddNamespace(ns)
}

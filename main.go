package main

import (
	"fmt"
	"user/models"
	_ "user/routers"

	"github.com/astaxie/beego"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	//设置日志，禁止终端打印
	log := fmt.Sprintf(`{"filename":"%s/user.log"}`, beego.AppConfig.String("log"))
	beego.SetLogger("file", log)
	beego.BeeLogger.DelLogger("console")

	models.Init()
	beego.Run()
}

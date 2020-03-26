package main

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/micro/go-log"
	"github.com/micro/go-web"
	"sss/IhomeWeb/handler"

	//"sss/IhomeWeb/handler"
	//_ "sss/IhomeWeb/models"
)

func main() {
	// create new web service
	service := web.NewService(
		web.Name("go.micro.web.ihomeweb"),
		web.Version("latest"),
		web.Address(":8080"),
	)

	// initialise service
	if err := service.Init(); err != nil {
		log.Fatal(err)
	}

	// 使用路由中间件来映射页面
	router := httprouter.New()
	router.NotFound = http.FileServer(http.Dir("html"))

	// register html handler
	// service.Handle("/", http.FileServer(http.Dir("html")))
	service.Handle("/", router)

	//获取地区信息
	router.GET("/api/v1.0/areas", handler.GetArea)
	//获取session
	router.GET("/api/v1.0/session", handler.GetSession)
	//获取首页轮播图
	router.GET("/api/v1.0/house/index", handler.GetIndex)
	//获取首页轮播图
	router.GET("/api/v1.0/imagecode/:uuid", handler.GetImageCd)
	//获取短信验证码
	router.GET("/api/v1.0/smscode/:mobile", handler.GetSmsCd)

	// run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

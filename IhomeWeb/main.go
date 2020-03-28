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
	//用户注册
	router.POST("/api/v1.0/users", handler.PostRet)
	//删除session
	router.DELETE("/api/v1.0/session", handler.DelSession)
	//登录
	router.POST("/api/v1.0/sessions", handler.PostLogin)
	//获取用户信息
	router.GET("/api/v1.0/user", handler.GetUserInfo)
	//上传头像 POST
	router.POST("/api/v1.0/user/avatar",handler.PostAvatar)
	//修改用户
	router.PUT("/api/v1.0/user/name",handler.PutUserInfo)
	//更改实名认证信息
	router.POST("/api/v1.0/user/auth",handler.PostUserAuth)
	//检查实名认证信息
	router.GET("/api/v1.0/user/auth",handler.GetUserAuth)
	//获取用户发布的房源
	router.GET("/api/v1.0/user/houses",handler.GetUserHouse)
	//发布房源信息
	router.POST("/api/v1.0/houses",handler.PostHouse)
	//获取房屋详情
	router.GET("/api/v1.0/houses/:id",handler.GetHouseInfo)
	//获取房屋详情
	router.POST("/api/v1.0/houses/:id/images",handler.PostHousesImage)
	//搜索房屋
	router.GET("/api/v1.0/houses",handler.GetHouses)
	//发布订单
	router.POST("/api/v1.0/orders",handler.PostOrders)
	//查看房东/住客的订单信息
	router.GET("/api/v1.0/user/orders",handler.GetUserOrder)
	// 审核订单
	router.PUT("/api/v1.0/orders/:id/status",handler.PutOrders)
	// 发表评价
	router.PUT("/api/v1.0/orders/:id/comment",handler.PutComment)

	// run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

package handler

import (
	"context"
	"encoding/json"
	"github.com/afocus/captcha"
	"github.com/julienschmidt/httprouter"
	"github.com/micro/go-grpc"
	"github.com/micro/go-log"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"time"

	DELSESSION "sss/DelSession/proto/example"
	GETAREA "sss/GetArea/proto/example"
	example "sss/GetArea/proto/example"
	GETHOUSEINFO "sss/GetHouseInfo/proto/example"
	GETHOUSES "sss/GetHouses/proto/example"
	GETIMAGECD "sss/GetImageCd/proto/example"
	GETINDEX "sss/GetIndex/proto/example"
	GETSESSION "sss/GetSession/proto/example"
	GETSMSCD "sss/GetSmsCd/proto/example"
	GETUSERAUTH "sss/GetUserAuth/proto/example"
	GETUSERHOUSE "sss/GetUserHouse/proto/example"
	GETUSERINFO "sss/GetUserInfo/proto/example"
	GETUSERORDER "sss/GetUserOrder/proto/example"
	POSTAVATAR "sss/PostAvatar/proto/example"
	POSTHOUSE "sss/PostHouse/proto/example"
	POSTHOUSESIMAGE "sss/PostHousesImage/proto/example"
	POSTLOGIN "sss/PostLogin/proto/example"
	POSTORDERS "sss/PostOrders/proto/example"
	POSTRET "sss/PostRet/proto/example"
	POSTUSERAUTH "sss/PostUserAuth/proto/example"
	PUTCOMMENT "sss/PutComment/proto/example"
	PUTORDERS "sss/PutOrders/proto/example"
	PUTUSERINFO "sss/PutUserInfo/proto/example"
)

func ExampleCall(w http.ResponseWriter, r *http.Request) {
	// decode the incoming request as json
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

    server := grpc.NewService()
	server.Init()

	// call the backend service
	exampleClient := example.NewExampleService("go.micro.srv.GetArea", server.Client())
	rsp, err := exampleClient.GetArea(context.TODO(), &example.Request{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// we want to augment the response
	response := map[string]interface{}{
		"msg": rsp.ErrMsg,
		"ref": time.Now().UnixNano(),
	}

	// encode and write the response as json
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

//获取地区
func GetArea(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("获取地区请求客户端 url：api/v1.0/areas")
	// 创建新的grpc返回句柄
	server := grpc.NewService()
	// 服务出初始化
	server.Init()
	// 创建获取地区的服务并且返回句柄
	exampleClient := GETAREA.NewExampleService("go.micro.srv.GetArea", server.Client())
	// 调用服务并且获得返回数据
	rsp, err := exampleClient.GetArea(context.TODO(), &GETAREA.Request{})
	if err != nil {
		http.Error(w, err.Error(), 502)
		log.Log(err)
		return
	}
	// 创建返回类型的切片
	area_list := []models.Area{}
	// 循环读取服务返回的数据
	for _,value := range rsp.Data{
		tmp := models.Area{Id:int(value.Aid), Name:value.Aname, Houses:nil}
		area_list = append(area_list, tmp)
	}
	// 创建返回数据map
	response := map[string]interface{}{
		"errno": rsp.ErrNo,
		"errmsg": rsp.ErrMsg,
		"data" : area_list,
	}
	// 注意的点
	w.Header().Set("Content-Type", "application/json")
	// 将返回数据map发送给前端
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 503)
		return
	}
}

//获取GetSession
func GetSession(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("获取Session请求 url：api/v1.0/session")

	//创建服务
	service := grpc.NewService()
	service.Init()

	//创建句柄
	exampleClient := GETSESSION.NewExampleService("go.micro.srv.GetSession", service.Client())

	//获取cookie
	userlogin, err := r.Cookie("userlogin")

	//如果不存在就返回
	if err != nil{
		//创建返回数据map
		response := map[string]interface{}{
			"errno": utils.RECODE_SESSIONERR,
			"errmsg": utils.RecodeText(utils.RECODE_SESSIONERR),
		}
		w.Header().Set("Content-Type", "application/json")
		// encode and write the response as json
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), 503)
			log.Log(err)
		}
		return
	}

	// 存在就发送数据给服务
	rsp, err := exampleClient.GetSession(context.TODO(),&GETSESSION.Request{
		Sessionid: userlogin.Value,
	})
	if err != nil {
		http.Error(w, err.Error(), 502)
		log.Log(err)
		return
	}

	//将获取到的用户名返回给前端
	data := make(map[string]string)
	data["name"] = rsp.Data
	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data" : data,
	}
	w.Header().Set("Content-Type", "application/json")
	// 将返回数据map发送给前端
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 503)
	}
}

//获取首页轮播图
func GetIndex(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("获取首页轮播图请求客户端 url：api/v1.0/houses/index")

	server :=grpc.NewService()
	server.Init()

	exampleClient := GETINDEX.NewExampleService("go.micro.srv.GetIndex", server.Client())
	rsp, err := exampleClient.GetIndex(context.TODO(),&GETINDEX.Request{})
	if err != nil {
		log.Log(err)
		http.Error(w, err.Error(), 502)
		return
	}
	data := []interface{}{}
	json.Unmarshal(rsp.Max, &data)

	//创建返回数据map
	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data": data,

	}
	w.Header().Set("Content-Type", "application/json")

	// 将返回数据map发送给前端
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 503)
	}
}

//获取图片验证码
func GetImageCd(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Log("获取地区请求客户端 url：/api/v1.0/imagecode/:uuid")

	// 创建新的grpc返回句柄
	server := grpc.NewService()
	// 服务出初始化
	server.Init()
	// 创建获取验证码的服务并且返回句柄
	exampleClient := GETIMAGECD.NewExampleService("go.micro.srv.GetImageCd", server.Client())

	// 获取客户端页面传过来的uuid，该uuid用于验证码存储到redis中的key值
	uuid := ps.ByName("uuid");
	log.Log("客户端发送过来的uuid：", uuid)

	// 调用服务并且获得返回数据
	rsp, err := exampleClient.GetImageCd(context.TODO(), &GETIMAGECD.Request{
		Uuid: uuid,
	})
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}

	// 处理前端发送过来的图片信息
	var img image.RGBA
	img.Stride = int(rsp.Stride) // 图片跨步
	img.Rect.Min.X = int(rsp.Min.X)
	img.Rect.Min.Y = int(rsp.Min.Y)
	img.Rect.Max.X = int(rsp.Max.X)
	img.Rect.Max.Y = int(rsp.Max.Y)
	img.Pix = []uint8(rsp.Pix)

	var image captcha.Image
	image.RGBA = &img

	// 将验证码图片发送给web页面
	png.Encode(w, image)
}

//获取短信验证码
func GetSmsCd(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Log("获取地区请求客户端 url：/api/v1.0/smscode/:mobile")

	// 创建新的grpc返回句柄
	server := grpc.NewService()
	// 服务出初始化
	server.Init()

	// 获取客户端页面传过来的mobile
	mobile := ps.ByName("mobile");
	log.Log("客户端发送过来的mobile：", mobile)

	// 使用正则验证手机号码格式是否正确
	myreg := regexp.MustCompile(`0?(13|14|15|17|18|19)[0-9]{9}`)
	bool := myreg.MatchString(mobile)
	if (!bool) {
		resp := map[string]interface{}{
			"errno": utils.RECODE_NODATA,
			"errmsg": "手机号错误",
		}
		// 设置返回数据格式
		w.Header().Set("Content-Type", "application/json")
		// 将错误发送给前端
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), 503)
			return
		}
		log.Log("手机号错误返回")
		return
	}
	// 获取前端发送过来的参数id和text
	id := r.URL.Query()["id"][0] // 验证码的uuid
	text := r.URL.Query()["text"][0] // 用户输入的验证码
	log.Log("获取到客户端发送过来的id ", id)
	log.Log("获取到客户端发送过来的验证码", text)
	//调用服务
	// 创建获取验证码的服务并且返回句柄
	exampleClient := GETSMSCD.NewExampleService("go.micro.srv.GetSmsCd", server.Client())
	rsp, err := exampleClient.GetSmsCd(context.TODO(), &GETSMSCD.Request{
		Mobile: mobile,
		Id: id,
		Text: text,
	})
	if err != nil {
		http.Error(w, err.Error(), 502)
		log.Log(err)
		return
	}
	// 创建返回map
	resp := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
	}
	//设置返回格式
	w.Header().Set("Content-Type", "application/json")
	//将数据回发给前端
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), 503)
		log.Log(err)
	}
}

//注册
func PostRet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Log("用户注册的请求客户端 url：/api/v1.0/users")

	/*获取前端发送过来的json数据*/
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// 打印参数
	for key, value := range request {
		log.Log(key, " = ", value,reflect.TypeOf(value))
	}
	//由于前端每作所以后端进行下操作
	if request["mobile"] == "" || request["password"] == "" || request["sms_code"] == "" {
		resp := map[string]interface{}{
			"errno": utils.RECODE_NODATA,
			"errmsg": "信息有误请从新输入",
		}
		//如果不存在直接给前端返回
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), 503)
			log.Log(err)
			return
		}
		log.Log("客户端传过来的参数有误！")
		return
	}

	//创建服务
	service := grpc.NewService()
	service.Init()

	// 连接服务将数据发送给注册服务进行注册
	exampleClient := POSTRET.NewExampleService("go.micro.srv.PostRet", service.Client())
	res, err := exampleClient.PostRet(context.TODO(), &POSTRET.Request{
		Mobile: request["mobile"].(string),
		Password: request["password"].(string),
		Smscode: request["sms_code"].(string),
	})
	if err != nil {
		http.Error(w, err.Error(), 502)
		log.Log(err)
		return
	}
	resp := map[string]interface{}{
		"errno": res.Errno,
		"errmsg": res.Errmsg,
	}

	//读取cookie
	cookie,err := r.Cookie("userlogin")
	// 如果读取失败或者cookie的value中不存在，则将SessionID保存在cookie中
	if err != nil || cookie.Value == "" {
		cookie := http.Cookie{Name: "userlogin", Value: res.SessionID, Path: "/", MaxAge: 600}
		http.SetCookie(w, &cookie)
	}

	//设置回发数据格式
	w.Header().Set("Content-Type", "application/json")
	//将数据回发给前端
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), 503)
		log.Log(err)
	}
}

//删除Session
func DelSession(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("DelSession url：api/v1.0/session")

	//创建服务
	service := grpc.NewService()
	service.Init()

	//创建句柄
	exampleClient := DELSESSION.NewExampleService("go.micro.srv.DelSession", service.Client())

	//从Cookie中获取userlogin
	userlogin, err := r.Cookie("userlogin")

	//如果没有数据说明没有的登陆，则直接返回错误
	if err != nil{
		resp := map[string]interface{}{
			"errno": utils.RECODE_SESSIONERR,
			"errmsg": utils.RecodeText(utils.RECODE_SESSIONERR),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), 503)
			log.Log(err)
			return
		}
		return
	}

	rsp, err := exampleClient.DelSession(context.TODO(),&DELSESSION.Request{
		Sessionid:userlogin.Value,
	})

	if err != nil {
		http.Error(w, err.Error(), 502)
		log.Log(err)
		return
	}
	//再次读取数据
	cookie,err :=r.Cookie("userlogin")
	//删除cookie中的userlogin，代表用户登录状态被清除
	if err !=nil || ""==cookie.Value{
		return
	}else {
		cookie := http.Cookie{Name: "userlogin", Path: "/", MaxAge: -1}
		http.SetCookie(w, &cookie)
	}

	// 准备响应数据
	resp := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
	}

	// 向客户端返回数据
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), 503)
		log.Log(err)
	}
}

//登录
func PostLogin(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("PostLogin url：api/v1.0/session")

	// 获取前端post请求发送的内容
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// 打印参数
	for key, value := range request {
		log.Log(key,value, " = ", reflect.TypeOf(value))
	}

	//判断账号密码是否为空
	if request["mobile"] == "" || request["password"] =="" {
		resp := map[string]interface{}{
			"errno": utils.RECODE_PARAMERR,
			"errmsg": utils.RecodeText(utils.RECODE_PARAMERR),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), 503)
			log.Log(err)
			return
		}
		log.Log("登录数据为空")
		return
	}

	//创建连接
	service := grpc.NewService()
	service.Init()
	exampleClient := POSTLOGIN.NewExampleService("go.micro.srv.PostLogin",service.Client())
	rsp, err := exampleClient.PostLogin(context.TODO(),&POSTLOGIN.Request{
		Password:request["password"].(string),
		Mobile:request["mobile"].(string),
	})
	if err != nil {
		http.Error(w, err.Error(), 502)
		log.Log(err)
		return
	}
	//从cookie中读取登录信息
	cookie,err := r.Cookie("userlogin")
	//如果cookie中没有登录信息，则将sessionID保存到cookie中，并设置有效时间为10分钟
	if err !=nil || ""==cookie.Value{
		cookie := http.Cookie{Name: "userlogin", Value: rsp.SessionID, Path: "/", MaxAge: 600}
		http.SetCookie(w, &cookie)
	}

	//向客户端返回结果
	resp := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), 503)
		log.Log(err)
		return
	}
}

//获取用户信息
func GetUserInfo(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("GetUserInfo url：api/v1.0/user")

	//初始化服务
	service := grpc.NewService()
	service.Init()
	//创建句柄
	exampleClient := GETUSERINFO.NewExampleService("go.micro.srv.GetUserInfo", service.Client())
	//获取用户的登陆信息
	userlogin, err:=r.Cookie("userlogin")
	//判断是否成功不成功就直接返回
	if err != nil{
		resp := map[string]interface{}{
			"errno": utils.RECODE_SESSIONERR,
			"errmsg": utils.RecodeText(utils.RECODE_SESSIONERR),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), 503)
			log.Log(err)
			return
		}
		return
	}
	//成功就将信息发送给前端
	rsp, err := exampleClient.GetUserInfo(context.TODO(),&GETUSERINFO.Request{
		Sessionid: userlogin.Value,
	})
	if err != nil {
		http.Error(w, err.Error(), 502)
		log.Log(err)
		return
	}

	// 将用户信息封装成map
	data := make(map[string]interface{})
	data["user_id"] = int(rsp.UserId)
	data["name"] = rsp.Name
	data["mobile"] = rsp.Mobile
	data["real_name"] = rsp.RealName
	data["id_card"] = rsp.IdCard
	data["avatar_url"] = utils.AddDomain2Url(rsp.AvatarUrl)

	//向客户端返回查询结果
	resp := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data" : data,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), 503)
		log.Log(err)
	}
}

//上传头像
func PostAvatar(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("PostAvatar url：api/v1.0/user/avatar")

	//创建服务
	service := grpc.NewService()
	service.Init()

	//创建句柄
	exampleClient := POSTAVATAR.NewExampleService("go.micro.srv.PostAvatar", service.Client())

	//查看登陆信息
	userlogin,err := r.Cookie("userlogin")

	//如果没有登陆就返回错误
	if err != nil{
		resp := map[string]interface{}{
			"errno": utils.RECODE_SESSIONERR,
			"errmsg": utils.RecodeText(utils.RECODE_SESSIONERR),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), 503)
			log.Log(err)
			return
		}
		return
	}

	//接收前端发送过来的文集
	file, hander, err := r.FormFile("avatar")

	//判断是否接受成功
	if err != nil{
		log.Log("Postupavatar  c.GetFile(avatar) err" ,err)
		printErr(w, utils.RECODE_IOERR, utils.RecodeText(utils.RECODE_IOERR))
		return
	}

	//打印基本信息
	log.Log("上传文件大小：", hander.Size)
	log.Log("上传文件名：", hander.Filename)

	if hander.Size > 4194304 {
		log.Log("上传文件超过4Mb！")
		printErr(w, utils.DATA_TOO_LONG, utils.RecodeText(utils.DATA_TOO_LONG))
		return
	}

	// 创建字节切片，用于存储文件数据
	filebuffer:= make([]byte, hander.Size)

	// 读取文件数据
	_, err = file.Read(filebuffer)
	if err !=nil{
		log.Log("Postupavatar  file.Read(filebuffer) err" ,err)
		printErr(w, utils.RECODE_IOERR, utils.RecodeText(utils.RECODE_IOERR))
		return
	}

	//调用函数传入数据
	rsp, err := exampleClient.PostAvatar(context.TODO(),&POSTAVATAR.Request{
		SessionID: userlogin.Value,
		Filename: hander.Filename,
		Filesize: hander.Size,
		Avatar: filebuffer, // 注意：grpc传递数据默认最大为4Mb
	})

	if err != nil {
		http.Error(w, err.Error(), 502)
		log.Log(err)
		return
	}

	// url拼接然回回传数据
	data := make(map[string]interface{})
	data["avatar_url"] = utils.AddDomain2Url(rsp.AvatarUrl)

	// 向客户端返回结果
	resp := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data":data,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), 503)
		log.Log(err)
	}
}

//修改用户
func PutUserInfo(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("PutUserInfo url：api/v1.0/user/name")

	//创建服务
	service := grpc.NewService()
	service.Init()

	// 接收前端发送内容
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 调用服务
	exampleClient := PUTUSERINFO.NewExampleService("go.micro.srv.PutUserInfo", service.Client())

	// 从cookie获取用户登陆信息
	userlogin, err := r.Cookie("userlogin")
	if err != nil{
		log.Log("cookie中没有用户登录信息存在")
		printErr(w, utils.RECODE_SESSIONERR, utils.RecodeText(utils.RECODE_SESSIONERR))
		return
	}

	// 调用服务
	rsp, err := exampleClient.PutUserInfo(context.TODO(), &PUTUSERINFO.Request{
		Sessionid:userlogin.Value,
		Username:request["name"].(string),
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//接收回发数据
	data := make(map[string]interface{})
	data["name"] = rsp.Username
	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data": data,
	}
	w.Header().Set("Content-Type", "application/json")
	// 返回前端
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}
}

//修改用户的实名认证信息
func PostUserAuth(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("PostUserAuth url：api/v1.0/user/auth")

	//创建服务
	service := grpc.NewService()
	service.Init()
	// 接收前端发送内容
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// 调用服务
	exampleClient := POSTUSERAUTH.NewExampleService("go.micro.srv.PostUserAuth", service.Client())
	// 从cookie获取用户登陆信息
	userlogin, err := r.Cookie("userlogin")
	if err != nil{
		log.Log("cookie中没有用户登录信息存在")
		printErr(w, utils.RECODE_SESSIONERR, utils.RecodeText(utils.RECODE_SESSIONERR))
		return
	}

	// 调用服务
	rsp, err := exampleClient.PostUserAuth(context.TODO(), &POSTUSERAUTH.Request{
		Sessionid: userlogin.Value,
		Realname: request["real_name"].(string),
		Idcard: request["id_card"].(string),
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//接收回发数据
	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
	}
	w.Header().Set("Content-Type", "application/json")
	// 返回前端
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}
}

//检查用户实名认证
func GetUserAuth(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("GetUserAuth url：api/v1.0/user/auth")

	//获取用户的登陆信息
	userlogin, err:=r.Cookie("userlogin")

	//判断是否成功不成功就直接返回
	if err != nil{
		log.Log("cookie中没有用户登录信息")
		printErr(w, utils.RECODE_SESSIONERR, utils.RecodeText(utils.RECODE_SESSIONERR))
		return
	}

	//初始化服务
	service := grpc.NewService()
	service.Init()
	//创建句柄
	exampleClient := GETUSERAUTH.NewExampleService("go.micro.srv.GetUserAuth", service.Client())

	//成功就将信息发送给前端
	rsp, err := exampleClient.GetUserAuth(context.TODO(),&GETUSERAUTH.Request{
		Sessionid: userlogin.Value,
	})
	if err != nil {
		http.Error(w, err.Error(), 502)
		log.Log(err)
		return
	}

	// 将用户信息封装成map
	data := make(map[string]interface{})
	data["real_name"] = rsp.RealName
	data["id_card"] = rsp.IdCard

	//向客户端返回查询结果
	resp := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data" : data,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), 503)
		log.Log(err)
	}
}

//查询用户发布的房源
func GetUserHouse(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("GetUserHouses api/v1.0/user/houses")

	// 从cookie中获取登录用户信息
	cookie ,err := r.Cookie("userlogin")
	if err != nil || cookie.Value == ""{
		log.Log("cookie中没有登录用户信息")
		printErr(w, utils.RECODE_DATAERR, utils.RecodeText(utils.RECODE_DATAERR))
		return
	}

	client := grpc.NewService()

	// call the backend service
	exampleClient := GETUSERHOUSE.NewExampleService("go.micro.srv.GetUserHouse", client.Client())
	rsp, err := exampleClient.GetUserHouse(context.TODO(), &GETUSERHOUSE.Request{
		Sessionid: cookie.Value,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 将服务端返回的二进制数据流解码到切片中
	houses_list := []models.House{}
	json.Unmarshal(rsp.Mix, &houses_list)

	var houses []interface{}

	// 将json中的房屋信息添加houses切片中
	for _, value := range houses_list {
		houses = append(houses, value.To_house_info())
	}

	//创建一个data的map
	data := make(map[string]interface{})
	data["houses"] = houses

	// we want to augment the response
	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data": data,
	}
	//设置返回数据的格式
	w.Header().Set("Content-Type","application/json")
	// encode and write the response as json
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

//发布房源信息
func PostHouse(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Log("PostHouses /api/v1.0/houses")

	//获取前端post请求发送的内容
	body, _ := ioutil.ReadAll(r.Body)

	//获取cookie
	userlogin,err:=r.Cookie("userlogin")
	if err != nil{
		log.Log("从cookie中查询用户信息失败！")
		printErr(w, utils.RECODE_SESSIONERR, utils.RecodeText(utils.RECODE_SESSIONERR))
		return
	}
	//创建连接
	service := grpc.NewService()
	service.Init()
	// 调用服务
	exampleClient := POSTHOUSE.NewExampleService("go.micro.srv.PostHouse",service.Client())
	rsp, err := exampleClient.PostHouse(context.TODO(),&POSTHOUSE.Request{
		Sessionid: userlogin.Value,
		Max: body,
	})
	if err != nil {
		http.Error(w, err.Error(), 502)
		log.Log(err)
		return
	}
	// 得到新增房源的ID
	houseid_map := make(map[string] interface{})
	houseid_map["house_id"] = int(rsp.House_Id)

	// 向页面返回结果
	resp := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data":houseid_map,

	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), 503)
		log.Log(err)
	}
}

//获取房源详情
func GetHouseInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Log("GetHouseInfo /api/v1.0/houses/:id")

	//创建服务
	server := grpc.NewService()
	server.Init()

	// 获取房屋id参数
	id := ps.ByName("id")
	log.Log("准备查询的房源：", id)

	// 从cookie中获取登录登录信息
	userlogin,err := r.Cookie("userlogin")
	if err != nil{
		log.Log("cookie中找不到用户登录信息")
		printErr(w, utils.RECODE_SESSIONERR, utils.RecodeText(utils.RECODE_SESSIONERR))
		return
	}

	// 调用服务
	// call the backend service
	exampleClient := GETHOUSEINFO.NewExampleService("go.micro.srv.GetHouseInfo", server.Client())
	rsp, err := exampleClient.GetHouseInfo(context.TODO(), &GETHOUSEINFO.Request{
		Sessionid: userlogin.Value,
		Houseid: id,
	})

	if err != nil {
		http.Error(w, err.Error(), 502)
		log.Log(err)
		return
	}

	// 把json数据封装House
	house := models.House{}
	json.Unmarshal(rsp.Housedata, &house)
	// 把House对象转换成map
	data_map := make(map[string]interface{})
	data_map["user_id"] = int(rsp.Userid)
	data_map["house"] = house.To_one_house_desc()
	// 返回结果给页面
	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data":data_map,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 501)
	}
}

//获取房源详情
func PostHousesImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Log("PostHousesImage /api/v1.0/houses/:id/images")

	//创建服务
	server :=grpc.NewService()
	server.Init()

	//获取houserid
	houseid := ps.ByName("id")

	//从cookie获取用户登录信息
	userlogin,err := r.Cookie("userlogin")
	if err != nil{
		log.Log("cookie中没有找到用户登录信息")
		printErr(w, utils.RECODE_SESSIONERR, utils.RecodeText(utils.RECODE_SESSIONERR))
		return
	}

	// 获取上传文件
	file, header, err := r.FormFile("house_image")
	if err != nil{
		log.Log("Postupavatar c.GetFile(avatar) err" ,err)
		printErr(w, utils.RECODE_IOERR, utils.RecodeText(utils.RECODE_IOERR))
		return
	}

	log.Log("上传文件大小: ",header.Size)
	log.Log("上传文件名: ",header.Filename)

	// 将上传文件内容读取出来，保存到filebuffer中
	filebuffer := make([]byte, header.Size)
	_,err = file.Read(filebuffer)
	if err !=nil{
		log.Log("Postupavatar   file.Read(filebuffer) err" ,err)
		printErr(w, utils.RECODE_IOERR, utils.RecodeText(utils.RECODE_IOERR))
		return
	}

	// call the backend service
	exampleClient := POSTHOUSESIMAGE.NewExampleService("go.micro.srv.PostHousesImage", server.Client())
	rsp, err := exampleClient.PostHousesImage(context.TODO(), &POSTHOUSESIMAGE.Request{
		Sessionid: userlogin.Value,
		Id: houseid,
		Image: filebuffer,
		Filesize: header.Size,
		Filename: header.Filename,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 将上传后的图片url返回给页面
	data := make(map[string]interface{})
	data["url"] = utils.AddDomain2Url(rsp.Url)
	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data":data,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}
}

//搜索房屋
func GetHouses(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Log("GetHouses /api/v1.0/houses/")

	server :=grpc.NewService()
	server.Init()

	// call the backend service
	exampleClient := GETHOUSES.NewExampleService("go.micro.srv.GetHouses", server.Client())

	//aid=5&sd=2017-11-12&ed=2017-11-30&sk=new&p=1
	aid := r.URL.Query()["aid"][0] //aid=5   地区编号
	sd := r.URL.Query()["sd"][0] //sd=2017-11-1   开始时间
	ed := r.URL.Query()["ed"][0] //ed=2017-11-3   结束时间
	sk := r.URL.Query()["sk"][0] //sk=new    第三栏条件
	p := r.URL.Query()["p"][0] //tp=1   当前页码

	rsp, err := exampleClient.GetHouses(context.TODO(), &GETHOUSES.Request{
		Aid:aid,
		Sd:sd,
		Ed:ed,
		Sk:sk,
		P:p,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	houses_list := []interface{}{}
	json.Unmarshal(rsp.Houses, &houses_list)

	data := map[string]interface{}{}
	data["current_page"] = rsp.CurrentPage
	data["houses"] = houses_list
	data["total_page"] = rsp.TotalPage

	// 向页面返回结果
	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data":data,
	}
	w.Header().Set("Content-Type", "application/json")
	// encode and write the response as json
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 501)
	}
}

//发布订单
func PostOrders(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Log("PostOrders /api/v1.0/orders")

	userlogin, err := r.Cookie("userlogin")
	if err != nil||userlogin.Value==""{
		log.Log("cookie没有用户登录数据")
		printErr(w, utils.RECODE_SESSIONERR, utils.RecodeText(utils.RECODE_SESSIONERR))
		return
	}

	service := grpc.NewService()
	service.Init()

	//创建ExampleService对象，用于调用服务
	exampleClient := POSTORDERS.NewExampleService("go.micro.srv.PostOrders", service.Client())

	// 获取参数
	body, _ := ioutil.ReadAll(r.Body)

	// 调用服务
	rsp, err := exampleClient.PostOrders(context.TODO(), &POSTORDERS.Request{
		Sessionid: userlogin.Value,
		Body: body,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 获取订单ID，然后存入map中
	orderid_map := make(map[string]interface{})
	orderid_map["order_id"] = int(rsp.OrderId)

	// 把订单ID发送给页面
	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data": orderid_map,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 501)
	}
}

//查看房东/住客订单
func GetUserOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Log("GetUserOrder /api/v1.0/user/orders")

	server :=grpc.NewService()
	server.Init()

	//获取cookie
	userlogin,err := r.Cookie("userlogin")
	if err != nil{
		log.Log("cookie中无法找到用户信息")
		printErr(w, utils.RECODE_SESSIONERR, utils.RecodeText(utils.RECODE_SESSIONERR))
		return
	}

	//获取请求参数role
	role := r.URL.Query()["role"][0] //role

	// 调用服务
	exampleClient := GETUSERORDER.NewExampleService("go.micro.srv.GetUserOrder", server.Client())
	rsp, err := exampleClient.GetUserOrder(context.TODO(), &GETUSERORDER.Request{
		Sessionid: userlogin.Value,
		Role: role,
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 将订单的json数据封装到map中
	order_list := []interface{}{}
	json.Unmarshal(rsp.Orders, &order_list)

	data := map[string]interface{}{}
	data["orders"] = order_list

	// 封装响应结果
	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
		"data":data,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 501)
	}
}

// 接单
func PutOrders(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Log("PutOrders /api/v1.0/orders")

	// 解析请求参数，并保存到map中
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//获取cookie
	userlogin,err := r.Cookie("userlogin")
	if err != nil{
		log.Log("Cookie中没有找到用户信息")
		printErr(w, utils.RECODE_SESSIONERR, utils.RecodeText(utils.RECODE_SESSIONERR))
		return
	}

	server:=grpc.NewService()
	server.Init()

	var reason string
	if request["reason"] != nil {
		reason = request["reason"].(string)
	}

	// call the backend service
	exampleClient := PUTORDERS.NewExampleService("go.micro.srv.PutOrders", server.Client())
	rsp, err := exampleClient.PutOrders(context.TODO(), &PUTORDERS.Request{
		Sessionid: userlogin.Value,
		Action: request["action"].(string),
		Orderid: ps.ByName("id"),
		Reason: reason,
	})

	if err != nil {
		http.Error(w, err.Error(), 503)
		return
	}

	// 处理响应结果
	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 504)
	}
}

// 发表评价
func PutComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Log("PutComment /api/v1.0/user/orders")

	// 将请求参数封装到map中
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	service := grpc.NewService()
	service.Init()

	//获取cookie
	userlogin,err:=r.Cookie("userlogin")
	if err != nil{
		log.Log("cookie中没有找到登录用户信息")
		printErr(w, utils.RECODE_SESSIONERR, utils.RecodeText(utils.RECODE_SESSIONERR))
		return
	}

	// 调用服务
	exampleClient := PUTCOMMENT.NewExampleService("go.micro.srv.PutComment", service.Client())
	rsp, err := exampleClient.PutComment(context.TODO(), &PUTCOMMENT.Request{
		Sessionid:userlogin.Value,
		Comment:request["comment"].(string),
		Orderid: ps.ByName("id"),
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	response := map[string]interface{}{
		"errno": rsp.Errno,
		"errmsg": rsp.Errmsg,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 501)
	}
}

func printErr(w http.ResponseWriter, errno string, errmsg string) {
	resp := map[string]interface{}{
		"errno": errno,
		"errmsg": errmsg,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), 503)
		log.Log(err)
	}
}
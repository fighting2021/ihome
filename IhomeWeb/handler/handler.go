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
	"net/http"
	"regexp"
	GETAREA "sss/GetArea/proto/example"
	example "sss/GetArea/proto/example"
	GETIMAGECD "sss/GetImageCd/proto/example"
	GETSMSCD "sss/GetSmsCd/proto/example"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"time"
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
	// we want to augment the response
	response := map[string]interface{}{
		"errno": utils.RECODE_SESSIONERR,
		"errmsg": utils.RecodeText(utils.RECODE_SESSIONERR),
	}
	// encode and write the response as json
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

//获取首页轮播图
func GetIndex(w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	log.Log("获取地区请求客户端 url：api/v1.0/areas")
	// we want to augment the response
	response := map[string]interface{}{
		"errno": utils.RECODE_OK,
		"errmsg": utils.RecodeText(utils.RECODE_OK),
	}
	//会传数据的时候三直接发送过去的并没有设置数据格式
	w.Header().Set("Content-Type","application/json")
	// encode and write the response as json
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 500)
		return
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
		return
	}

}


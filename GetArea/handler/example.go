package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/micro/go-log"
	example "sss/GetArea/proto/example"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"time"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) GetArea(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log(" GetArea api/v1.0/areas !!!")

	// 初始化返回值
	rsp.ErrNo = utils.RECODE_OK
	rsp.ErrMsg = utils.RecodeText(rsp.ErrNo)

	// 连接redis创建句柄
	redis_config_map := map[string]string{
		"key": utils.G_server_name,
		"conn": utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNum": utils.G_redis_dbnum,
	}

	// 将map转化为json
	redis_config_json,_ := json.Marshal(redis_config_map)
	// 连接redis
	bm, err := cache.NewCache("redis", string(redis_config_json))
	if err != nil {
		log.Fatal("缓存创建失败", err)
		rsp.ErrNo = utils.RECODE_DBERR
		rsp.ErrMsg = utils.RecodeText(rsp.ErrNo)
		return nil
	}

	// 读取Redis的缓存数据
	areas_info_value := bm.Get("areas_info")

	//如果不为空则说明成功
	if areas_info_value != nil {
		log.Log("获取到缓存发送给前端")
		// 创建map用来存放解码的json
		ares_info_map := []map[string]interface{}{}
		// 解码
		json.Unmarshal(areas_info_value.([]byte), &ares_info_map)
		// 进行循环赋值
		for _, value := range ares_info_map {
			// 创建对于数据类型并进行赋值
			area := example.Response_Address{Aid :int32(value["aid"].(float64)),
				Aname :value["aname"].(string)}
			// 递增到切片
			rsp.Data = append(rsp.Data, &area)
		}
		return nil
	}

	log.Log("没有拿到缓存，准备从mysql数据库中读取数据...")

	// 创建orm句柄
	o := orm.NewOrm()
	// 定义一个切片，用于存储读取到的地址数据
	var areas []models.Area
	// 执行查询
	num, err := o.QueryTable("area").All(&areas)
	if err != nil {
		rsp.ErrNo = utils.RECODE_DBERR
		rsp.ErrMsg = utils.RecodeText(rsp.ErrNo)
		return nil
	}
	if num == 0 {
		rsp.ErrNo = utils.RECODE_NODATA
		rsp.ErrMsg = utils.RecodeText(rsp.ErrNo)
		return nil
	}

	log.Log("把地区数据写入缓存.")
	// 将查询到的数据编码成json格式
	ares_info_str, _ := json.Marshal(areas)
	// 写入缓存
	err = bm.Put("areas_info", ares_info_str, time.Second * 3600)
	if err != nil {
		log.Log("数据库中查出数据信息存入缓存中失误",err)
		rsp.ErrNo = utils.RECODE_NODATA
		rsp.ErrMsg = utils.RecodeText(rsp.ErrNo)
		return nil
	}
	// 返回地区给页面
	for _, value := range areas {
		area := example.Response_Address{Aid :int32(value.Id), Aname:string(value.Name)}
		rsp.Data = append(rsp.Data, &area)
	}
	return nil
}

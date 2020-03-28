package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"github.com/micro/go-log"

	example "sss/GetUserHouse/proto/example"
	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) GetUserHouse(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("GetUserHouse api/v1.0/user/houses")

	rsp.Errno= utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	// 获取sessionid
	sessionid := req.Sessionid

	// redis数据库参数
	redis_conf := map[string]string{
		"key":utils.G_server_name,
		//127.0.0.1:6379
		"conn":utils.G_redis_addr+":"+utils.G_redis_port,
		"dbNum":utils.G_redis_dbnum,
	}
	//将map进行转化成为json
	redis_conf_js,_ :=json.Marshal(redis_conf)

	//创建redis句柄
	bm ,err := cache.NewCache("redis", string(redis_conf_js))
	if err!=nil{
		log.Log("redis连接失败", err)
		rsp.Errno= utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 准备查询的key值
	sessionid_user_id := sessionid + "user_id"
	// 查询对应的user_id
	value_id := bm.Get(sessionid_user_id)
	// 转换格式
	id := int(value_id.([]uint8)[0])

	// 从mysql数据库中查询用户的房屋信息
	o := orm.NewOrm()
	qs := o.QueryTable("house")

	houses_list:= []models.House{}
	// 获得当前用户房屋信息
	_,err = qs.Filter("user_id", id).All(&houses_list)
	if err!=nil {
		log.Log("查询房屋数据失败",err)
		rsp.Errno= utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	// 转换格式
	house, _ := json.Marshal(houses_list)
	rsp.Mix =house
	return nil
}

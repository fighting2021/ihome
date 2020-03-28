package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/go-log/log"
	"reflect"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"strconv"
	"time"

	example "sss/GetHouseInfo/proto/example"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) GetHouseInfo(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("GetHouseInfo  api/v1.0/houses/:id ")

	rsp.Errno  =  utils.RECODE_OK
	rsp.Errmsg  = utils.RecodeText(rsp.Errno)

	// redis数据库参数
	redis_config_map := map[string]string{
		"key":utils.G_server_name,
		"conn":utils.G_redis_addr+":"+utils.G_redis_port,
		"dbNum":utils.G_redis_dbnum,
	}
	redis_config ,_:=json.Marshal(redis_config_map)

	//连接redis数据库 创建句柄
	bm, err := cache.NewCache("redis", string(redis_config) )
	if err != nil {
		log.Log("缓存创建失败",err)
		rsp.Errno  =  utils.RECODE_DBERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return  nil
	}
	// 获取登录用户ID
	sessioniduserid :=  req.Sessionid + "user_id"
	value_id :=bm.Get(sessioniduserid)
	log.Log(value_id,reflect.TypeOf(value_id))
	id :=  int(value_id.([]uint8)[0])

	// 获取房源ID
	houseid,_ := strconv.Atoi(req.Houseid)

	// 查询redis中是否有查询房源的信息，
	// 如果有则返回给web服务端，如果没有则查询数据库
	house_info_key := fmt.Sprintf("house_info_%s", houseid)
	house_info_value := bm.Get(house_info_key)
	if house_info_value!=nil{
		rsp.Userid= int64(id)
		rsp.Housedata= house_info_value.([]byte)
		return nil
	}

	//创建数据对象
	house := models.House{Id:houseid}
	//创建数据库句柄
	o:= orm.NewOrm()
	o.Read(&house)
	/*关联查询 area user images fac等表*/
	o.LoadRelated(&house,"Area")
	o.LoadRelated(&house,"User")
	o.LoadRelated(&house,"Images")
	o.LoadRelated(&house,"Facilities")

	log.Log("查询到的房源详情：", house)

	// 将查询结果保存在redis中
	housemix ,err := json.Marshal(house)
	bm.Put(house_info_key, housemix, time.Second*3600)

	// 返回数据给web服务端
	rsp.Userid= int64(id)
	rsp.Housedata= housemix
	return nil
}

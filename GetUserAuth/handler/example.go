package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/go-log/log"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"

	example "sss/GetUserAuth/proto/example"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) GetUserAuth(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("GET /api/v1.0/user/auth GetUserAuth()")
	//错误码
	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	// redis配置信息
	redis_config_map := map[string]string{
		"key":utils.G_server_name,
		//"conn":"127.0.0.1:6379",
		"conn":utils.G_redis_addr+":"+utils.G_redis_port,
		"dbNum":utils.G_redis_dbnum,
	}
	redis_config ,_:=json.Marshal(redis_config_map)
	// 创建redis句柄
	bm, err := cache.NewCache("redis", string(redis_config) )
	if err != nil {
		log.Log("缓存创建失败",err)
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	// 准备查询的key值
	sessioniduserid := req.Sessionid + "user_id"
	// 从redis中获取用户ID
	value_id := bm.Get(sessioniduserid)
	// 将用户ID的格式转换成int类型
	id := int(value_id.([]uint8)[0])

	// 创建数据库orm句柄
	o := orm.NewOrm()
	// 根据用户ID查询
	user := models.User{Id: id}
	err = o.Read(&user)
	if err !=nil {
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
	}

	//将查询到的数据依次赋值
	rsp.RealName = user.Real_name
	rsp.IdCard = user.Id_card
	return nil
}

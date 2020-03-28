package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/go-log/log"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"time"

	example "sss/PutUserInfo/proto/example"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) PutUserInfo(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("PUT /api/v1.0/user/name PutUersinfo()")

	//创建返回空间
	rsp.Errno= utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	/*从从sessionid获取当前的userid*/
	//连接redis
	redis_config_map := map[string]string{
		"key":utils.G_server_name,
		//"conn":"127.0.0.1:6379",
		"conn":utils.G_redis_addr+":"+utils.G_redis_port,
		"dbNum":utils.G_redis_dbnum,
	}
	redis_config ,_:=json.Marshal(redis_config_map)
	//连接redis数据库 创建句柄
	bm, err := cache.NewCache("redis", string(redis_config) )
	if err != nil {
		log.Log("缓存创建失败",err)
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	//拼接key
	sessioniduserid := req.Sessionid + "user_id"
	//获取userid
	value_id := bm.Get(sessioniduserid)
	id := int(value_id.([]uint8)[0])

	//创建表对象
	user := models.User{ Id: id, Name: req.Username }

	//创建数据库句柄
	o:= orm.NewOrm()

	//更新
	_ , err = o.Update(&user ,"name")
	if err !=nil{
		rsp.Errno= utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 更新缓存
	sessionidname := req.Sessionid + "name"
	bm.Put(sessioniduserid, string(user.Id), time.Second * 600)
	bm.Put(sessionidname, string(user.Name), time.Second * 600)

	// 向web返回数据
	rsp.Username = user.Name
	return nil
}

package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/micro/go-log"
	"sss/IhomeWeb/utils"

	example "sss/DelSession/proto/example"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) DelSession(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log(" DELETE session  /api/v1.0/session !!!")
	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	redis_config_map := map[string]string{
		"key": utils.G_server_name,
		"conn": utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNum": utils.G_redis_dbnum,
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

	// 准备要删除数据的key
	sessionidname := req.Sessionid + "name"
	sessioniduserid := req.Sessionid + "user_id"
	sessionidmobile := req.Sessionid + "mobile"

	//从redis数据库中删除用户登录数据
	bm.Delete(sessionidname)
	bm.Delete(sessioniduserid)
	bm.Delete(sessionidmobile)

	return nil
}

package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/garyburd/redigo/redis"
	"github.com/go-log/log"
	"reflect"
	"sss/IhomeWeb/utils"

	example "sss/GetSession/proto/example"
	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) GetSession(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log(" GET session  /api/v1.0/session !!!")

	rsp.Errno = utils.RECODE_SESSIONERR
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	//redis数据库配置信息
	redis_config_map := map[string]string{
		"key":   utils.G_server_name,
		"conn":  utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNum": utils.G_redis_dbnum,
	}
	redis_config, _ := json.Marshal(redis_config_map)

	//连接redis数据库 创建句柄
	bm, err := cache.NewCache("redis", string(redis_config))
	if err != nil {
		log.Log("缓存创建失败", err)
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}

	//从缓存中获取登录用户名
	key := req.Sessionid + "name"
	value := bm.Get(key)
	name, err := redis.String(value, nil)

	if err != nil {
		rsp.Errno = utils.RECODE_DATAERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return err
	}
	log.Log("从redis中获取到的登录用户", reflect.TypeOf(name))

	//获取到了session
	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)
	rsp.Data = name
	return nil
}

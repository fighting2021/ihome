package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/go-log/log"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"time"

	example "sss/PostLogin/proto/example"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) PostLogin(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("登陆 api/v1.0/sessions")

	//返回给前端的map结构体
	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	//查询数据库
	var user models.User
	o:=orm.NewOrm()
	//创建查询句柄
	qs:=o.QueryTable("user")
	//查询符合的数据
	err := qs.Filter("mobile", req.Mobile).One(&user)
	if err != nil {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	//判断密码是否正确
	if req.Password != user.Password_hash{
		rsp.Errno = utils.RECODE_PWDERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	//将用户登录信息写入redis数据库中
	redis_config_map := map[string]string{
		"key": utils.G_server_name,
		"conn": utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNum": utils.G_redis_dbnum,
	}
	redis_config ,_:=json.Marshal(redis_config_map)
	bm, err := cache.NewCache("redis", string(redis_config) )
	if err != nil {
		log.Log("缓存创建失败", err)
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	//生成sessionID
	h := GetMd5String(req.Mobile+req.Password)
	rsp.SessionID = h
	//拼接key sessionid + name
	bm.Put(h + "name",string(user.Name),time.Second * 3600)
	//拼接key sessionid + user_id
	bm.Put(h + "user_id", string(user.Id) ,time.Second * 3600)
	//拼接key sessionid + mobile
	bm.Put(h + "mobile",string(user.Mobile) ,time.Second * 3600)
	return nil
}

func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

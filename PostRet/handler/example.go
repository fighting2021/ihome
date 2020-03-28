package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/micro/go-log"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	example "sss/PostRet/proto/example"
	"strconv"
	"time"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}


// Call is a single request handler called via client.Call or the generated client code
func (e *Example) PostRet(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log(" POST PostRet /api/v1.0/rs !!!")
	//初始化错误码
	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	//数据库配置信息
	redis_config_map := map[string]string{
		"key": utils.G_server_name,
		"conn": utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNum": utils.G_redis_dbnum,
	}
	redis_config ,_:=json.Marshal(redis_config_map)

	//连接redis数据库 创建句柄
	bm, err := cache.NewCache("redis", string(redis_config) )
	if err != nil {
		log.Log("缓存创建失败", err)
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	//从redis中查询手机短信验证码
	value := bm.Get(req.Mobile)
	if value ==nil {
		log.Log("获取到缓存数据查询失败",value)
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	//进行解码
	var info interface{}
	json.Unmarshal(value.([]byte), &info)
	//类型转换
	vercode := int(info.(float64))
	// 获取用户输入的短信验证码
	value, err = strconv.Atoi( req.Smscode)
	// 验证用户输入的短信验证码
	if vercode != value {
		log.Log("短信验证码错误")
		rsp.Errno = utils.RECODE_INCOREECT
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 构建user实例
	user := models.User{}
	user.Name = req.Mobile //就用手机号登陆
	//密码正常情况下 md5 sha256 sm9 存入数据库的是你加密后的编码不是明文存入
	//user.Password_hash = GetMd5String(req.Password)
	user.Password_hash = req.Password
	user.Mobile = req.Mobile

	// 创建orm句柄
	o:=orm.NewOrm()
	//插入数据库
	id,err := o.Insert(&user)
	if err != nil {
		rsp.Errno = utils.RECODE_USER_EXISTS
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	log.Log("添加用户成功，用户ID为：",id)

	//生成sessionID
	sessionId := GetMd5String(req.Mobile + req.Password)

	//返回给客户端session
	rsp.SessionID = sessionId

	//拼接key sessionid + name
	bm.Put(sessionId + "name",string(user.Mobile),time.Second * 3600)
	//拼接key sessionid + user_id
	bm.Put(sessionId + "user_id", string(user.Id) ,time.Second * 3600)
	//拼接key sessionid + mobile
	bm.Put(sessionId + "mobile",string(user.Mobile) ,time.Second * 3600)
	return nil
}

// md5加密
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

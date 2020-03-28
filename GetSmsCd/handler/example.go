package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"github.com/micro/go-log"
	"math/rand"
	example "sss/GetSmsCd/proto/example"
	"sss/GetSmsCd/sms"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"strconv"
	"time"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) GetSmsCd(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log(" GET smscd api/v1.0/smscode/:mobile ")
	//初始化返回正确的返回值
	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	//验证手机号
	o := orm.NewOrm()
	user := models.User{ Mobile: req.Mobile }
	err := o.Read(&user)
	if err == nil{
		log.Log("手机号码" + req.Mobile + "已经被使用!")
		rsp.Errno = utils.MOBILE_EXISTS
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	//连接redis数据库
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
	//查询相关数据
	value := bm.Get(req.Id)
	if value == nil{
		log.Log("从redis读取验证码失败", value)
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	// 将value转换成string类型
	value_str ,_ := redis.String(value, nil)
	// 判断用户输入验证码是否正确
	if req.Text != value_str{
		log.Log("图片验证码错误 ")
		rsp.Errno = utils.RECODE_INCOREECT
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	vercode := r.Intn(9999) + 1001
	log.Log("生成的手机验证码：", vercode)

	// 发送短信验证码
	msg := sms.NewMSG2()
	sendContent := "您的手机验证码为：" + strconv.Itoa(vercode)
	data := []string{sendContent, "1"}
	log.Log("准备发送短信验证码...")
	_, err = msg.SendMsg(data, user.Mobile)
	if err != nil {
		log.Log("短信验证码发送失败：", err)
		rsp.Errno = utils.RECODE_SMSERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	log.Log("验证码发送成功！")

	// 将验证码存入redis中,key为手机号码，value为验证码，有效时间为1分钟
	err = bm.Put(req.Mobile, vercode, time.Second * 60)
	if err != nil{
		log.Log("缓存出现问题")
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
	}
	return nil
}

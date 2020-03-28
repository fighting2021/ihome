package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/go-log/log"
	"path"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	example "sss/PostAvatar/proto/example"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) PostAvatar(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("PostAvatar /api/v1.0/user/avatar")

	//初始化返回正确的返回值
	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)
	
	// 获取文件的扩展名
	fileext := path.Ext(req.Filename)
	//group1 group1/M00/00/00/wKgLg1t08pmANXH1AAaInSze-cQ589.jpg
	// 存储文件到fastdfs，得到“组名”和“文件ID”
	Group, FileId ,err := models.UploadByBuffer(req.Avatar, fileext[1:])
	if err != nil {
		log.Log("上传文件失败：" ,err)
		rsp.Errno = utils.RECODE_IOERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	log.Log("上传文件的组名：", Group)
	log.Log("上传文件的ID：", FileId)

	// 从redis中读取登录用户的ID
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
	// 准备查询的key值
	sessioniduserid := req.SessionID + "user_id"

	// 从redis中查询用户ID
	_id := bm.Get(sessioniduserid)
	id := int(_id.([]uint8)[0])

	// 创建表对象
	user := models.User{Id: id, Avatar_url: FileId}

	// 将上传图片的url存储到用户表中
	o := orm.NewOrm()
	_ ,err = o.Update(&user ,"avatar_url")
	if err !=nil {
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
	}

	// 回传文件IDweb服务
	rsp.AvatarUrl = FileId
	return nil
}


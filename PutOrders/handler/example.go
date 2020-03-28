package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/go-log/log"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"strconv"

	example "sss/PutOrders/proto/example"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) PutOrders(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("PutOrders /api/v1.0/orders")

	//创建返回空间
	rsp.Errno  =  utils.RECODE_OK
	rsp.Errmsg  = utils.RecodeText(rsp.Errno)

	// redis数据库参数
	redis_config_map := map[string]string{
		"key":utils.G_server_name,
		"conn":utils.G_redis_addr+":"+utils.G_redis_port,
		"dbNum":utils.G_redis_dbnum,
	}
	redis_config ,_:=json.Marshal(redis_config_map)
	bm, err := cache.NewCache("redis", string(redis_config) )
	if err != nil {
		log.Log("缓存创建失败",err)
		rsp.Errno  =  utils.RECODE_DBERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return  nil
	}

	// 准备查询的key值
	sessioniduserid :=  req.Sessionid + "user_id"

	// 获取登录用户ID
	value_id := bm.Get(sessioniduserid)
	user_id :=  int(value_id.([]uint8)[0])

	// 得到当前订单id
	order_id,_ := strconv.Atoi(req.Orderid)
	// 得到执行指令
	action := req.Action

	// 订单对象
	order := models.OrderHouse{}

	// 根据ID查找状态为WAIT_ACCEPT的订单信息
	o := orm.NewOrm()
	err  = o.QueryTable("order_house").Filter("id", order_id).Filter("status", models.ORDER_STATUS_WAIT_ACCEPT).One(&order)
	if err != nil {
		log.Log("根据ID查询订单失败")
		rsp.Errno  =  utils.RECODE_DATAERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 加载订单对应的房屋信息
	if _, err := o.LoadRelated(&order, "House"); err != nil {
		rsp.Errno  =  utils.RECODE_DATAERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}
	house := order.House

	// 如果当前登录用户不是房屋的屋主，无法继续操作
	if house.User.Id != user_id {
		rsp.Errno  =  utils.RECODE_DATAERR
		rsp.Errmsg  = "订单用户不匹配,操作无效"
		return nil
	}

	// 修改订单状态
	if action == "accept" {
		// 待评价状态
		order.Status = models.ORDER_STATUS_WAIT_COMMENT
	} else if action == "reject" {
		// 拒绝状态
		order.Status = models.ORDER_STATUS_REJECTED
		// 添加评论
		order.Comment = req.Reason
	}

	// 更新订单
	if _, err := o.Update(&order); err != nil {
		rsp.Errno = utils.RECODE_DATAERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
	}
	return nil
}

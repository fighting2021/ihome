package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"strconv"

	"github.com/micro/go-log"

	example "sss/PutComment/proto/example"
	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) PutComment(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("PutComment /api/v1.0/user/orders")

	rsp.Errno  =  utils.RECODE_OK
	rsp.Errmsg  = utils.RecodeText(rsp.Errno)

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

	sessioniduserid :=  req.Sessionid + "user_id"

	// 从获取中得到登录用户ID
	value_id :=bm.Get(sessioniduserid)
	user_id :=  int(value_id.([]uint8)[0])
	//得到订单id
	order_id ,_ := strconv.Atoi(req.Orderid)
	//获得评价内容
	comment := req.Comment

	//检验评价信息是否合法 确保不为空
	if comment == "" {
		rsp.Errno  =  utils.RECODE_PARAMERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 根据ID查询状态为待评价的订单
	order := models.OrderHouse{}
	o := orm.NewOrm()
	if err := o.QueryTable("order_house").Filter("id", order_id).Filter("status", models.ORDER_STATUS_WAIT_COMMENT).One(&order); err != nil {
		rsp.Errno  =  utils.RECODE_DATAERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	//关联查询order订单所关联的user信息
	if _, err := o.LoadRelated(&order, "User"); err != nil {
		rsp.Errno  =  utils.RECODE_DATAERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 如果登录用户不是该订单的创建人，无法继续操作
	if user_id != order.User.Id {
		rsp.Errno  =  utils.RECODE_DATAERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	//关联查询order订单所关联的House信息
	if _, err := o.LoadRelated(&order, "House"); err != nil {
		rsp.Errno  =  utils.RECODE_DATAERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	//将房屋订单成交量+1
	house := order.House
	house.Order_count++

	// 修改订单状态为COMPLETE
	order.Status = models.ORDER_STATUS_COMPLETE
	order.Comment = comment

	//将order和house更新数据库
	if _, err := o.Update(&order, "status", "comment"); err != nil {
		log.Log("update order status, comment error, err = ", err)
		rsp.Errno  =  utils.RECODE_DATAERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	if _, err := o.Update(house, "order_count"); err != nil {
		log.Log("update house order_count error, err = ", err)
		rsp.Errno  =  utils.RECODE_DATAERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 将house_info_[house_id]的缓存key删除
	house_info_key := "house_info_" + strconv.Itoa(house.Id)
	if err := bm.Delete(house_info_key); err != nil {
		log.Log("delete ", house_info_key, "error , err = ", err)
	}
	return nil
}


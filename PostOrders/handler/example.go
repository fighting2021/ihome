package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/go-log/log"
	"reflect"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"strconv"
	"time"

	example "sss/PostOrders/proto/example"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) PostOrders(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("Postorders api/v1.0/orders")

	rsp.Errno  =  utils.RECODE_OK
	rsp.Errmsg  = utils.RecodeText(rsp.Errno)

	// redis数据库参数
	redis_config_map := map[string]string{
		"key":utils.G_server_name,
		"conn":utils.G_redis_addr+":"+utils.G_redis_port,
		"dbNum":utils.G_redis_dbnum,
	}
	redis_config ,_:=json.Marshal(redis_config_map)
	// 连接redis
	bm, err := cache.NewCache("redis", string(redis_config) )
	if err != nil {
		log.Log("缓存创建失败",err)
		rsp.Errno  =  utils.RECODE_DBERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return  nil
	}

	// 获取登录用户ID
	sessioniduserid :=  req.Sessionid + "user_id"
	value_id :=bm.Get(sessioniduserid)
	log.Log(value_id,reflect.TypeOf(value_id))
	userid := int(value_id.([]uint8)[0])

	// 将请求参数封装成map
	var RequestMap = make(map[string]interface{})
	err  =json.Unmarshal(req.Body, &RequestMap)
	if err != nil {
		rsp.Errno  =  utils.RECODE_REQERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 判断请求参数是否为空
	if RequestMap["house_id:"]== "" || RequestMap["start_date"] == "" || RequestMap["end_date"] == "" {
		rsp.Errno  =  utils.RECODE_REQERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 将日期格式字符串转换成Time类型
	start_date_time, _ := time.Parse("2006-01-02 15:04:05",RequestMap["start_date"].(string)+" 00:00:00")
	end_date_time, _ := time.Parse("2006-01-02 15:04:05", RequestMap["end_date"].(string)+" 00:00:00")

	// 计算入住天数
	days := end_date_time.Sub(start_date_time).Hours() / 24 + 1

	// 根据house_id得到关联的房源信息
	house_id, _ := strconv.Atoi(RequestMap["house_id"].(string))
	//房屋对象
	house := models.House{Id: house_id}
	o := orm.NewOrm()
	if err := o.Read(&house); err != nil {
		rsp.Errno  =  utils.RECODE_NODATA
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}
	// 加载User
	o.LoadRelated(&house, "User")

	// 判断当前登录用户是否是预定房屋的屋主，如果是则无法将继续预定
	if userid == house.User.Id {
		rsp.Errno  =  utils.RECODE_ROLEERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 确保用户选择的房屋未被预定,日期没有冲突
	if end_date_time.Before(start_date_time) {
		rsp.Errno  =  utils.RECODE_ROLEERR
		rsp.Errmsg  = "结束时间在开始时间之前"
		return nil
	}

	// TODO 添加征信步骤

	order := models.OrderHouse{}
	order.House = &house
	user := models.User{Id: userid}
	order.User = &user
	order.Begin_date = start_date_time
	order.End_date = end_date_time
	order.Days = int(days) // 入住天数
	order.House_price = house.Price // 房间价格
	// 封装order订单
	amount := days * float64(house.Price) // 订单金额
	order.Amount = int(amount)
	order.Status = models.ORDER_STATUS_WAIT_ACCEPT // 订单状态，最开始为“未接单状态”
	order.Credit = false // 征信情况，true代表征信良好

	// 将订单信息入库表中
	if _, err := o.Insert(&order); err != nil {
		rsp.Errno  =  utils.RECODE_DBERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 延长用户登录状态2小时
	bm.Put(sessioniduserid, string(userid) ,time.Second * 7200)

	// 返回order_id
	rsp.OrderId = int64(order.Id)
	return nil
}

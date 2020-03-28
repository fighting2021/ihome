package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/go-log/log"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"

	example "sss/GetUserOrder/proto/example"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) GetUserOrder(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("PostOrders /api/v1.0/orders")

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
	userid :=  int(value_id.([]uint8)[0])

	//得到用户角色
	log.Log(req.Role)

	o := orm.NewOrm()
	orders := []models.OrderHouse{}
	order_list := []interface{}{} //存放订单的切片

	// 房东：landlord，住客：custom
	if "landlord" == req.Role {
		//现找到自己目前已经发布了哪些房子
		landLordHouses := []models.House{}
		o.QueryTable("house").Filter("user__id", userid).All(&landLordHouses)

		housesIds := []int{}
		for _, house := range landLordHouses {
			housesIds = append(housesIds, house.Id)
		}
		//在从订单中找到房屋id为自己房源的id
		o.QueryTable("order_house").Filter("house__id__in", housesIds).OrderBy("ctime").All(&orders)
	} else {
		//角色为租客
		_,err := o.QueryTable("order_house").Filter("user__id", userid).OrderBy("ctime").All(&orders)
		if err != nil {
			log.Log(err)
		}
	}

	//循环将数据放到切片中
	for _, order := range orders {
		o.LoadRelated(&order, "User")
		o.LoadRelated(&order, "House")
		order_list = append(order_list,order.To_order_info())
	}

	// 转换json格式后发送给web服务端
	rsp.Orders, _ = json.Marshal(order_list)
	return nil

}


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

	example "sss/PostHouse/proto/example"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) PostHouse(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("PostHouses 发布房源信息 /api/v1.0/houses ")

	rsp.Errno  =  utils.RECODE_OK
	rsp.Errmsg  = utils.RecodeText(rsp.Errno)

	// 定义一个map，用于接收页面参数
	var Requestmap = make(map[string]interface{})
	// 解析json
	json.Unmarshal(req.Max, &Requestmap)
	// 打印
	for key, value := range Requestmap {
		log.Log(key, value)
	}
	// 定义House表对象
	house :=models.House{}
	// 准备数据
	house.Title = Requestmap["title"].(string)
	price , _ := strconv.Atoi(Requestmap["price"].(string))
	house.Price = price * 100
	house.Address = Requestmap["address"].(string)
	house.Room_count,_ = strconv.Atoi(Requestmap["room_count"].(string))
	house.Acreage,_ =  strconv.Atoi(Requestmap["acreage"].(string))
	house.Unit = Requestmap["unit"].(string)
	house.Capacity,_ = strconv.Atoi(Requestmap["capacity"].(string))
	house.Beds = Requestmap["beds"].(string)
	deposit,_ := strconv.Atoi(Requestmap["deposit"].(string))
	house.Deposit = deposit * 100
	house.Min_days,_ = strconv.Atoi(Requestmap["min_days"].(string))
	house.Max_days,_ = strconv.Atoi(Requestmap["max_days"].(string))
	//	"facility":["1","2","3","7","12","14","16","17","18","21","22"]
	facility := []*models.Facility{}
	for _, f_id := range Requestmap["facility"].([]interface{}) {
		fid,_ :=strconv.Atoi(f_id.(string))
		fac := &models.Facility{Id:fid}
		facility = append(facility,fac)
	}

	area_id ,_ := strconv.Atoi(Requestmap["area_id"].(string))
	area := models.Area{Id: area_id}
	house.Area= &area
	// redis数据库配置参数
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
	// 准备查询的Key值
	sessioniduserid :=  req.Sessionid + "user_id"

	// 从redis查询登录用户
	value_id := bm.Get(sessioniduserid)
	userid :=  int(value_id.([]uint8)[0])

	user := models.User{Id: userid}
	house.User = &user

	// 插入数据，并得到新增房屋的ID
	orm := orm.NewOrm()
	_ , err =orm.Insert(&house)
	if err !=nil{
		rsp.Errno  =  utils.RECODE_DBERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 新增房屋设施数据（facility_houses）
	num, err := orm.QueryM2M(&house,"Facilities").Add(facility)
	if err != nil{
		rsp.Errno  =  utils.RECODE_DBERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}
	if num == 0 {
		rsp.Errno  =  utils.RECODE_NODATA
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}
	rsp.House_Id = int64(house.Id)
	log.Log("发布房源成功！新增房屋ID为：", rsp.House_Id)
	return nil
}

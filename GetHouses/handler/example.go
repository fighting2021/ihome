package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/orm"
	"github.com/go-log/log"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"strconv"

	example "sss/GetHouses/proto/example"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) GetHouses(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("GetHourses api/v1.0/houses")

	rsp.Errno  =  utils.RECODE_OK
	rsp.Errmsg  = utils.RecodeText(rsp.Errno)

	//获取url上的参数信息
	var aid int //地区
	aid, _ = strconv.Atoi(req.Aid)
	var sd string //起始时间
	sd  = req.Sd
	var ed string //结束时间
	ed = req.Ed
	var sk string //第三栏的信息
	sk = req.Sk
	var page int // 当前第几页
	page ,_ = strconv.Atoi(req.P)
	log.Log(aid, sd, ed, sk, page)

	// 定义切片存储查询结果
	houses := []models.House{}
	//创建orm句柄
	o := orm.NewOrm()
	//设置查找的表
	qs := o.QueryTable("house")
	//查找某个地区的所有房屋
	num, err := qs.Filter("area_id", aid).All(&houses)
	if err != nil {
		rsp.Errno  =  utils.RECODE_PARAMERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}
	// 计算总页数
	total_page := int(num) / models.HOUSE_LIST_PAGE_CAPACITY + 1
	house_page := 1

	// 把结果封装到一个map类型的切片
	house_list := []interface{}{}
	for _, house := range houses {
		o.LoadRelated(&house, "Area")
		o.LoadRelated(&house, "User")
		o.LoadRelated(&house, "Images")
		o.LoadRelated(&house, "Facilities")
		house_list = append(house_list, house.To_house_info())
	}

	// 返回结果给web服务端
	rsp.TotalPage = int64(total_page)
	rsp.CurrentPage = int64(house_page)
	rsp.Houses,_ = json.Marshal(house_list)
	return nil
}

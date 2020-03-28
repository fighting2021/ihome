package handler

import (
	"context"
	"encoding/json"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/orm"
	"github.com/go-log/log"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"time"

	example "sss/GetIndex/proto/example"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

const (
	house_page_key = "home_page_data"
)

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) GetIndex(ctx context.Context, req *example.Request, rsp *example.Response) error {
	//创建返回空间
	rsp.Errno  =  utils.RECODE_OK
	rsp.Errmsg  = utils.RecodeText(rsp.Errno)

	data := []interface{}{}
	//1 从缓存服务器中请求 "home_page_data" 字段,如果有值就直接返回
	//先从缓存中获取房屋数据,将缓存数据返回前端即可
	redis_config_map := map[string]string{
		"key":   utils.G_server_name,
		"conn":  utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNum": utils.G_redis_dbnum,
	}

	redis_config, _ := json.Marshal(redis_config_map)
	cache_conn, err := cache.NewCache("redis", string(redis_config))
	if err != nil {
		log.Log("访问redis失败：", err)
	} else {
		house_page_value := cache_conn.Get(house_page_key)
		if house_page_value != nil {
			log.Log("======= get house page info  from CACHE!!! ========")
			//直接将二进制发送给客户端
			rsp.Max = house_page_value.([]byte)
			return nil
		}
	}

	houses := []models.House{}

	// 从数据库中查询到房屋列表
	o := orm.NewOrm()
	if _, err := o.QueryTable("house").Limit(models.HOME_PAGE_MAX_HOUSES).All(&houses); err == nil {
		// 加载房屋关联数据
		for _, house := range houses {
			o.LoadRelated(&house, "Area")
			o.LoadRelated(&house, "User")
			o.LoadRelated(&house, "Images")
			o.LoadRelated(&house, "Facilities")
			data = append(data, house.To_house_info())
		}
	}

	// 将数据存入缓存数据
	temp, _ := json.Marshal(data)
	cache_conn.Put(house_page_key, temp, 3600 * time.Second)
	rsp.Max = temp
	return nil
}

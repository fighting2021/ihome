package handler

import (
	"context"
	"encoding/json"
	"github.com/afocus/captcha"
	"github.com/astaxie/beego/cache"
	"github.com/micro/go-log"
	"image/color"
	example "sss/GetImageCd/proto/example"
	"sss/IhomeWeb/utils"
	"time"

	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) GetImageCd(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("----------- GET /api/v1.0/imagecode/:uuid GetImage() --------")

	//创建1个句柄
	cap := captcha.New()

	//通过句柄调用 字体文件
	if err := cap.SetFont("comic.ttf"); err != nil {
		log.Log("找不到comic.ttf字体文件！")
		panic(err.Error())
	}

	//设置图片的大小
	cap.SetSize(91, 41)
	// 设置干扰强度
	cap.SetDisturbance(captcha.MEDIUM)
	// 设置前景色 可以多个 随机替换文字颜色 默认黑色
	//SetFrontColor(colors ...color.Color) 这两个颜色设置的函数属于不定参函数
	cap.SetFrontColor(color.RGBA{255, 255, 255, 255})
	// 设置背景色 可以多个 随机替换背景色 默认白色
	cap.SetBkgColor(color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255},
	color.RGBA{0, 153, 0, 255})
	//生成图片 返回图片和 字符串(图片内容的文本形式)
	img, str := cap.Create(4, captcha.NUM)
	log.Log("生成验证码：", str)

	b := *img //解引用
	c := *(b.RGBA) //解引用

	// 成功返回
	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)

	// 图片信息
	rsp.Pix = []byte(c.Pix)
	rsp.Stride = int64(c.Stride)
	rsp.Max = &example.Response_Point{ X: int64(c.Rect.Max.X), Y: int64(c.Rect.Max.Y) }
	rsp.Min = &example.Response_Point{ X: int64(c.Rect.Min.X), Y: int64(c.Rect.Min.Y) }

	// redis配置信息
	redis_config_map := map[string]string{
		"key": "ihome",
		"conn": utils.G_redis_addr + ":" + utils.G_redis_port,
		"dbNum": utils.G_redis_dbnum,
	}
	redis_config, _ := json.Marshal(redis_config_map)

	// 连接redis
	bm, err := cache.NewCache("redis", string(redis_config))
	if err != nil{
		log.Fatal("缓存创建失败", err)
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 验证码进行1个小时缓存
	bm.Put(req.Uuid, str, 3600 * time.Second)

	return nil
}

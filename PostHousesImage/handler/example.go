package handler

import (
	"context"
	"github.com/astaxie/beego/orm"
	"github.com/go-log/log"
	"path"
	"sss/IhomeWeb/models"
	"sss/IhomeWeb/utils"
	"strconv"

	example "sss/PostHousesImage/proto/example"
)

type Example struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) PostHousesImage(ctx context.Context, req *example.Request, rsp *example.Response) error {
	//打印被调用的函数
	log.Log("PostHousesImage  /api/v1.0/houses/:id/images")

	//初始化返回正确的返回值
	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)


	// 获取文件的后缀名
	fileext := path.Ext(req.Filename)

	// 将图片上传到fastdfs服务器上
	_, RemoteFileId, err := models.UploadByBuffer(req.Image, fileext[1:])
	if err !=nil{
		log.Log("Postupavatar  models.UploadByBuffer err" ,err)
		rsp.Errno = utils.RECODE_IOERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}

	// 从请求url中得到我们的house_id
	houseid, _ := strconv.Atoi(req.Id)

	// 创建house 对象
	house := models.House{Id:houseid}

	// 根据ID查询房屋信息
	o := orm.NewOrm()
	err = o.Read(&house)
	if err !=nil{
		rsp.Errno  =  utils.RECODE_DBERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return  nil
	}

	/*判断index_image_url 是否为空 */
	if house.Index_image_url == ""{
		/*空就把这张图片设置为主图片*/
		house.Index_image_url = RemoteFileId
	}

	// 将房屋图片添加到house_image表中
	img := models.HouseImage{House:&house, Url:RemoteFileId}
	house.Images = append(house.Images, &img)

	//将图片对象插入表单之中
	_,err  =o.Insert(&img)
	if  err !=nil{
		rsp.Errno  =  utils.RECODE_DBERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	//更新house表
	_ , err =o.Update(&house)
	if err !=nil{
		rsp.Errno  =  utils.RECODE_DBERR
		rsp.Errmsg  = utils.RecodeText(rsp.Errno)
		return nil
	}

	/*返回正确的数据回显给前端*/
	rsp.Url=RemoteFileId
	return nil
}


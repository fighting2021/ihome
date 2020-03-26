package sms

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const BaseUrl = "https://app.cloopen.com:8883"

// base url

/*
URL 格式/2013-12-26/Accounts/{accountSid}/SMS/TemplateSMS?sig={SigParameter}
参数1：accountSid = "8a216da860bad76d0160bfdd6a640279"
参数2：SigParameter  md5 加密下面的3个数据
		（accountid   控制台提供
			令牌  提供
			时间戳（当前时间） 需要格式化---20140416142030）
*/


// 请求头格式要求
/*
Accept:application/json;

Content-Type:application/json;charset=utf-8;

Content-Length:;   // 要自己计算

Authorization:	计算base64（id:time）  要求时间戳与URL 中时间戳是一个
*/

/*
{"to":"13911281234,15010151234,13811431234","appId":
 "ff8080813fc70a7b013fc72312324213","templateId":"1","datas":["替换内容","替换内容"]}
*/

/*
*/

// 把发送消息 封装成一个对象
type RLYMSG struct {
	// 属性 需要哪些属性，分析
	accountSid string	// 账户ID
	authToken string	// 账户令牌
	appId string		// 应用ID
	templateId string	// 模板ID
	flag bool  // 用于校验sdk使用这是否是通过NewMSG 创建短信对象
}

func NewMSG2() *RLYMSG {
	accountSid := "8a216da867e886be0167f82da186003f"
	authToken := "e07ebcb4aea345a5bb4802d16eddc82f"
	appId := "8a216da867e886be0167f82da1a60040"
	templateId  := "1"

	// 创建对象
	msgInstance:= new(RLYMSG)
	// 对对象属性赋值
	msgInstance.appId = appId
	msgInstance.accountSid = accountSid
	msgInstance.templateId = templateId
	msgInstance.authToken = authToken
	msgInstance.flag = true
	// 返回创建好的对象
	return msgInstance
}

// RLYMSG{}
func NewMSG(accountSid string,authToken string,appId string, templateId string) *RLYMSG {
	// 创建对象
	msgInstance:= new(RLYMSG)
	// 对对象属性赋值
	msgInstance.appId = appId
	msgInstance.accountSid = accountSid
	msgInstance.templateId = templateId
	msgInstance.authToken = authToken
	msgInstance.flag = true
	// 返回创建好的对象
	return msgInstance
}

// 生成签名信息 以及 验证信息
func (this *RLYMSG)genSigAndAuth() (sig string, auth string) {
	// 计算时间戳   20140416142030 ,按照自己的需求格式化
	curTime:=time.Now().Format("20060102150405")
	fmt.Println(curTime)
	// 生成签名参数
	sigstr:=this.accountSid+this.authToken+curTime  // 要加密的字符串
	// md5 加密
	sigB:=md5.Sum([]byte(sigstr))  // 返回是一个字节数组

	// 把字节数组转化成字符串，而且字符转大写
	sig=strings.ToUpper(fmt.Sprintf("%x",sigB))  // fmt.Sprintf("%X",sigB)  完成签名的计算

	// 生成编码对象，用于base64编码生成验证信息
	enc:=base64.StdEncoding

	// 要编码的字符串
	encstr:=this.accountSid+":"+curTime

	// 编码数据
	auth = enc.EncodeToString([]byte(encstr))

	return
}


// 构造短信模板内容
func (this *RLYMSG)genBodyContent(data[]string, mobile ...string) (body string,err error) {

	// 正常需要跟进我们的模板id进行不同的处理，这里按照模板1进行处理

	//{"to":"13911281234,15010151234,13811431234","appId":
	// "ff8080813fc70a7b013fc72312324213","templateId":"1","datas":["替换内容","替换内容"]}

	// 格式化要发送的短信内容 传参就是一个切片 不需要再做处理
	// 校验内容
	if len(data)==0{
		return "",fmt.Errorf("发送的消息为空!")
	}

	// 校验手机号码。。。。 可以使用正则进行手机号的的校验
	if len(mobile) == 0{
		return "",fmt.Errorf("没有手机号!")
	}

	// 根据不同的模板 进行不同的格式化
	switch this.templateId {
	case "1":

		// 返回的内容map
		content:=map[string]interface{}{
			"to":strings.Join(mobile,","),
			"appId":this.appId,
			"templateId":this.templateId,
			"datas":data,
		}

		// 把content 序列化转成json格式
		rsp,err:=json.Marshal(content)
		if err != nil{
			return "",err
		}
		body = string(rsp)
		return body, nil
	case "2":
		// 模板2的内容格式化 等等
		;
	default:
		;

	}
	return
}

// 实现发送消息的方法
func (this *RLYMSG)SendMsg(data []string, mobile ...string) (rspcode string, err error)  {
	// 校验标志 如果不是通过new创建的对象 则 不运行调用
	if this.flag == false{
		return "",fmt.Errorf("请使用NewMSG方法创建对象")
	}

	// 创建http的一个客户端
	client:=http.Client{}  // 创建完成http客户端

	// 构造http请求
	// 构造URL  格式/2013-12-26/Accounts/{accountSid}/SMS/TemplateSMS?sig={SigParameter}
	sig,auth:=this.genSigAndAuth()
	reqUrl:=BaseUrl+ "/2013-12-26/Accounts/"+this.accountSid+"/SMS/TemplateSMS?sig="+sig

	// 创建http 的 Reqeust对象
	// 根据传递进来的参数，生成request body内容
	content,err:=this.genBodyContent(data, mobile...)
	if err!=nil{
		return "11111", err
	}

	// 创建body   要求body对象，可以转化成io.Reader 接口
	var body bytes.Buffer  // buffer 实现了io.ReadWriter 接口
	body.Write([]byte(content))  // 把content 写入body，正常还需要错误判断

	// 创建http请求
	// 参数1 指定请求方式  参数2： 要访问的url 参数3： 请求体内容，要求对象实现了io.Reader 接口
	req,err:=http.NewRequest("POST",reqUrl,&body)
	if err!=nil{
		return "11111",err
	}

	// 为Request，添加请求头  // 初始化的时候没有添加请求头的参数所以需要通过方法添加
	//Accept:application/json;
	//Content-Type:application/json;charset=utf-8;
	//	Content-Length:;   // 要自己计算
	//Authorization:	计算base64（id:time）  要求时间戳与URL 中时间戳是一个
	req.Header.Add("Accept","application/json")  // 指定客户端接收json格式数据
	req.Header.Add("Content-Type","application/json;charset=utf-8;")  // 请求携带的数据格式是json
	contentLg:=len(content)	// 计算请求体的长度
	req.Header.Add("Content-Length",strconv.Itoa(contentLg))	 // 指定请求体数据长度，对方需要校验
	req.Header.Add("Authorization",auth)		// 校验信息

	// 请求Request构造完成 ，发送http请求
	response,err:=client.Do(req)   //do 方法传递不同的请求 则发送不同的请请求，返回Http Response 以及错误
	if err!=nil{
		return "11111",err
	}


	// 接收应答，把body内容转化成byte数据便于数据解析
	buf:=make([]byte,response.ContentLength)
	response.Body.Read(buf)

	// 用于反序列化得到应答内容，存储到go对象便于处理
	var rspContent map[string]interface{}   // 应为response中可能有前途map 所有我们的value需要是接口类型，可以存储任何数据

	//json模块把body内容反序列化到 rspContent
	json.Unmarshal(buf,&rspContent)
	fmt.Println(rspContent)

	// 获取响应码
	rspcode = rspContent["statusCode"].(string) // 断言转化状态码
	return
}

func main() {
	accountSid := "8a216da867e886be0167f82da186003f"
	authToken := "e07ebcb4aea345a5bb4802d16eddc82f"
	appId := "8a216da867e886be0167f82da1a60040"
	templateId  := "1"
	msg := NewMSG(accountSid,authToken,appId,templateId)
	data:=[]string{"888888","5"}  //您的验证码为{1}，请于{2}分钟内正确输入，如非本人操作，请忽略此短信。
	code,err:=msg.SendMsg(data,"13622298413")
	fmt.Println(code,err)
}

// go msg:=NewMSG(accountSid,authToken,appId,templateId)
// 1. web 服务端受请求以后
// 2. 解析请求 生成验证码
// 3 发送验证码
// 4 把发送的验证码保证redis，---- redis存储可以设置过期时间  5分过期 ---5分钟过期后 数据从redis 中删除
// 5 给前短信响应

// 验证码校验， 从redis 提取验证码，提取不到，说明验证码过期，提取到了比对不正确说明 验证码错误
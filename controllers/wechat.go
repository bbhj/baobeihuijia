package controllers

import (
	_ "encoding/json"
	"encoding/base64"
	_ "reflect"

	_ "fmt"
	"github.com/astaxie/beego/logs"
	"github.com/bbhj/bbac/models"
	"github.com/esap/wechat"
	"github.com/imroc/req"
	_ "strings"
	"time"

	"github.com/astaxie/beego"
)

//  /getphonenum  获取微信授权手机号

// https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=ACCESS_TOKEN
// Operations about Articles
type WechatController struct {
	beego.Controller
}

// @Title CreateArticle
// @Description create users
// @Param	body		body 	models.Article	true		"body for user content"
// @Success 200 {int} models.Article.Id
// @Failure 403 body is empty
// @router /createqrcode [post]
func (u *WechatController) CreateQRcode() {
	beego.Info("create qr code ====")
	var params models.QRCodeRequestParms
	var qrret models.QRCodeReturn

	params.AccessToken = wechat.GetAccessToken()
	params.Page = "/pages/home/main"
	params.Scene = "uuid=1111111111111111"
	apiurl := "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=" + params.AccessToken

	param := req.Param{
		// "access_token": wechat.GetAccessToken(),
		"page":  "pages/profile/main",
		"scene": "forward?user=1111",
		"width": 430,
		"auto_color": false,
		"is_hyaline": false,
	}

	// req.Debug = false
	
	// r, err := req.Post(apiurl, "application/json", req.BodyJSON(param))
	r, err := req.Post(apiurl, req.BodyJSON(param))
	if err != nil {
		beego.Error(err)
	}

	var ret models.RetMsg
	r.ToJSON(&qrret)
	if qrret.Errcode != 0 {
		beego.Error("get qr code from wechat failed, ", ret)
		ret.Status = -1
        	ret.Errcode= qrret.Errcode
		ret.Errmsg = qrret.Errmsg
		
		u.Data["json"] = ret
		u.ServeJSON()
		return
	}
	// beego.Error(r)
	// Read entire JPG into byte slice.
	// reader := bufio.NewReader(r)

	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(r.Bytes())

	// beego.Error(encoded)

	// r.ToFile("qq3.jpg")

	u.Ctx.WriteString(encoded)
	return
}

// @Title Get
// @Description get user by uid
// @Param	uid		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Article
// @Failure 403 :uid is empty
// @router /login [get]
func (u *WechatController) Login() {
	var wxlogin models.OpenWeixinAccessToken

	// uri?code=021cljRR1DUhk81klvSR1xCxRR1cljRM&state=bbhj
	var wxauth models.WechatAuth
	wxauth.Code = u.GetString("code")
	wxauth.State = u.GetString("state")
	wxauth.Scene = u.GetString("scene")
	wxauth.Path = u.GetString("path")
	wxauth.ShareTicket = u.GetString("shareTicket")

	// json.Unmarshal(u.Ctx.Input.RequestBody, &wxauth)
	logs.Info("-------------", wxauth)
	if "" == wxauth.Code {
		return
	}

	if "" != wxauth.Scene {
		// apiurl := fmt.Sprintf("%sappid=%s&secret=%s&js_code=%s&grant_type=authorization_code", beego.AppConfig.String("WechatAuthUrl"), beego.AppConfig.String("wechat_appid"), beego.AppConfig.String("wechat_secret"), wxauth.Code)
		apiurl := beego.AppConfig.String("wechat_mina") + wxauth.Code

		req.SetTimeout(50 * time.Second)
		a, _ := req.Get(apiurl)

		logs.Debug(a)
		a.ToJSON(&wxlogin)
		logs.Debug("xxxxx", wxlogin)

		return
	} else {
		apiurl := beego.AppConfig.String("weixin_oauth") + wxauth.Code

		r, _ := req.Get(apiurl)
		// logs.Info(r)
		r.ToJSON(&wxlogin)
		if wxlogin.Openid == "" {
			return
		}
		models.AddOpenWeixinAccessToken(wxlogin)
		logs.Info("weixin login: ", wxlogin)

		// get userinfo
		apiurl = "https://api.weixin.qq.com/sns/userinfo?access_token=" + wxlogin.AccessToken + "&openid=" + wxlogin.Openid
		r, _ = req.Get(apiurl)
		var webuserinfo models.User
		r.ToJSON(&webuserinfo)
		models.AddUserInfo(webuserinfo)
		logs.Info("get weixin userinfo: ", webuserinfo)

		// u.Ctx.SetCookie("access_token", wxlogin.AccessToken, 7200)
		u.Ctx.SetCookie("token", wxlogin.AccessToken, 7200)
	}
	// u.Ctx.Redirect(302, "https://wechat.baobeihuijia.com/index.html")
	u.Data["json"] = wxlogin
        u.Data["json"] = map[string]string{"status": "0"}
	u.ServeJSON()
}

// // @Title Get
// // @Description get user by uid
// // @Param	uid		path 	string	true		"The key for staticblock"
// // @Success 200 {object} models.Article
// // @Failure 403 :uid is empty
// // @router /login [post]
// func (u *WechatController) Login() {
// 	// uri?code=021cljRR1DUhk81klvSR1xCxRR1cljRM&state=bbhj
// 	code := u.GetString("code")
// 	logs.Info("=====code====", code)
// 	u.ServeJSON()
// }
// @Title Get

// @Description get user by uid
// @Param	uid		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Article
// @Failure 403 :uid is empty
// @router /userinfolist [get]
func (u *WechatController) UserInfoList() {
	// uri?code=021cljRR1DUhk81klvSR1xCxRR1cljRM&state=bbhj
	var userInfoList []models.User
	userInfoList = models.GetUserInfoList()
	u.Data["json"] = userInfoList
	u.ServeJSON()
}

// func GetMiniProgramQrcode(c *gin.Context) {
//     // 获取应用的appid
//     appid := c.PostForm("appid")
//     // 后端获取小程序传来的page及scene
//     page := c.PostForm("page")
//     scene := c.PostForm("scene")
//
//     // 考虑到反复请求微信接口的耗时及服务器io消耗，我打算把图片通过MD5(page+scene)的方式命名
//     h := md5.New()
//     h.Write([]byte(page + scene)) // 需要加密的字符串
//     cipherStr := h.Sum(nil)
//     result := hex.EncodeToString(cipherStr)
//     exist, _ := model.PathExists("服务器文件存储路径" + result + ".jpg") // 检测图片是否已经存在(即之前是否有人分享过相同页面)
//     if exist {
//         // 若二维码文件存在直接返回路径
//         c.String(200, "URL访问路径"+result+".jpg")
//     } else {
//         // 不存在则直接请求微信获取二维码
//         token, ok := GetAccessToken(appid) // 首先获取access_token 这里大家根据自己的业务方式来获取
//         if !ok {
//             c.JSON(200, gin.H{
//                 "code": 4001,
//                 "msg":  "accesstoken 获取失败",
//             })
//         } else {
//             // 向微信请求小程序二维码图片
//             // 这里需要注意！！！ 官方只介绍了通过该接口以post的形式传参，但其实参数是要严格的json格式传递才能正常获取
//             resp, err := http.Post("https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token="+token,
//                 "application/x-www-form-urlencoded",
//                 strings.NewReader(`{"page":"`+page+`", "scene":"`+scene+`"}`))
//             if err != nil {
//                 fmt.Println(err)
//             }
//
//             defer resp.Body.Close()
//             body, err := ioutil.ReadAll(resp.Body)
//             if err != nil {
//                 // handle error
//             }
//             // 图片写入本地
//             err = ioutil.WriteFile("服务器文件存储路径"+result+".jpg", body, 0755)
//             if err != nil {
//                 fmt.Println(err)
//                 c.JSON(200, gin.H{
//                     "code": 4002,
//                     "msg":  "文件写入服务器失败",
//                 })
//             }else{
//                 // 写入成功 直接返回url
//                 c.String(200, "URL访问路径"+result+".jpg")
//             }
//         }
//     }
// }

// Bilibili 扫码登录相关功能，包括官网登录和三方校验。
//
// 开放平台文档：https://open.bilibili.com/doc/4/eaf0e2b5-bde9-b9a0-9be1-019bb455701c
package biliqr

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/skip2/go-qrcode"
)

// NewLoginQRInfo 获取全新的登录二维码的信息，其中的 LoginInfo.URL 用于生成二维码。
func NewLoginQRInfo() (*LoginQRInfo, error) {
	url := "https://passport.bilibili.com/qrcode/getLoginUrl"
	data, err := SimpleGet(url)
	if err != nil {
		return nil, err
	}
	var res struct {
		Data    LoginQRInfo `json:"data"`
		Code    int         `json:"code"`
		Message string      `json:"message"`
	}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, errors.New(res.Message)
	}
	return &res.Data, err
}

type LoginQRInfo struct {
	URL      string `json:"url"`
	OauthKey string `json:"oauthKey"`
}

// GetLoginQRImage 获取等待扫描的登录二维码。
// 可以使用 ToSmallString 方法生成可输出终端的二维码。
//
// level - qrcode.Low - 7% - int(0)
//
// level - qrcode.Medium - 15% - int(1)
//
// level - qrcode.High - 25% - int(2)
//
// level - qrcode.Highest - 30% - int(3)
func NewLoginQR(level qrcode.RecoveryLevel) (*qrcode.QRCode, *LoginQRInfo, error) {
	info, err := NewLoginQRInfo()
	if err != nil {
		return nil, nil, err
	}
	qr, err := qrcode.New(info.URL, level)
	if err != nil {
		return nil, nil, err
	}
	return qr, info, nil
}

// GetQRStatus 获取二维码状态，返回用于登录官网的 SESSDATA。
// 轮询调用本方法可获取实时状态。状态分为未扫码(-3)、扫码未确认(-5)、扫码已确认(0)、二维码失效(-2)。
// 扫码确认后返回一个 URL，请求该 URL 后 Set-Cookie 包含 SESSDATA，SESSDATA 用于官网登录。
func GetQRStatus(oauthKey string) (*QRStatus, error) {
	var res struct {
		Data    QRStatus `json:"data"`
		Code    int      `json:"code"`
		Message string   `json:"message"`
	}
	data, _, cookies, err := SimpleRequest("GET", "https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key="+oauthKey, nil, nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	sessdata, err := GetCookieValue(cookies, "SESSDATA")
	if err == nil {
		res.Data.SESSDATA = sessdata
	}
	if res.Code != 0 {
		return nil, errors.New(res.Message)
	}
	return &res.Data, nil
}

// GetCookieValue 获取指定 Name 的 Cookie 值
func GetCookieValue(cookies []*http.Cookie, name string) (string, error) {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value, nil
		}
	}
	return "", errors.New("cookie with name " + name + " not found")
}

// QRStatus 二维码状态。
// 状态分为未扫码(-3)、扫码未确认(-5)、扫码已确认(0)、二维码失效(-2)。
// 扫码确认后，返回 TmpToken，用于获取用户登录后的数据。
type QRStatus struct {
	Url          string `json:"string"`
	RefreshToken string `json:"refresh_token"`
	Code         int    `json:"code"`
	Message      string `json:"message"`
	SESSDATA     string
}

// GetThirdQRStatus 获取二维码状态，返回 TmpToken。
//
// 轮询调用本方法可获取实时状态。状态分为未扫码(-3)、扫码未确认(-5)、扫码已确认(0)、二维码失效(-2)。
//
// oauthKey 由 [NewLoginQR] 或 [NewLoginQRInfo] 返回。
func GetThirdQRStatus(oauthKey string) (*ThirdQRStatus, error) {
	body := url.Values{}
	body.Set("oauthKey", oauthKey)
	body.Set("source", "oauth2")
	var res ThirdQRStatus
	data, err := SimplePost("https://passport.bilibili.com/qrcode/authorize/poll", &body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	codeTips := map[int]string{
		-3: "未扫码",
		-5: "扫码未确认",
		0:  "扫码已确认",
		-2: "二维码失效",
	}
	if res.Code == -2 {
		return nil, errors.New(codeTips[res.Code])
	}
	res.Message = codeTips[res.Code]
	return &res, nil
}

// QRStatus 二维码状态。
// 状态分为未扫码(-3)、扫码未确认(-5)、扫码已确认(0)、二维码失效(-2)。
// 扫码确认后，返回 TmpToken，用于获取用户登录后的数据。
type ThirdQRStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
	Data    struct {
		TmpToken string `json:"tmp_token"`
	} `json:"data"`
}

func (t ThirdQRStatus) Success() bool {
	return t.Code == 0
}

// GetAuthorizeCode 获取 Bilibili 重定向到三方地址时携带的 Code。
//
// clientId 哔哩哔哩开放平台申请应用时分配。
//
// tmpToken 扫码确认后由 [GetThirdQRStatus] 返回。
//
// returnURL 应用授权回调地址。该参数为创建应用时填写的「授权回调域」。
//
// 参数的详细说明，请参考 https://open.bilibili.com/doc/4/aac73b2e-4ff2-b75c-4c96-35ced865797b
func GetAuthorizeCode(clientId string, tmpToken string, returnURL string) (codeInfo *AuthorizeCodeInfo, err error) {
	body := url.Values{}
	body.Set("client_id", clientId)
	body.Set("tmp_token", tmpToken)
	body.Set("scopes", "NFT_BASE,LIVER_BASE,FANS_BASE,USER_INFO")
	body.Set("state", "1")
	body.Set("return_url", returnURL)
	data, err := SimplePost("https://api.bilibili.com/x/account-oauth2/v1/authorize", &body)
	if err != nil {
		return nil, err
	}
	var info AuthorizeCodeInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}
	if info.Code != 0 {
		return nil, errors.New(info.Message)
	}
	return &info, nil
}

type AuthorizeCodeInfo struct {
	// 非 0 时表示操作失败
	Code int `json:"code"`
	// 操作失败时的提示文本
	Message string `json:"message"`
	Data    struct {
		// Bilibili 重定向到三方地址时携带的 Code。
		Code        string `json:"code"`
		RedirectUrl string `json:"redirect_url"`
	} `json:"data"`
}

// Dodo 文件上传模块。
//
// 文件上传步骤：
//
//	// 获取 Token 和 UID。
//	token, uid, _ := GetTokenAndUID(tmpToken)
//	// 创建文件上传任务。
//	work, _ := NewUploadWork("path/abc.mp3", token, uid)
//	// 获取当前文件 MD5 的历史上传记录。
//	history, _ := work.History()
//	// 如果文件已经被上传过了，直接拿到下载地址。
//	if history.HasRecord {
//		fmt.Println(history.ResourceUrl)
//	} else {
//		// 上传文件。
//		work.Upload()
//		// 上报文件信息。
//		url, _ := work.Record()
//		fmt.Println(url)
//	}
package dodo

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"strings"
	"time"

	"github.com/iuroc/gododo/biliqr"
)

// GetTokenAndUID 获取 Token 和 UID。
//
// tmpToken 在扫码确认后由 [biliqr.GetThirdQRStatus] 返回。
func GetTokenAndUID(tmpToken string) (token string, uid string, err error) {
	// clientId 和 returnURL 可以在三方网站跳转到 B 站授权页面时携带的 GET 参数获得。
	clientId := "0c95e37758534eb7"
	returnURL := "https://www.imdodo.com/thirdLogin/biliLogin"
	codeInfo, err := biliqr.GetAuthorizeCode(clientId, tmpToken, returnURL)
	if err != nil {
		return "", "", err
	}
	apiKey, sha1Key := RandKeyConfig()
	body := url.Values{
		"code":   {codeInfo.Data.Code},
		"apikey": {apiKey},
	}
	sig := HmacSha1Encrypt([]byte(sha1Key), []byte(body.Encode()))
	body.Set("sig", sig)
	data, err := biliqr.SimplePost("https://apis.imdodo.com/web/login/fetch-bilibili-user-info", &body)
	if err != nil {
		return "", "", err
	}
	var info struct {
		Data struct {
			Token string `json:"token"`
			User  struct {
				UID int `json:"uid"`
			} `json:"user"`
		} `json:"data"`
		Message string `json:"message"`
		Status  int    `json:"status"`
	}
	err = json.Unmarshal(data, &info)
	if err != nil {
		return "", "", err
	}
	if info.Status != 0 {
		return "", "", errors.New(info.Message)
	}
	return info.Data.Token, strconv.Itoa(info.Data.User.UID), nil
}

// RandKeyConfig 获取解密参数，其中 ApiKey 作为 POST 请求的参数，sha1Key 用于加密生成 sig。
func RandKeyConfig() (apiKey string, sha1Key string) {
	configs := [][]string{
		{"CK18tnKeKDN", "t8yqYCqv68rKOwgPRUBv4Z2hS4kKajHc0yYzrXLf"},
		{"CGrmRus4Xl4", "BrxswEvSCZK0fTvN5rGyQNqqZAL7vjzZHjDfOXXZ"},
		{"9mEnDRJrkl6", "0ZFDcgZX9iigWbbzmHmqcMFFpZFZcrOu91TsRVCU"},
	}
	rand.NewSource(time.Now().UnixNano())
	index := rand.Intn(len(configs))
	return configs[index][0], configs[index][1]
}

// HmacSha1Encrypt 加密
func HmacSha1Encrypt(key []byte, data []byte) string {
	h := hmac.New(sha1.New, key)
	h.Write(data)
	signature := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
}

type UploadConfig struct {
	OSSAccessKeyId string `json:"OSSAccessKeyId"`
	Policy         string `json:"policy"`
	Signature      string `json:"signature"`
	Dir            string `json:"dir"`
	Host           string `json:"host"`
	Expire         int    `json:"expire"`
}

type UploadConfigMap map[string]string

// NewUploadWork 新的上传任务。
//
// Path: 需要上传的文件的路径。
//
// token 和 uid: [GetTokenAndUID] 获取得到。
func NewUploadWork(path string, token string, uid string) (*UploadWork, error) {
	work := UploadWork{
		Path:  path,
		Token: token,
		UID:   uid,
		Ext:   filepath.Ext(path),
		Base:  filepath.Base(path),
	}
	stat, err := os.Stat(work.Path)
	if err != nil {
		return nil, err
	}
	md5, err := GetFileMD5(path)
	if err != nil {
		return nil, err
	}
	work.MD5 = md5
	work.Stat = stat
	return &work, nil
}

// UploadWork 文件上传任务。
type UploadWork struct {
	Path string
	// [GetTokenAndUID] 获取得到。
	Token string
	// [GetTokenAndUID] 获取得到。
	UID  string
	Stat fs.FileInfo
	Base string
	Ext  string
	MD5  string
}

func (w UploadWork) Record() (string, error) {
	apiKey, sha1Key := RandKeyConfig()
	resourceUrl := "https://files.imdodo.com/dodo/" + w.MD5 + w.Ext
	params, body := ParseParamArray([][2]string{
		{"MD5Str", w.MD5},
		{"apikey", apiKey},
		{"clientType", "3"},
		{"clientVersion", "0.14.2"},
		{"fileName", w.Base},
		{"fileSize", strconv.FormatInt(w.Stat.Size(), 10)},
		{"resourceType", "5"},
		{"resourceUrl", resourceUrl},
		{"timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10)},
		{"token", w.Token},
		{"uid", w.UID},
	})
	sig := HmacSha1Encrypt([]byte(sha1Key), []byte(params))
	body.Set("sig", sig)
	header := http.Header{}
	header.Set("Token", w.Token)
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	request, err := http.NewRequest("POST", "https://apis.imdodo.com/api/oss/file/record", strings.NewReader(body.Encode()))
	if err != nil {
		return "", err
	}
	request.Header.Set("Token", w.Token)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	var record struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(responseData, &record)
	if err != nil {
		return "", err
	}
	if record.Status != 0 {
		return "", errors.New(record.Message)
	}
	return resourceUrl, nil
}

func (w UploadWork) Upload() error {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	defer writer.Close()
	config, err := w.Config()
	if err != nil {
		return err
	}
	configMap := map[string]string{
		"OSSAccessKeyId": config.OSSAccessKeyId,
		"policy":         config.Policy,
		"signature":      config.Signature,
		"dir":            config.Dir,
		"host":           config.Host,
		"expire":         strconv.Itoa(config.Expire),
	}
	for key, value := range configMap {
		if err := writer.WriteField(key, value); err != nil {
			return err
		}
	}
	err = writer.WriteField("key", "dodo/"+w.MD5+w.Ext)
	if err != nil {
		return err
	}
	part, err := writer.CreateFormFile("file", w.Base)
	if err != nil {
		return err
	}
	file, err := os.Open(w.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	writer.Close()
	request, err := http.NewRequest("POST", config.Host, &b)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if strings.Contains(string(responseBody), "<Error>") {
		match := regexp.MustCompile(`<Message>(.*?)</Message>`)
		result := match.FindStringSubmatch(string(responseBody))
		errorMessage := ""
		if len(result) == 1 {
			errorMessage = result[1]
		}
		return errors.New(errorMessage)
	}
	return nil
}

// History 获取当前文件的 MD5 的历史上传记录，如果存在历史记录，则直接可获得文件下载地址。
func (w UploadWork) History() (*UploadHistory, error) {
	apiKey, sha1Key := RandKeyConfig()
	body := url.Values{
		"MD5Str":        {w.MD5},
		"apikey":        {apiKey},
		"clientType":    {"3"},
		"clientVersion": {"0.14.2"},
		"timestamp":     {strconv.FormatInt(time.Now().Unix(), 10)}, // 当前时间戳
		"token":         {w.Token},
		"uid":           {w.UID},
	}
	sig := HmacSha1Encrypt([]byte(sha1Key), []byte(body.Encode()))
	body.Set("sig", sig)
	data, err := biliqr.SimplePost("https://apis.imdodo.com/api/oss/file/history", &body)
	if err != nil {
		return nil, err
	}
	var history struct {
		Data    UploadHistory `json:"data"`
		Message string        `json:"message"`
		Status  int           `json:"status"`
	}
	err = json.Unmarshal(data, &history)
	if err != nil {
		return nil, err
	}
	if history.Status != 0 {
		return nil, errors.New(history.Message)
	}
	return &history.Data, nil
}

type UploadHistory struct {
	HasRecord   bool   `json:"hasRecord"`
	ResourceUrl string `json:"resourceUrl"`
}

func (w UploadWork) Config() (*UploadConfig, error) {
	body := url.Values{
		"bucket": {"oss-dodo-upload"},
		"dir":    {"dodo/"},
		"uid":    {w.UID},
	}
	data, err := biliqr.SimplePost("https://apis.imdodo.com/api/oss/fetchUploadSign", &body)
	if err != nil {
		return nil, err
	}
	var config struct {
		Data    UploadConfig `json:"data"`
		Message string       `json:"message"`
		Status  int          `json:"status"`
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	if config.Status != 0 {
		return nil, errors.New(config.Message)
	}
	return &config.Data, nil
}

// ParseParamArray 将二维数组拼接为 Params 字符串。
func ParseParamArray(array [][2]string) (params string, outBody *url.Values) {
	var str string
	body := url.Values{}
	for index, item := range array {
		str += item[0] + "=" + item[1]
		if index != len(array)-1 {
			str += "&"
		}
		body.Set(item[0], item[1])
	}
	return str, &body
}

// GetFileMD5 获取文件的 MD5 Hex。
func GetFileMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	if _, err = io.Copy(hash, file); err != nil {
		return "", err
	}
	md5Hash := hash.Sum(nil)
	return hex.EncodeToString(md5Hash), nil
}

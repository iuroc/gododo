package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/iuroc/gododo/biliqr"
	"github.com/iuroc/gododo/dodo"
	"github.com/skip2/go-qrcode"
)

func main() {
	fmt.Println("DoDo 文件直链获取工具 [github.com/iuroc/gododo]")
	scanner := bufio.NewScanner(os.Stdin)
	userInfo := GetUserInfo()
	for {
		fmt.Printf("\n%s\n\n", strings.Repeat("-", 40))
		fmt.Print("🚩 输入文件路径或拖拽文件到此处: ")
		if !scanner.Scan() {
			fmt.Println("❗ ", scanner.Err())
			continue
		}
		path := TrimPathInput(scanner.Text())
		work, err := dodo.NewUploadWork(path, userInfo.Token, userInfo.UID)
		if err != nil {
			fmt.Println("❗ 错误:", err)
			continue
		}
		history, err := work.History()
		if err != nil {
			fmt.Println("❗ 错误:", err)
			continue
		}
		if history.HasRecord {
			fmt.Println("🎉 上传成功:", history.ResourceURL)
			continue
		}
		if err = work.Upload(); err != nil {
			fmt.Println("❗ 错误:", err)
			continue
		}
		resourceURL, err := work.Record()
		if err != nil {
			fmt.Println("❗ 错误:", err)
			continue
		}
		fmt.Println("🎉 上传成功:", resourceURL)
	}
}

// TrimPathInput 去除路径两端的特殊字符。
func TrimPathInput(input string) string {
	return regexp.MustCompile(`^[\s&'"]+|[\s&'"]+$`).ReplaceAllString(input, "")
}

// GetUserInfo 获取用户信息，如果不存在，则要求用户扫码登录。
func GetUserInfo() *UserInfo {
	data, err := os.ReadFile("userInfo.json")
	needQR := false
	if os.IsNotExist(err) {
		needQR = true
	} else if err != nil {
		log.Fatalln("[os.ReadFile] userInfo 文件读取失败", err)
	}
	var userInfo UserInfo
	err = json.Unmarshal(data, &userInfo)
	needQR = err != nil || !userInfo.Check()
	if needQR {
		qr, info, err := biliqr.NewLoginQR(qrcode.Low)
		if err != nil {
			log.Fatalln("[biliqr.NewLoginQR] 创建二维码失败", err)
		}
		fmt.Println(qr.ToSmallString(false))
		for {
			status, err := biliqr.GetThirdQRStatus(info.OauthKey)
			if err != nil {
				log.Fatalln("[biliqr.GetThirdQRStatus]", err)
			}
			if status.Success() {
				token, uid, err := dodo.GetTokenAndUID(status.Data.TmpToken)
				if err != nil {
					log.Fatalln("dodo.GetTokenAndUID", err)
				}
				userInfo.Token = token
				userInfo.UID = uid
				userInfo.Save("userInfo.json")
				return &userInfo
			}
		}
	}
	return &userInfo
}

type UserInfo struct {
	Token string `json:"token"`
	UID   string `json:"uid"`
}

// Check 校验 Token 和 UID 的有效性。
func (info UserInfo) Check() bool {
	return dodo.CheckTokenAndUID(info.Token, info.UID)
}

// Save 以 JSON 格式保存 Token 和 UID 到文件。
func (info UserInfo) Save(path string) {
	data, err := json.Marshal(info)
	if err != nil {
		log.Fatalln("[json.Marshal]", err)
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		log.Fatalln("[os.WriteFile]", err)
	}
}

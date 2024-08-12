package dodo_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/iuroc/gododo/biliqr"
	"github.com/iuroc/gododo/dodo"
	"github.com/skip2/go-qrcode"
)

func TestGetTokenAndUID(t *testing.T) {
	_, _, err := dodo.GetTokenAndUID("1234")
	if err == nil {
		t.Fatal("未检查出错误的 tmpToken")
	}

	qr, info, err := biliqr.NewLoginQR(qrcode.Low)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(qr.ToSmallString(false))
	for {
		status, err := biliqr.GetThirdQRStatus(info.OauthKey)
		if err != nil {
			if err.Error() != "二维码失效" {
				t.Fatal(err)
			} else {
				t.Error(err)
				break
			}
		} else if status.Success() {
			_, _, err := dodo.GetTokenAndUID(status.Data.TmpToken)
			if err != nil {
				t.Fatal(err)
			}
			break
		}
		t.Log(status.Code, status.Message)
		time.Sleep(time.Second)
	}
}

func TestUpload(t *testing.T) {
	path := "test-file-123"
	err := os.WriteFile(path, []byte(strconv.FormatInt(time.Now().UnixNano(), 10)), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	qr, info, err := biliqr.NewLoginQR(qrcode.Low)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(qr.ToSmallString(false))
	num := 0
	for {
		status, err := biliqr.GetThirdQRStatus(info.OauthKey)
		if err != nil {
			if err.Error() != "二维码失效" {
				t.Fatal(err)
			} else {
				t.Error(err)
				break
			}
		} else if status.Success() {
			token, uid, err := dodo.GetTokenAndUID(status.Data.TmpToken)
			if err != nil {
				t.Fatal(err)
			}
			work, err := dodo.NewUploadWork(path, token, uid)
			if err != nil {
				t.Fatal(err)
			}
			history, err := work.History()
			if err != nil {
				t.Fatal(err)
			}
			if history.HasRecord {
				fmt.Println("读取历史记录成功 👉", history.ResourceUrl)
			}
			if err = work.Upload(); err != nil {
				t.Fatal(err)
			}
			url, err := work.Record()
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("文件上传成功 👉", url)
			break
		} else if status.Code == -5 {
			num++
			if num == 30 {
				t.Fatal("用户扫码后长时间未确认")
			}
		}
		time.Sleep(time.Second)
	}
}

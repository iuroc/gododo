package biliqr_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/iuroc/gododo/biliqr"
	"github.com/skip2/go-qrcode"
)

func TestLoginQRInfo(t *testing.T) {
	info, err := biliqr.NewLoginQRInfo()
	if err != err {
		t.Fatal(err)
	}
	t.Log(info)
}

func TestGetQRStatus(t *testing.T) {
	qr, info, err := biliqr.NewLoginQR(qrcode.Low)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(qr.ToSmallString(false))
	i := 0
	for {
		i++
		if i == 20 {
			t.Fatal("timeout")
		}
		status, err := biliqr.GetQRStatus(info.OauthKey)
		if err != nil {
			t.Fatal(err)
		}
		if status.Code == 0 {
			t.Log(status.SESSDATA)
			break
		}
		time.Sleep(time.Second)
	}
}

func TestThirdGetQRStatus(t *testing.T) {
	qr, info, err := biliqr.NewLoginQR(qrcode.Low)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(qr.ToSmallString(false))
	var tmpToken string
	clientId := "0c95e37758534eb7"
	returnURL := "https://www.imdodo.com/thirdLogin/biliLogin"
	i := 0
	for {
		i++
		if i == 20 {
			t.Fatal("timeout")
		}
		status, err := biliqr.GetThirdQRStatus(info.OauthKey)
		if err != nil {
			t.Fatal(err)
		}
		if status.Code == 0 {
			tmpToken = status.Data.TmpToken
			break
		}
		time.Sleep(time.Second)
	}
	codeInfo, err := biliqr.GetAuthorizeCode(clientId, tmpToken, returnURL)
	if err != nil {
		t.Fatal(err)
	}
	if codeInfo.Code != 0 {
		t.Fatal(codeInfo.Message)
	}
	t.Log("Code:", codeInfo.Data.Code)
	t.Log("RedirectUrl:", codeInfo.Data.RedirectUrl)
}

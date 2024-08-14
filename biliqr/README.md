
## Bilibili 扫码授权模块 `/biliqr`

```go
// 生成用于登录的二维码
qr, info, _ := biliqr.NewLoginQR(qrcode.Low)
// 将二维码输出到终端
fmt.Println(qr.ToSmallString(false))
for {
    // 获取二维码目前的状态（第三方，返回 TmpToken）
    statusThird, _ := biliqr.GetThirdQRStatus(info.OauthKey)

    // 获取二维码目前的状态（官网，返回 SESSDATA）
    status, _ := biliqr.GetQRStatus(info.OauthKey)

    // 判断已经扫码并确认，获得 TmpToken。
    if status.Success() {
        // 根据 TmpToken 获取 Bilibili 服务器返回的 Code。
        codeInfo, _ := biliqr.GetAuthorizeCode("clientId", statusThird.Data.TmpToken, "returnURL")

        // 三方校验
        fmt.Println("Code:", codeInfo.Data.Code)
        fmt.Println("TmpToken:", statusThird.Data.TmpToken)

        // 官网登录
        fmt.Println("SESSDATA:", status.SESSDATA)

        break
    }
    time.Sleep(time.Second)
}
```

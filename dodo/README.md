## DoDo 文件上传模块 `/dodo`

```go
// 生成用于登录的二维码
qr, info, _ := biliqr.NewLoginQR(qrcode.Low)
// 将二维码输出到终端
fmt.Println(qr.ToSmallString(false))
for {
    // 获取二维码目前的状态（第三方）
    status, err := biliqr.GetThirdQRStatus(info.OauthKey)
    // 判断已经扫码并确认，获得 TmpToken。
    if status.Success() {。
        // 通过 TmpToken 获取 Token 和 UID 用于 Dodo 文件上传。
        token, uid, _ := dodo.GetTokenAndUID(status.Data.TmpToken)
        // 创建文件上传任务
        work, err := dodo.NewUploadWork("/path/book.pdf", token, uid)
        // 通过该文件的 MD5 判断文件是否有历史上传记录
        history, err := work.History()
        if history.HasRecord {
            fmt.Println("读取历史记录成功，文件下载地址 👉", history.ResourceUrl)
        } else {
            // 上传文件
            work.Upload()
            // 提交文件上传记录，使文件直链生效。
            resourceUrl, _ := work.Record()
            fmt.Println("文件上传成功，文件下载地址 👉", url)
        }
        break
    }
    time.Sleep(time.Second)
}
```

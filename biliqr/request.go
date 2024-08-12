package biliqr

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

// SimpleGet 简单的发送 GET 请求，返回响应内容。
func SimpleGet(url string) ([]byte, error) {
	body, _, _, err := SimpleRequest("GET", url, nil, nil)
	return body, err
}

// SimpleGet 简单的发送 POST 请求，返回响应内容。
func SimplePost(url string, body *url.Values) ([]byte, error) {
	header := http.Header{}
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseBody, _, _, err := SimpleRequest("POST", url, strings.NewReader(body.Encode()), header)
	return responseBody, err
}

// SimpleRequest 简单的发送 HTTP 请求，返回响应内容。
func SimpleRequest(method string, _url string, body io.Reader, header http.Header) (
	responseBody []byte,
	responseHeader http.Header,
	cookies []*http.Cookie,
	err error,
) {
	request, err := http.NewRequest(method, _url, body)
	if err != nil {
		return nil, nil, nil, err
	}
	request.Header = header
	if err != nil {
		return nil, nil, nil, err
	}
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, nil, err
	}
	defer response.Body.Close()
	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, nil, err
	}
	return resBody, response.Header, response.Cookies(), nil
}

package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestParsePathInput(t *testing.T) {
	str := `&& '1234''`
	match := regexp.MustCompile(`^[\s&'"]+|[\s&'"]+$`)
	fmt.Println(match.ReplaceAllString(str, ""))
}

func TestEncrypt(t *testing.T) {
	config, err := NewAESConfig()
	if err != nil {
		t.Fatal(err)
	}
	result := config.Encrypt([]byte("Hello World"))
	fmt.Println(result)
	fmt.Println(config.Decrypt(result))
}

func TestGetUserInfo(t *testing.T) {
	fmt.Printf("%#v", GetUserInfo())
}

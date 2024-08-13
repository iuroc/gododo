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

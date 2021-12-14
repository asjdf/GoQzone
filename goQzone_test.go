package goQzone

import (
	"fmt"
	"io/ioutil"
	"testing"
)

// 发送说说（文字+图片）
func TestSendEmotion(t *testing.T) {
	client := Init()
	if err := client.QrLogin(); err != nil {
		panic(err)
	}

	f, err := ioutil.ReadFile("./pic/logo.png")
	if err != nil {
		panic(err)
	}

	err = client.NewPost().Content("github.com/asjdf/GoQzone Test").Pic(f).Send()
	fmt.Println(err)
}

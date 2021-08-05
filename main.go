package main

import (
	"GoQzone/pkg/goQzone"
	"fmt"
	"io/ioutil"
)

func main() {
	client := goQzone.Init()
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
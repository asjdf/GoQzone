package main

import (
	"GoQzone-Demo/pkg/goQzone"
	"fmt"
	"math/rand"
	"strconv"
)

func main() {
	fmt.Println("hello,world!")
	fmt.Println(strconv.FormatFloat(rand.Float64(),'f',-1,64))
	goQzone.Init()
}
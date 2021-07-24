package goQzone

import (
	"github.com/parnurzeal/gorequest"
	_ "image/png"
)

var s *service

type service struct {
	Request *gorequest.SuperAgent

}

//var request *gorequest.SuperAgent

func Init() {
	s = new(service)
	s.Request = gorequest.New()
	s.Request.DoNotClearSuperAgent = true
	s.Request.Header.Add("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36")
	s.Request.Proxy("http://127.0.0.1:8866")
	//loginSig,err := getLoginSig()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(loginSig)
	//_, err = quickLoginCheck(loginSig)
	//if err != nil {
	//	panic(err)
	//}
	//_ = quickLoginPtqrshow()

	//fmt.Println(getAction())

	getPtqrtoken("vce3hXUN9bwffYVFSI3bUZk7IPFaDrCaPFI4AN-bVhEhEn7IBqRPLeb-tlyQItKt")
	if err := qrLogin();err != nil{
		panic(err)
	}

}
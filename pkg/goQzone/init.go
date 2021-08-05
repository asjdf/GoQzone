package goQzone

import (
	"github.com/parnurzeal/gorequest"
	_ "image/png"
)

var s *Service

type Service struct {
	Request *gorequest.SuperAgent
}

func Init() *Service {
	s = new(Service)
	s.Request = gorequest.New()
	s.Request.DoNotClearSuperAgent = true
	s.Request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36")
	//s.Request.Proxy("http://127.0.0.1:8866")
	go s.keepCookieAlive()
	return s
}

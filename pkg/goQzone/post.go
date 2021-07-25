package goQzone

import (
	"fmt"
	"net/url"
	"strconv"
)

func (s *service)getGtk() string {
	h := int64(5381)
	skey := ""
	url,_:=url.Parse("https://user.qzone.qq.com/")
	for _,v := range s.Request.Client.Jar.Cookies(url){
		if v.Name == "skey"{
			skey = v.Value
		}
	}
	for _,v := range skey{
		h += (h << 5) + int64(v)
	}
	fmt.Println(strconv.FormatInt(h&0x7fffffff, 10))
	return strconv.FormatInt(h&0x7fffffff, 10)
}

//func (s *service)post()  {
//
//}
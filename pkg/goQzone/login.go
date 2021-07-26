package goQzone

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func convert2Ascii(img image.Image) string {
	table := []string{" ", "░", "▒", "▓", "█"}
	temp := ""

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y += 3 {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			level := c.Y / 51 // 51 * 5 = 255
			if level == 5 {
				level--
			}
			temp += table[level]
		}
		temp += "\n"
	}
	return temp
}

func hash33(t string) int64 {
	var e int64
	e = 0
	for _, v := range t {
		e += (e << 5) + int64(v)
	}
	return 2147483647 & e
}

func getPtqrtoken(qrsig string) string {
	return strconv.FormatInt(hash33(qrsig), 10)
}

func getAction() string {
	return "0-0-" + strconv.FormatInt(time.Now().Unix()*1000, 10)
}

func (s *service)getQrCode() (qrSig string, img image.Image, err error) {
	s.Request.QueryData = url.Values{}
	resp, body, errs := s.Request.Get("https://ssl.ptlogin2.qq.com/ptqrshow").
		Query(struct {
			Appid      string
			E          string
			L          string
			S          string
			D          string
			V          string
			T          string
			Daid       string
			Pt_3rd_aid string
		}{
			Appid:      "549000912",
			E:          "2",
			L:          "M",
			S:          "3",
			D:          "72",
			V:          "4",
			Daid:       "5",
			Pt_3rd_aid: "0",
			T:          strconv.FormatFloat(rand.Float64(), 'f', -1, 64),
		}).
		EndBytes()
	if errs != nil {
		return "", nil, errors.New("qrLogin() 获取二维码登录二维码出错")
	}
	imgDecode, _, err := image.Decode(bytes.NewReader(body))
	if resp.Header.Get("set-cookie") == "" {
		return "", imgDecode, errors.New("qrLogin() 无set-cookie")
	}
	r, _ := regexp.Compile("qrsig=(.*);Path=/;")
	qrSig = ""
	for _, v := range resp.Header.Values("set-cookie") {
		if r.MatchString(v) {
			qrSig = r.FindStringSubmatch(v)[1]
		}
	}
	if qrSig == "" {
		return "", imgDecode, errors.New("qrLogin() 没找到qrsig")
	}
	return qrSig, imgDecode, nil
}

func (s *service)qrLogin() error {
	loginSig, err := s.getLoginSig()
	if err != nil {
		panic(err)
	}

	qrSig, qrImg, err := s.getQrCode()
	if err != nil {
		panic(err)
	}

	fmt.Println(convert2Ascii(qrImg))

	loginUrl := ""
	for {
		output, err := s.qrLoginStateCheck(getPtqrtoken(qrSig), loginSig)
		if err != nil {
			panic(err)
		}
		fmt.Printf("\r%s %v", time.Now().Format("2006-01-02 15:04:05"), output[4])
		if output[0] == "0" {
			loginUrl = output[2]
			break
		}
		time.Sleep(2500 * time.Millisecond)
	}
	fmt.Println(loginUrl)

	_, _, errs := s.Request.Get(loginUrl).End()
	if errs != nil {
		return errors.New("qrlogin() 刷新cookie时出现网络错误")
	}

	return nil
}

func (s *service)qrLoginStateCheck(ptqrtoken string, loginSig string) (output []string, err error) {
	s.Request.QueryData = url.Values{}
	u := "https://ssl.ptlogin2.qq.com/ptqrlogin?u1=https://qzs.qzone.qq.com/qzone/v5/loginsucc.html?para=izone&ptqrtoken=" + ptqrtoken + "&ptredirect=0&h=1&t=1&g=1&from_ui=1&ptlang=2052&action=" + getAction() + "&js_ver=19112817&js_type=1&login_sig=" + loginSig + "&pt_uistyle=40&aid=549000912&daid=5&"
	_, body, errs := s.Request.Get(u).End()
	if errs != nil {
		return nil, errors.New("qrLoginStateCheck() 网络请求错误")
	}
	body = body[7 : len(body)-1]
	body = strings.ReplaceAll(body, "'", "")
	return strings.Split(body, ","), nil
}

func (s *service)getLoginSig() (string, error) {
	//request := gorequest.New()
	resp, _, errs := s.Request.Get("https://xui.ptlogin2.qq.com/cgi-bin/xlogin?proxy_url=https%3A//qzs.qq.com/qzone/v6/portal/proxy.html&daid=5&&hide_title_bar=1&low_login=0&qlogin_auto_login=1&no_verifyimg=1&link_target=blank&appid=549000912&style=22&target=self&s_url=https%3A%2F%2Fqzs.qzone.qq.com%2Fqzone%2Fv5%2Floginsucc.html%3Fpara%3Dizone&pt_qr_app=%E6%89%8B%E6%9C%BAQQ%E7%A9%BA%E9%97%B4&pt_qr_link=http%3A//z.qzone.com/download.html&self_regurl=https%3A//qzs.qq.com/qzone/v6/reg/index.html&pt_qr_help_link=http%3A//z.qzone.com/download.html&pt_no_auth=1").
		End()
	if errs != nil {
		return "", errors.New("getLoginSig() 网络请求错误")
	}
	//fmt.Println(resp.Header.Values("set-cookie"))
	if resp.Header.Get("set-cookie") == "" {
		return "", errors.New("getLoginSig() 无set-cookie")
	}
	r, _ := regexp.Compile("pt_login_sig=(.*); PATH=/;")
	for _, v := range resp.Header.Values("set-cookie") {
		if r.MatchString(v) {
			return r.FindStringSubmatch(v)[1], nil
		}
	}

	return "", errors.New("getLoginSig() 没找到pt_login_sig")
}

//快速登录流程：
//1. getLoginSig()
//2. quickLoginCheck()
//3. quickLoginPtqrshow()
func quickLoginCheck(loginSig string) (ptdrvs string, err error) {
	resp, _, errs := s.Request.Get("https://ssl.ptlogin2.qq.com/check").
		Query(struct {
			Regmaster  string
			Pt_tea     string
			Pt_vcode   string
			Uin        string
			Appid      string
			Js_ver     string
			Js_type    string
			Login_sig  string
			U1         string
			R          string
			Pt_uistyle string
		}{
			Regmaster:  "",
			Pt_tea:     "2",
			Pt_vcode:   "1",
			Uin:        "243852814",
			Appid:      "549000912",
			Js_ver:     "21072114",
			Js_type:    "1",
			Login_sig:  loginSig,
			U1:         "https://qzs.qzone.qq.com/qzone/v5/loginsucc.html?para=izone",
			R:          strconv.FormatFloat(rand.Float64(), 'f', -1, 64),
			Pt_uistyle: "40",
		}).
		End()
	if errs != nil {
		return "", errors.New("quickLoginCheck() 网络通信错误")
	}
	//fmt.Println(body)
	if resp.Header.Get("set-cookie") == "" {
		return "", errors.New("quickLoginCheck() 无set-cookie")
	}
	r, _ := regexp.Compile("ptdrvs=(.*);Path=/;")
	for _, v := range resp.Header.Values("set-cookie") {
		if r.MatchString(v) {
			return r.FindStringSubmatch(v)[1], nil
		}
	}
	return "", errors.New("quickLoginCheck() 没找到ptdrvs")
}

func quickLoginPtqrshow() error {
	s.Request.QueryData = url.Values{}
	_, _, errs := s.Request.Get("https://ssl.ptlogin2.qq.com/ptqrshow").
		Query(struct {
			Qr_push_uin string
			Type        string
			Qr_push     string
			Appid       string
			T           string
			Ptlang      string
			Daid        string
			Pt_3rd_aid  string
		}{
			Qr_push_uin: "243852814",
			Type:        "1",
			Qr_push:     "1",
			Appid:       "549000912",
			T:           strconv.FormatFloat(rand.Float64(), 'f', -1, 64),
			Ptlang:      "2052",
			Daid:        "5",
			Pt_3rd_aid:  "0"}).
		End()
	if errs != nil {
		return errors.New("quickLoginPtqrshow() 网络通信错误")
	}
	fmt.Println(s.Request.Cookies)
	for _, v := range s.Request.Cookies {
		fmt.Println(v)
	}
	return nil
}

//func quickLoginStateCheck(loginSig string, ptdrvs string) bool {
//	s.Request.QueryData = url.Values{}
//	s.Request.Get("https://ssl.ptlogin2.qq.com/ptqrlogin").
//		Query()
//}

func (s *service)CheckCookieValid() bool {
	//https://user.qzone.qq.com/10001
	//"<title>QQ空间-分享生活，留住感动</title>"
	s.Request.QueryData = url.Values{}
	resp, _, err := s.Request.
		Get(fmt.Sprintf("https://h5.qzone.qq.com/proxy/domain/base.qzone.qq.com/cgi-bin/user/cgi_userinfo_get_all?uin=%s&g_tk=%s",s.getUin(),s.getGtk())).End()
	if err != nil {
		return false
	}
	if resp.StatusCode != 403 {
		return true
	}
	return false
}

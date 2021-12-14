package goQzone

import (
	"bytes"
	"errors"
	"fmt"
	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"image/jpeg"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tuotoo/qrcode"
	"image"
	"image/color"
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

func (s *Service) getPtqrtoken() string {
	return strconv.FormatInt(hash33(s.getQrsig()), 10)
}

func (s *Service) getPtLoginSig() string {
	return s.getCookie("https://ptlogin2.qq.com/", "pt_login_sig")
}

func (s *Service) getQrsig() string {
	return s.getCookie("https://ptlogin2.qq.com/", "qrsig")
}

func (s *Service) getPtGuidSig() string {
	return s.getCookie("https://ptlogin2.qq.com/", "pt_guid_sig")
}

func (s *Service) getPtGuidToken() string {
	return strconv.FormatInt(hash33(s.getPtGuidSig()), 10)
}

func getAction() string {
	return "0-0-" + strconv.FormatInt(time.Now().Unix()*1000, 10)
}

func (s *Service) getQrCode() (img image.Image, err error) {
	s.Request.QueryData = url.Values{}
	_, body, errs := s.Request.Get("https://ssl.ptlogin2.qq.com/ptqrshow").
		Query(map[string]string{
			"appid":      "549000912",
			"e":          "2",
			"l":          "M",
			"s":          "3",
			"d":          "72",
			"v":          "4",
			"daid":       "5",
			"pt_3rd_aid": "0",
			"t":          strconv.FormatFloat(rand.Float64(), 'f', -1, 64),
		}).
		EndBytes()
	if errs != nil {
		return nil, errors.New("getQrCode() 获取二维码登录二维码出错")
	}
	imgDecode, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {

	}
	return imgDecode, nil
}

// QrLogin 扫码登录
func (s *Service) QrLogin() error {
	err := s.getXlogin()
	if err != nil {
		panic(err)
	}

	qrImg, err := s.getQrCode()
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, qrImg, nil)
	if err != nil {
		return err
	}
	fi, err := qrcode.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	if f, err := os.Create("qrcode.jpeg"); err == nil {
		defer func() { _ = os.Remove("qrcode.jpeg") }()
		if err := jpeg.Encode(f, qrImg, nil); err != nil {
			return err
		}
		fmt.Println("您可以使用QQ扫描控制台中的二维码或是扫描程序运行目录中的qrcode.jpeg")
	}
	qrcodeTerminal.New2(
		qrcodeTerminal.ConsoleColors.BrightBlack,
		qrcodeTerminal.ConsoleColors.BrightWhite,
		qrcodeTerminal.QRCodeRecoveryLevels.Low).
		Get(fi.Content).Print()
	//fmt.Println(convert2Ascii(qrImg))

	loginUrl := ""
	for {
		output, err := s.qrLoginStateCheck()
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
		return errors.New("QrLogin() 刷新cookie时出现网络错误")
	}

	return nil
}

func (s *Service) qrLoginStateCheck() (output []string, err error) {
	ptqrtoken := s.getPtqrtoken()
	loginSig := s.getPtLoginSig()
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

func (s *Service) getXlogin() error {
	_, _, errs := s.Request.Get("https://xui.ptlogin2.qq.com/cgi-bin/xlogin?proxy_url=https%3A//qzs.qq.com/qzone/v6/portal/proxy.html&daid=5&&hide_title_bar=1&low_login=0&qlogin_auto_login=1&no_verifyimg=1&link_target=blank&appid=549000912&style=22&target=self&s_url=https%3A%2F%2Fqzs.qzone.qq.com%2Fqzone%2Fv5%2Floginsucc.html%3Fpara%3Dizone&pt_qr_app=%E6%89%8B%E6%9C%BAQQ%E7%A9%BA%E9%97%B4&pt_qr_link=http%3A//z.qzone.com/download.html&self_regurl=https%3A//qzs.qq.com/qzone/v6/reg/index.html&pt_qr_help_link=http%3A//z.qzone.com/download.html&pt_no_auth=1").
		End()
	if errs != nil {
		return errors.New("getXlogin() 网络请求错误")
	}
	return nil
}

//快速登录流程：
//1. fetchOnekeyListByGUID()
//2. getXlogin()
//3. quickLoginCheck()
//4. quickLoginPtqrshow()

func (s *Service) QuickLogin() error {
	_ = s.fetchOnekeyListByGUID()
	_ = s.getXlogin()
	_, _ = s.quickLoginCheck()
	//s.quickLoginPtqrshow()
	//for {
	//
	//}
	return nil
}

func (s *Service) fetchOnekeyListByGUID() error {
	s.Request.QueryData = url.Values{}
	_, _, errs := s.Request.Get("https://ssl.ptlogin2.qq.com/pt_fetch_dev_uin").
		Query(map[string]interface{}{
			"r":             strconv.FormatFloat(rand.Float64(), 'f', -1, 64),
			"pt_guid_token": s.getPtGuidToken(),
		}).End()
	if errs != nil {
		return errors.New("fetchOnekeyListByGUID() 网络请求错误")
	}
	return nil
}

func (s *Service) quickLoginCheck() (ptdrvs string, err error) {
	s.Request.QueryData = url.Values{}
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
			Login_sig:  s.getPtLoginSig(),
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

func (s *Service) quickLoginPtqrshow() error {
	s.Request.QueryData = url.Values{}
	_, _, errs := s.Request.Get("https://ssl.ptlogin2.qq.com/ptqrshow").
		Query(map[string]interface{}{
			"qr_push_uin": "243852814",
			"type":        "1",
			"qr_push":     "1",
			"appid":       "549000912",
			"t":           strconv.FormatFloat(rand.Float64(), 'f', -1, 64) + "2",
			"ptlang":      "2052",
			"daid":        "5",
			"pt_3rd_aid":  "0",
		}).
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

func (s *Service) quickLoginStateCheck(uin string, loginSig string, ptdrvs string) bool {
	s.Request.QueryData = url.Values{}
	s.Request.Get("https://ssl.ptlogin2.qq.com/ptqrlogin").
		Query(map[string]interface{}{
			"u1":         "https://qzs.qzone.qq.com/qzone/v5/loginsucc.html?para=izone&specifyurl=http%3A%2F%2Fuser.qzone.qq.com%2F" + uin,
			"ptqrtoken":  "2075302471",
			"ptredirect": "0",
			"h":          "1",
			"t":          "1",
			"g":          "1",
			"from_ui":    "1",
			"ptlang":     "2052",
			"action":     "1-1-1627316805329",
			"js_ver":     "21072114",
			"js_type":    "1",
			"login_sig":  "HgIgOy0SzMe9K4xO87ehX*-MfeZK5iEgV9yJyHtRlniJfw91q9vrFZl0BCNMpkh3",
			"pt_uistyle": "40",
			"aid":        "549000912",
			"daid":       "5",
			"ptdrvs":     "bTGGjNrzJCurCjRcwMeU4mcJE5lzQPX2m1EwcBwy129droVbdDHSzsW-NxVv2ohx",
			"sid":        "3222316841997099117",
			"has_onekey": "1",
		})
	return false
}

func (s *Service) CheckCookieValid() bool {
	s.Request.QueryData = url.Values{}
	resp, _, err := s.Request.
		Get(fmt.Sprintf("https://h5.qzone.qq.com/proxy/domain/base.qzone.qq.com/cgi-bin/user/cgi_userinfo_get_all?uin=%s&g_tk=%s", s.getUin(), s.getGtk())).End()
	if err != nil {
		return false
	}
	if resp.StatusCode != 403 {
		return true
	}
	return false
}

// 定期访问保证Cookie有效性
func (s *Service) keepCookieAlive() {
	for {
		s.Request.Get("https://user.qzone.qq.com/" + s.getUin()).End()
		time.Sleep(time.Duration(rand.Intn(30)+15) * time.Second)
	}
}

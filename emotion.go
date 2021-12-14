package goQzone

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// VisibleRight 说说可见权限
type VisibleRight int

const (
	VisibleAll    VisibleRight = 1  // 所有人可见
	VisibleFriend VisibleRight = 4  // 好友可见
	VisibleSelf   VisibleRight = 64 // 仅自己可见
	//VisiblePartFriend VisibleRight = 16  // 部分好友可见 暂不资瓷
	//VisibleBanPart    VisibleRight = 128 // 仅部分好友不可见 暂不资瓷
)

type superEmotion struct {
	Emotion Emotion
	Pics    [][]byte
	Service *Service
}

type Emotion struct {
	SynTweetVersion string       `json:"syn_tweet_verson"`
	ParamStr        string       `json:"paramstr"`
	PicTemplate     string       `json:"pic_template"`
	RichType        string       `json:"richtype"`
	RichVal         string       `json:"richval"`
	SpecialUrl      string       `json:"special_url"`
	SubRichType     string       `json:"subrichtype"`
	Con             string       `json:"con"` // 文字内容
	FeedVersion     string       `json:"feedversion"`
	Ver             string       `json:"ver"`
	UgcRight        VisibleRight `json:"ugc_right"` // 可见权限
	ToSign          string       `json:"to_sign"`
	HostUin         string       `json:"hostuin"`
	CodeVersion     string       `json:"code_version"`
	Format          string       `json:"format"`
	QzReferrer      string       `json:"qzreferrer"`
	PicBo           string       `json:"pic_bo,omitempty"`
}

func (s *Service) getGtk() string {
	h := int64(5381)
	skey := s.getCookie("https://user.qzone.qq.com/", "p_skey")
	for _, v := range skey {
		h += (h << 5) + int64(v)
	}
	return strconv.FormatInt(h&0x7fffffff, 10)
}

func (s *Service) getUin() string {
	uin := s.getCookie("https://user.qzone.qq.com/", "p_uin")
	uin = regexp.MustCompile("([1-9].[0-9]*)$").FindString(uin)
	return uin
}

func (s *Service) getSkey() string {
	return s.getCookie("https://user.qzone.qq.com/", "skey")
}

func (s *Service) getPSkey() string {
	return s.getCookie("https://user.qzone.qq.com/", "p_skey")
}

func (s *Service) getCookie(urlString string, cookieKey string) (cookieValue string) {
	urlDecode, _ := url.Parse(urlString)
	for _, v := range s.Request.Client.Jar.Cookies(urlDecode) {
		if v.Name == cookieKey {
			cookieValue = v.Value
			return
		}
	}
	return ""
}

// NewPost 新建说说
func (s *Service) NewPost() *superEmotion {
	post := new(superEmotion)
	// 初始化常量
	post.Emotion.SynTweetVersion = "1"
	post.Emotion.ParamStr = "1"
	post.Emotion.FeedVersion = "1"
	post.Emotion.Ver = "1"
	post.Emotion.ToSign = "0"
	post.Emotion.HostUin = s.getUin()
	post.Emotion.CodeVersion = "1"
	post.Emotion.Format = "fs"
	post.Emotion.QzReferrer = "https://user.qzone.qq.com/" + s.getUin()

	post.Emotion.UgcRight = 1
	post.Service = s
	return post
}

// Content 添加文字内容
func (p *superEmotion) Content(content string) *superEmotion {
	p.Emotion.Con += content
	return p
}

// Pic 添加图片
func (p *superEmotion) Pic(img []byte) *superEmotion {
	p.Pics = append(p.Pics, img)
	return p
}

// Right 设置说说权限
func (p *superEmotion) Right(right VisibleRight) *superEmotion {
	p.Emotion.UgcRight = right
	return p
}

func (p *superEmotion) Send() error {
	if p.Service.CheckCookieValid() != true {
		return errors.New("登录状态过期")
	}
	if len(p.Pics) == 0 {
		p.Service.Request.QueryData = url.Values{}
		p.Service.Request.Post("https://user.qzone.qq.com/proxy/domain/taotao.qzone.qq.com/cgi-bin/emotion_cgi_publish_v6?g_tk=" + p.Service.getGtk()).
			Type("form").
			Send(p.Emotion).
			End()
	} else {
		var picUploadResp []*picUploadRespJsonT
		for _, v := range p.Pics {
			uploadResp, _ := p.Service.uploadPic(v)
			picUploadResp = append(picUploadResp, uploadResp)
		}
		//generate richVal, bos and pic_template
		richVal := ""
		bos := ""
		picCount := 0
		for _, v := range picUploadResp {
			if v != nil {
				picCount++
				richVal += "," + v.Data.Albumid + "," + v.Data.Lloc + "," + v.Data.Sloc + "," + strconv.Itoa(v.Data.Type) + "," + strconv.Itoa(v.Data.Height) + "," + strconv.Itoa(v.Data.Width) + ",," + strconv.Itoa(v.Data.Height) + "," + strconv.Itoa(v.Data.Width) + "\t"
				bos += regexp.MustCompile("bo=(.+?)$").FindStringSubmatch(v.Data.Url)[1] + ","
			}
		}
		bos = strings.TrimSuffix(bos, ",")
		p.Emotion.RichVal = strings.TrimSuffix(richVal, "\t")
		p.Emotion.PicBo = bos + "\t" + bos
		p.Emotion.RichType = "1"
		p.Emotion.SubRichType = "1"
		if picCount != 1 {
			p.Emotion.PicTemplate = "tpl-" + strconv.Itoa(picCount) + "-1"
		}

		p.Service.Request.QueryData = url.Values{}
		p.Service.Request.Post("https://user.qzone.qq.com/proxy/domain/taotao.qzone.qq.com/cgi-bin/emotion_cgi_publish_v6?g_tk=" + p.Service.getGtk()).
			Type("form").
			Send(p.Emotion).
			End()

	}
	return nil
}

type picUploadRespJsonT struct {
	Data struct {
		Pre          string `json:"pre"`
		Url          string `json:"url"`
		Lloc         string `json:"lloc"`
		Sloc         string `json:"sloc"`
		Type         int    `json:"type"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		Albumid      string `json:"albumid"`
		Totalpic     int    `json:"totalpic"`
		Limitpic     int    `json:"limitpic"`
		OriginUrl    string `json:"origin_url"`
		OriginUuid   string `json:"origin_uuid"`
		OriginWidth  int    `json:"origin_width"`
		OriginHeight int    `json:"origin_height"`
		Contentlen   int    `json:"contentlen"`
	} `json:"data"`
	Ret int `json:"ret"` //等于-100的时候上传失败（上传失败的时候好像结构还不同，没测试）
}

func (s *Service) uploadPic(pic []byte) (*picUploadRespJsonT, error) {
	picb64 := base64.StdEncoding.EncodeToString(pic)
	_, body, errs := s.Request.Post("https://up.qzone.qq.com/cgi-bin/upload/cgi_upload_image?g_tk=" + s.getGtk() + "&&g_tk=" + s.getGtk()).
		Type("form").
		Send(map[string]interface{}{
			"filename":          "filename",
			"uin":               s.getUin(),
			"skey":              s.getSkey(),
			"zzpaneluin":        s.getUin(),
			"zzpanelkey":        "",
			"p_uin":             s.getUin(),
			"p_skey":            s.getPSkey(),
			"qzonetoken":        "",
			"uploadtype":        "1",
			"albumtype":         "7",
			"exttype":           "0",
			"refer":             "shuoshuo",
			"output_type":       "jsonhtml",
			"charset":           "utf-8",
			"output_charset":    "utf-8",
			"upload_hd":         "1",
			"hd_width":          "2048",
			"hd_height":         "10000",
			"hd_quality":        "96",
			"backUrls":          "http://upbak.photo.qzone.qq.com/cgi-bin/upload/cgi_upload_image,http://119.147.64.75/cgi-bin/upload/cgi_upload_image",
			"url":               "https://up.qzone.qq.com/cgi-bin/upload/cgi_upload_image?g_tk=" + s.getGtk(),
			"base64":            "1",
			"jsonhtml_callback": "callback",
			"picfile":           picb64,
		}).End()
	if errs != nil {
		return nil, errors.New("uploadPic() 网络出错")
	}
	respJson := new(picUploadRespJsonT)
	err := json.Unmarshal([]byte(regexp.MustCompile(`frameElement.callback\((.*)\);</script></body></html>`).FindStringSubmatch(body)[1]), &respJson)
	if err != nil {
		return nil, err
	}
	return respJson, nil
}

type emotionListT struct {
	Code    int    `json:"code"`
	Subcode int    `json:"subcode"`
	Message string `json:"message"`
	Default int    `json:"default"`
	Data    struct {
		Main struct {
			Attach               string        `json:"attach"`
			Searchtype           string        `json:"searchtype"`
			HasMoreFeeds         bool          `json:"hasMoreFeeds"`
			Daylist              string        `json:"daylist"`
			Uinlist              string        `json:"uinlist"`
			Error                string        `json:"error"`
			Hotkey               string        `json:"hotkey"`
			IcGroupData          []interface{} `json:"icGroupData"`
			HostLevel            string        `json:"host_level"`
			FriendLevel          string        `json:"friend_level"`
			Lastaccesstime       string        `json:"lastaccesstime"`
			LastAccessRelateTime string        `json:"lastAccessRelateTime"`
			Begintime            string        `json:"begintime"`
			Endtime              string        `json:"endtime"`
			Dayspac              string        `json:"dayspac"`
			HidedNameList        []interface{} `json:"hidedNameList"`
			AisortBeginTime      string        `json:"aisortBeginTime"`
			AisortEndTime        string        `json:"aisortEndTime"`
			AisortOffset         string        `json:"aisortOffset"`
			AisortNextTime       string        `json:"aisortNextTime"`
			OwnerBitmap          string        `json:"owner_bitmap"`
			Pagenum              string        `json:"pagenum"`
			Externparam          string        `json:"externparam"`
		} `json:"main"`
		Data []interface{} `json:"data"`
	} `json:"data"`
}

func (s *Service) getEmotionList(uin string) {

}

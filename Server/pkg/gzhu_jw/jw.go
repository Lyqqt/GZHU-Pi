/**
广州大学教务系统客户端接口
*/

package gzhu_jw

import (
	"GZHU-Pi/env"
	"crypto/tls"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/rpc"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type JWClient struct {
	Username  string
	Password  string
	ExpiresAt time.Time //客户端过期时间
	Client    *http.Client
}

func (c *JWClient) GetExpiresAt() time.Time {
	return c.ExpiresAt
}

func (c *JWClient) SetExpiresAt(t time.Time) {
	c.ExpiresAt = t
}

func (c *JWClient) GetUsername() string {
	return c.Username
}

func (c *JWClient) doRequest(method, url string, header http.Header, body io.Reader) (*http.Response, error) {
	t1 := time.Now()

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header[k] = v
	}
	resp, err := c.Client.Do(req)
	//defer c.Client.CloseIdleConnections()

	logs.Debug("请求耗时：", time.Since(t1), url)
	return resp, err
}

//https://my.oschina.net/u/2950272/blog/1634815
var transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

func newClient(username, password string) *JWClient {
	// Allocate a new cookie jar to mimic the browser behavior:
	cookieJar, _ := cookiejar.New(nil)
	// Fill up basic data:
	c := &JWClient{
		Username: username,
		Password: password,
	}
	//设置客户端20分钟后过期
	c.ExpiresAt = time.Now().Add(20 * time.Minute)
	// Whn initializing the http.Client, copy default values from http.DefaultClient
	// Pass a pointer to the cookie jar that was created earlier:
	c.Client = &http.Client{
		Jar:       cookieJar,
		Timeout:   time.Minute,
		Transport: transport,
	}
	return c
}

func isAvailable() error {
	h := time.Now().Hour()
	if h >= 0 && h < 7 {
		err := fmt.Errorf("当前时间段，教务系统通道关闭，服务不可用")
		logs.Warn(err)
		return err
	}
	return nil
}

func BasicAuthClient(username, password string) (client *JWClient, err error) {
	if err = isAvailable(); err != nil {
		return
	}
	if username == "" {
		return nil, fmt.Errorf("not init username or password")
	}
	c := newClient(username, password)
	//发送get请求，获取登录页面信息
	resp, err := c.doRequest("GET", Urls["jw-login"], nil, nil)
	if err != nil {
		logs.Error(err)
		return nil, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	//提取登录表单
	r1, _ := regexp.Compile(`name="lt" value="(.*?)" />`)
	lt := r1.FindStringSubmatch(string(body))

	r2, _ := regexp.Compile(`name="execution" value="(.*?)" />`)
	execution := r2.FindStringSubmatch(string(body))

	if len(lt) < 2 || len(execution) < 2 {
		return nil, fmt.Errorf("get login form failed")
	}

	captcha, err := c.GetCaptcha()
	if err != nil {
		logs.Error(err)
		return
	}

	postValue := url.Values{
		"username":  {c.Username},
		"password":  {c.Password},
		"captcha":   {captcha},
		"warn":      {"true"},
		"lt":        {lt[1]},
		"execution": {execution[1]},
		"_eventId":  {"submit"},
		"submit":    {"登录"},
	}
	//编码表单
	postString := postValue.Encode()

	resp, err = c.doRequest("POST", Urls["jw-login"], urlencodedHeader, strings.NewReader(postString))
	if err != nil {
		return nil, err
	}
	body, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	//判断是否登录成功
	if strings.Contains(string(body), "验证码不正确") {
		return nil, fmt.Errorf("验证码识别不正确，请重试")
	}
	if strings.Contains(string(body), "必须输入验证码") {
		return nil, fmt.Errorf("验证码识别失败，请重试")
	}
	if strings.Contains(string(body), "重新设置密码") {
		return nil, fmt.Errorf("温馨提示: 该账号需重置密码，且不能与初始密码相同，请登录 my.gzhu.edu.cn 操作！")
	}
	if strings.Contains(string(body), "账号或密码错误") {
		return nil, LoginError
	}
	ok, _ := regexp.MatchString(`用户名[\s\S]*密码`, string(body))
	if ok {
		return nil, fmt.Errorf("不知为啥，就是没有登录进去，请告知开发者")
	}
	return c, nil
}

var rpcClient *rpc.Client

func (c *JWClient) GetCaptcha() (capture string, err error) {

	if rpcClient == nil {
		rpcClient, err = rpc.DialHTTP("tcp", env.Conf.Rpc.Addr)
		if err != nil {
			logs.Error(err)
			return
		}
	}

	const maxTry = 3
	for range [maxTry]int{} {

		resp, err := c.doRequest("GET", "https://cas.gzhu.edu.cn/cas_server/captcha.jsp", nil, nil)
		if err != nil {
			logs.Error(err)
			return capture, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logs.Error(err)
			return capture, err
		}

		err = rpcClient.Call("OcrService.Capture", body, &capture)
		if err != nil {
			rpcClient = nil
			logs.Error(err)
			return capture, err
		}

		reg := regexp.MustCompile(`\d+`)
		res := reg.FindStringSubmatch(capture)
		if len(res) == 0 || len(res[0]) != 4 {
			logs.Debug("capture ocr failed: ", res)
			continue
		}
		capture = res[0]
		break
	}
	logs.Info("验证码：", capture)
	return
}

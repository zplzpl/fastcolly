package fastcolly

import (
	"net/http"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestFastHttp_Client1(t *testing.T){

	req,res := fasthttp.AcquireRequest(),fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI("http://shopee.com.my/api/v2/search_items/?by=sales&keyword=cat&limit=1&newest=0&order=desc&page_type=search")
	req.Header.SetUserAgent("Mozilla/5.0 (compatible; bingbot/{{.Ver}}; +http://www.bing.com/bingbot.htm{{.Coms}})")
	req.Header.Set("X-Forwarded-For","127.0.0.1")

	if err := fasthttp.Do(req,res);err!=nil {
		t.Fatal(err.Error())
	}

	if res.StatusCode() != http.StatusOK {
		t.Fatal("not status 200")
	}

	body := res.Body()

	t.Log(string(body))


}

func TestFast2(t *testing.T){

	status,buf,err := fasthttp.Get(nil,"http://www.shopee.com.my/robots.txt")
	if err!=nil {
		t.Fatal(err)
	}

	t.Log(status)
	t.Log(string(buf))
}
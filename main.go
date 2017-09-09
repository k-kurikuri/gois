package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"github.com/PuerkitoBio/goquery"
)

const REQUEST_URL = "http://whois.jprs.jp"

// コマンドライン引数から登録者名を取得
var (
	companyName = flag.String("c", "", "c option is 'CompanyName'")
)

func main() {
	fmt.Println("===== func main start!! =====")

	flag.Parse()

	// postパラメータ生成
	postParams := url.Values{}
	postParams.Add("type", "DOM-HOLDER")
	postParams.Add("key", *companyName)

	// postリクエスト
	resp, _ := http.PostForm(REQUEST_URL, postParams)
	// レスポンスをgoqueryで解析
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		fmt.Println("Oops Sorry Http Error...")
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	// preタグ内のaタグ内テキストを出力
	doc.Find("pre").Each(func(_ int, s *goquery.Selection) {
		s.Find("a").Each(func(_ int, aSec *goquery.Selection){
			fmt.Println(aSec.Text())
		})
	})
}

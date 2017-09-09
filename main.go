package main

import (
	"fmt"
	"net/http"
	"net/url"
	"database/sql"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

const COMPANY_NAME_COLUMN = "name"
const REQUEST_URL = "http://whois.jprs.jp"
const SLEEP_TIME = 10;

func main() {
	fmt.Println("===== func main start!! =====")

	//flag.Parse()

	db, err := sql.Open("mysql", "root:root@/gois")

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	rows, err := db.Query("SELECT name FROM m_company")
	if err != nil {
		panic(err.Error())
	}

	columns, err := rows.Columns()

	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	companyNames := make([]string, 0)
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		// カラム番号決め打ち
		companyNames = append(companyNames, string(values[0]))
	}

	for _, companyName := range companyNames {
		// postパラメータ生成
		postParams := url.Values{}
		postParams.Add("type", "DOM-HOLDER")
		postParams.Add("key", companyName)

		// postリクエスト
		resp, _ := http.PostForm(REQUEST_URL, postParams)
		// レスポンスをgo-queryで解析
		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			fmt.Println("Oops Sorry Http Error...")
			fmt.Println(err)
			return
		}

		defer resp.Body.Close()

		fmt.Println(companyName)

		// preタグ内のaタグ内テキストを出力
		doc.Find("pre").Each(func(_ int, s *goquery.Selection) {
			s.Find("a").Each(func(_ int, aSec *goquery.Selection){
				fmt.Println(aSec.Text())
			})
		})
		fmt.Println("==========================================")
		fmt.Println()

		// 同一IP制限に引っかかるのでN秒待機
		time.Sleep(SLEEP_TIME * time.Second)
	}
}

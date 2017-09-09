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

	db := sqlOpen()

	defer db.Close()

	rows := query(db, "SELECT code, name FROM m_company")

	columns, _ := rows.Columns()

	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	companyNames := make(map[string]string, 0)
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		ifErrorNilIsPanic(err)

		companyNames[string(values[0])] = string(values[1])
	}

	for code, companyName := range companyNames {
		// postパラメータ生成
		postParams := makePostParams(companyName)

		// postリクエスト
		resp, _ := http.PostForm(REQUEST_URL, postParams)
		// レスポンスをgo-queryで解析
		doc, err := goquery.NewDocumentFromResponse(resp)
		ifErrorNilIsPanic(err)

		defer resp.Body.Close()

		fmt.Println(companyName)

		// preタグ内のaタグ内テキストを出力
		doc.Find("pre").Each(func(_ int, s *goquery.Selection) {
			s.Find("a").Each(func(_ int, aSec *goquery.Selection){
				_, err = db.Exec(
					"INSERT INTO domain_list (m_company_code, domain, reporting_date) VALUES (?, ?, ?)",
					code,
					aSec.Text(),
					"2017-09-09 00:00:00",
				)
				ifErrorNilIsPanic(err)

				fmt.Println(aSec.Text())
			})
		})
		fmt.Println("==========================================")
		fmt.Println()

		// 同一IP制限に引っかかるのでN秒待機
		sleep()
	}
}

// mysql open
func sqlOpen() *sql.DB {
	db, err := sql.Open("mysql", "root:root@/gois")
	ifErrorNilIsPanic(err)

	return db
}

// if err != nil is panic!
func ifErrorNilIsPanic(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// make url.Values{}
func makePostParams(value string) url.Values {
	postParams := url.Values{}
	postParams.Add("type", "DOM-HOLDER")
	postParams.Add("key", value)

	return postParams
}

// time.Sleep wrapper function
func sleep() {
	time.Sleep(SLEEP_TIME * time.Second)
}

// db.Query wrapper function
func query(db *sql.DB, sql string) *sql.Rows {
	rows, err := db.Query(sql)
	ifErrorNilIsPanic(err)

	return rows
}

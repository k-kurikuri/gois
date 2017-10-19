package main

import (
	"fmt"
	"net/http"
	"net/url"
	"database/sql"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"encoding/json"
	"io/ioutil"
	"github.com/joho/godotenv"
	"log"
	"os"
)

const (
	COMPANY_NAME_COLUMN = "name"
	REQUEST_URL = "http://whois.jprs.jp"
	SLEEP_TIME = 10;
)

type Slack struct {
	Text        string `json:"text"`        //投稿内容
	Username    string `json:"username"`    //投稿者名 or Bot名（存在しなくてOK）
	Icon_emoji  string `json:"icon_emoji"`  //アイコン絵文字
	Icon_url    string `json:"icon_url"`    //アイコンURL（icon_emojiが存在する場合は、適応されない）
	Channel     string `json:"channel"`     //#部屋名
}

func main() {
	fmt.Println("===== func main start!! =====")

	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file not found");
	}

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

		fmt.Println(code + ":" + companyName)

		// preタグ内のaタグ内テキストを出力
		doc.Find("pre").Each(func(_ int, s *goquery.Selection) {
			s.Find("a").Each(func(_ int, aSec *goquery.Selection){
				db := sqlOpen()
				rows := query(db, "SELECT domain FROM domain_list WHERE m_company_code = " + code)
				columns, _ := rows.Columns()
				values := make([]sql.RawBytes, len(columns))
				scanArgs := make([]interface{}, len(values))
				for i := range values {
					scanArgs[i] = &values[i]
				}

				var whoisDomain string = aSec.Text()

				isExist := false
				for rows.Next() {
					rows.Scan(scanArgs...)
					if whoisDomain == string(values[0]) {
						isExist = true
					}
				}

				// 新しいドメインが存在した
				if !isExist {
					fmt.Println("new Domain -> " + whoisDomain)
					_, err = db.Exec(
						"INSERT INTO domain_list (m_company_code, domain, reporting_date) VALUES (?, ?, ?)",
						code,
						whoisDomain,
						"2017-09-09 00:00:00",
					)
					ifErrorNilIsPanic(err)

					noticeToSlack(code, companyName, whoisDomain)
				}

				defer db.Close()
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
	dataSource := os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@/"+os.Getenv("DB_DATABASE")

	db, err := sql.Open("mysql", dataSource)
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

// slackのincoming-web-hookへ通知
func noticeToSlack(code string, companyName string, domain string) {
	params, _ := json.Marshal(Slack{
		"new web-domain register\n" + code + ":" + companyName + "\n" + domain,
		"gois",
		":sushi:",
		"",
		"#company-domain"})

	resp, _ := http.PostForm(
		os.Getenv("INCOMING_URL"),
		url.Values{"payload": {string(params)}},
	)

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	println(string(body))
}

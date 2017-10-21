package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"net/http"
	"net/url"
	"time"
	"github.com/k-kurikuri/gois/slack"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	models "github.com/k-kurikuri/gois/db"
	"strconv"
)

const (
	companyNameColumn = "name"
	requestUrl = "http://whois.jprs.jp"
	sleepTime = 10
)

func main() {
	fmt.Println("===== func main start!! =====")

	err := godotenv.Load()
	ifErrorNilIsPanic(err)

	db := models.DbOpen()

	defer db.Close()

	companies := []models.MCompany{}
	db.Find(&companies)

	for _, company := range companies {
		fmt.Println(strconv.Itoa(company.Code) + ":" + company.Name)

		// postパラメータ生成
		postParams := makePostParams(company.Name)

		// postリクエスト
		resp, _ := http.PostForm(requestUrl, postParams)
		defer resp.Body.Close()

		// レスポンスをgo-queryで解析
		doc, err := goquery.NewDocumentFromResponse(resp)
		ifErrorNilIsPanic(err)

		// preタグ内のaタグ内テキストを出力
		doc.Find("pre").Each(func(_ int, s *goquery.Selection) {
			s.Find("a").Each(func(_ int, aSec *goquery.Selection) {
				domainLists := []models.DomainList{};
				db.Find(&domainLists, "m_company_code=?", company.Code)

				var whoisDomain string = aSec.Text()

				isExist := false
				for _, domainList := range domainLists {
					if whoisDomain == domainList.Domain {
						isExist = true
					}
				}

				// 新しいドメインが存在した
				if !isExist {
					fmt.Println("new Domain -> " + whoisDomain)

					domain := models.DomainList{}
					domain.MCompanyCode = company.Code
					domain.Domain = whoisDomain
					domain.ReportingDate = time.Now()
					db.Create(&domain)

					slack.IncomingWebHook(strconv.Itoa(company.Code), company.Name, whoisDomain)
				}
			})
		})

		fmt.Println("==========================================")
		fmt.Println()

		// 同一IP制限に引っかかるのでN秒待機
		sleep()
	}
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
	time.Sleep(sleepTime * time.Second)
}

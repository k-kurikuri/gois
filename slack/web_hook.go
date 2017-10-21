package slack

import (
	"encoding/json"
	"net/http"
	"io/ioutil"
	"os"
	"net/url"
)

type WebHookParam struct {
	Text      string `json:"text"`
	UserName  string `json:"username"`
	IconEmoji string `json:"icon_emoji"`
	IconUrl   string `json:"icon_url"`
	Channel   string `json:"channel"`
}

// slackのincoming-web-hookへ通知
func IncomingWebHook(code string, companyName string, domain string) {
	params, _ := json.Marshal(WebHookParam{
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

package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"log"
	"net/url"
	"strings"
	"time"
)

type SearchSetting struct {
	TargetUrl string
	TargetRootHost string
	Parse *url.URL
}

func NewSearchTarget(targetUrl string,targetRootHost string)*SearchSetting{
	var err error
	ss := SearchSetting{TargetUrl: targetUrl,TargetRootHost: targetRootHost}
	ss.Parse,err = url.Parse(targetUrl)
	if err != nil{
		panic(err)
	}
	if ss.TargetRootHost == ""{
		ss.TargetRootHost = ss.Parse.Host
	}
	return &ss
}

func main() {
	ss := NewSearchTarget("https://blog.azunyan1111.com/","")
	saveStatus = make(map[string]SaveStatus)
	mapLocker = make(chan bool, 1)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),
		//chromedp.Flag("headless", false),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.ProxyServer("http://127.0.0.1:8085"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	spiderChromeDp(ss, &ctx)
}

func spiderChromeDp(ss *SearchSetting, c *context.Context) {
	// 初回探索でURLを取得していく
	GetUrls(c,ss, ss.TargetUrl)
	// 全ての追加されたURLが探索済になるまで再起的に探し続ける
re:
	i := len(saveStatus) // 残りの数を記録
	for _, value := range saveStatus {
		i--
		if value.SearchStatus == SearchStatusSaved {
			continue
		}
		log.Println("残り:", GetLastCount(), value.Url)
		// 収集
		GetUrls(c,ss,value.Url)
	}

	// 終了再度してなかったら再度探索する
	for _, value := range saveStatus {
		fmt.Print(value)
		if value.SearchStatus == SearchStatusNotSaved {
			goto re
		}
	}
	log.Println("保存したリクエスト数", len(saveStatus))
}

func GetLastCount()int{
	var i int
	for _,v := range saveStatus{
		if v.SearchStatus != SearchStatusSaved{
			i++
		}
	}
	return i
}

type SaveStatus struct {
	Url string
	SearchStatus int
}

var saveStatus map[string]SaveStatus
var mapLocker chan bool

const SearchStatusSaved = 1
const SearchStatusNotSaved = 2

func GetSaveUrl(u string) SaveStatus {
	mapLocker <- true
	defer func() { <-mapLocker }()
	return saveStatus[u]
}

func AddUrl(addUrl, currentUrl string,saveStatusCode int,ss *SearchSetting) {
	mapLocker <- true
	defer func() { <-mapLocker }()
	// 既に探索済の場合はスキップ

	// パースする
	u, err := url.Parse(addUrl)
	if err != nil {
		panic(err)
	}

	// スキームを合わせる
	if u.Scheme == "" {
		u.Scheme = ss.Parse.Scheme
	}
	// フラグメントは削除する
	if u.Fragment != ""{
		u.Fragment = ""
	}
	// ホスト名が無い場合は取得した時のターゲットのURLと同じホスト
	if u.Host == ""{
		u2,err := url.Parse(currentUrl)
		if err != nil{
			panic(err)
		}
		u.Host = u2.Host
	}
	// ホスト名チェックして違うホストは探しに行かない
	if u.Host != ss.Parse.Host && !strings.Contains(u.Host,ss.TargetRootHost){
		return
	}

	var status SaveStatus
	status.Url = u.String()
	status.SearchStatus = saveStatusCode
	if saveStatus[status.Url].SearchStatus == SearchStatusSaved{
		return
	}
	saveStatus[status.Url] = status
}

func GetUrls(c *context.Context,ss *SearchSetting,targetUrl string) {
	defer AddUrl(targetUrl,targetUrl, SearchStatusSaved,ss) // 最後に探索済みにする

	var resp string
	var err error
	err = chromedp.Run(*c,
		chromedp.Navigate(targetUrl),
		chromedp.OuterHTML("html", &resp),
		chromedp.Sleep(time.Second),
	)
	if err != nil && err.Error() != "page load error net::ERR_ABORTED" {
		panic(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp))
	if err != nil {
		panic(err)
	}
	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		val, exists := selection.Attr("href")
		if !exists {
			return
		}
		AddUrl(val,targetUrl, SearchStatusNotSaved,ss)
	})
}

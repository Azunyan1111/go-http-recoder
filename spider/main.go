package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/Azunyan1111/http-recoder/keys"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

func main() {
	// Root CA Client
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(keys.GetCaCert())

	proxyUrl, err := url.Parse("http://127.0.0.1:8080")
	if err != nil{
		panic(err)
	}

	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
			},
			Proxy: http.ProxyURL(proxyUrl),
		},
	}

	saveUrl = make(map[string]bool)

	targetUrl := "https://blog.azunyan1111.com/"
	spider(targetUrl,c,true,10)
}

func spider(targetUrl string,c *http.Client,isSubDomain bool,maxParallelRequest int){
	wg := sync.WaitGroup{}
	ch := make(chan bool,maxParallelRequest)
	mapLocker = make(chan bool,1)

	// 初回探索
	GetUrls(targetUrl,c,isSubDomain)

	re:
	i := len(saveUrl)
	for key,value := range saveUrl{
		i--
		if value{
			continue
		}
		log.Println("残り:",len(saveUrl)-i,key)
		// 収集
		wg.Add(1)
		ch<-true
		go func(targetUrl string,c *http.Client,isSubDomain bool) {
			defer wg.Done()
			defer func() {<-ch}()
			GetUrls(targetUrl,c,isSubDomain)
		}(key,c,isSubDomain)
	}
	wg.Wait()

	// 終了再度してなかったら再度探索する
	for _,value := range saveUrl{
		if !value{
			goto re
		}
	}
	log.Println("保存リクエスト数",len(saveUrl))
}

var saveUrl map[string]bool
var mapLocker chan bool

func GetSaveUrl(u string)bool{
	mapLocker<-true
	defer func() {<-mapLocker}()
	return saveUrl[u]
}

func SetSaveUrl(u string,b bool){
	mapLocker<-true
	defer func() {<-mapLocker}()
	saveUrl[u] = b
}

func GetUrls(u string,c *http.Client,isSubDomain bool){
	// 探索済みにする
	SetSaveUrl(u,true)
	path,err := url.Parse(u)
	if err != nil{
		panic(err)
	}
	// ゴミパスを破棄する
	if path.Scheme == ""{
		log.Println("Error:",u)
		return
	}

	resp,err := c.Get(u)
	if err != nil{
		panic(err)
	}
	doc,err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil{
		panic(err)
	}
	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		val, exists := selection.Attr("href")
		if !exists{
			return
		}
		p,err := url.Parse(val)
		if err != nil{
			return
		}
		// 普通にホスト名だけで判定。ホスト名と/から始まる場合に対応
		if p.Host == path.Host || p.Host == ""{
			if p.Host == ""{
				val = path.Scheme + "://" + path.Host + val
			}
			// 保存
			if !GetSaveUrl(val){
				SetSaveUrl(val,false)
			}
		}
		// サブドメインの許可
		if isSubDomain{
			if strings.Contains(val,path.Host){
				// 保存
				if !GetSaveUrl(val){
					SetSaveUrl(val,false)
				}
			}
		}
	})
}
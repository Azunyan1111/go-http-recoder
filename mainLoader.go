package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/Azunyan1111/http-recoder/keys"
	"github.com/elazarl/goproxy"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	goproxyCa, _ := tls.X509KeyPair(keys.GetCaCert(), keys.GetCaKey())
	goproxyCa.Leaf, _ = x509.ParseCertificate(goproxyCa.Certificate[0])
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}

	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*$"))).
		HandleConnect(goproxy.AlwaysMitm)

	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		// Generate file path
		var filePath []string
		filePath = append(filePath, resp.Request.URL.Scheme)
		filePath = append(filePath, resp.Request.URL.Host)
		paths := strings.Split(resp.Request.URL.Path, "/")
		filePath = append(filePath, paths...)
		filePath = append(filePath, resp.Request.URL.RawQuery)
		filePath = append(filePath, strconv.Itoa(resp.StatusCode))
		rowFilePath := "save-requests"
		for _, a := range filePath {
			if a == "" {
				continue
			}
			rowFilePath += "/" + a
		}
		rowFilePath = strings.ReplaceAll(rowFilePath,":","")
		// フォルダの作成
		fmt.Println(rowFilePath)
		// ファイル存在確認
		if Exists(rowFilePath + ".dump"){
			// 存在するのでそれを返す
			b, err := ioutil.ReadFile(rowFilePath + ".dump")
			if err != nil {
				panic(err)
			}
			cache, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(b)), resp.Request)
			if err != nil {
				panic(err)
			}
			return cache
		}
		return resp
	})
	log.Fatal(http.ListenAndServe(":8080", proxy))
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
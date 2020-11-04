package main

import (
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/Azunyan1111/http-recoder/keys"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"net/http/httputil"
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

	// ----------------------
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp.Request.URL.Scheme == "http://" {
			return resp
		}
		// Generate file path
		var filePath []string
		filePath = append(filePath, resp.Request.URL.Scheme)
		filePath = append(filePath, resp.Request.URL.Host)
		paths := strings.Split(resp.Request.URL.Path, "/")
		filePath = append(filePath, paths...)
		filePath = append(filePath, fmt.Sprintf("%x",md5.Sum([]byte(resp.Request.URL.RawQuery))))
		filePath = append(filePath, strconv.Itoa(resp.StatusCode))
		rowFilePath := "save-requests"
		for _, a := range filePath {
			if a == "" {
				continue
			}
			rowFilePath += "/" + a
		}
		// フォルダの作成
		fmt.Println(rowFilePath)
		err := os.MkdirAll(strings.Join(strings.Split(rowFilePath, "/")[:len(strings.Split(rowFilePath, "/"))-1], "/"), 0755)
		if err != nil {
			panic(err)
		}
		// ファイルの書き込み
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			panic(err)
		}
		file, err := os.OpenFile(rowFilePath+".dump", os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			panic(err)
		}
		_, err = file.Write(dump)
		if err != nil {
			panic(err)
		}
		return resp
	})

	log.Fatal(http.ListenAndServe(":8080", proxy))
}


# Introduction

This is an application that records http requests with the recorder and returns the file recorded by the reader
これはhttpリクエストをレコーダーで記録してリーダーで記録したファイルを返すアプリケーョンです

It is easy to create a phishing site by modifying the recorded response content.
記録したレスポンス内容を改変してフィッシングサイトを簡単に作成することができます。

You will be warned if you do not load the certificate you generated into your browser
あなたが生成した証明書をブラウザに読み込まないと警告がでます

# How To Use

Generate CA
```
cd keys
openssl genrsa -aes256 -passout pass:1 -out ca.key.pem 4096
openssl rsa -passin pass:1 -in ca.key.pem -out ca.key.pem.tmp
mv ca.key.pem.tmp ca.key.pem
openssl req -config openssl.cnf -key ca.key.pem -new -x509 -days 7300 -sha256 -extensions v3_ca -out ca.pem
```

Recode
```
go run mainRecode.go
```

Modify the response saved in "save-requests"!

Load
```
go run mainLoader.go
```

You Win^〜

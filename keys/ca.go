package keys

import (
	"io/ioutil"
	"os"
)


var myCaCert []byte
var myCaKey []byte

func init(){
	key,err := os.Open("./keys/ca.key.pem")
	if err != nil{
		panic(err)
	}
	myCaKey,err = ioutil.ReadAll(key)
	if err != nil{
		panic(err)
	}
	ca,err := os.Open("./keys/ca.pem")
	if err != nil{
		panic(err)
	}
	myCaCert,err = ioutil.ReadAll(ca)
	if err != nil{
		panic(err)
	}
}

func GetCaCert()[]byte{
	return myCaCert
}
func GetCaKey()[]byte{
	return myCaKey
}

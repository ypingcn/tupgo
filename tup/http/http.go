package http

import (
	"io/ioutil"
	"net/http"
	"strings"

	tup "github.com/ypingcn/tupgo/tup"
)

// DoSimpleHTTPRequest HTTP request
func DoSimpleHTTPRequest(method string, url string, reqBody string, headers map[string]string) (rspBody string, err error) {
	httpClient := &http.Client{}

	httpReq, err := http.NewRequest(method, url, strings.NewReader(string(reqBody)))

	if err != nil {
		return "", err
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	httpRsp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}

	defer httpRsp.Body.Close()

	httpBody, err := ioutil.ReadAll(httpRsp.Body)

	if err != nil {
		return "", err
	}

	return string(httpBody), nil
}

//DoSimpleTUPHTTPRequest TUP request using TUPStruct
func DoSimpleTUPHTTPRequest(method string, url string, servantName string, funcName string, req map[string]tup.TUPStruct, rsp map[string]tup.TUPStruct) error {
	tupReq := tup.TarsUniPacket{}

	tupReq.Init()
	tupReq.SetVersion(3)
	tupReq.SetServantName(servantName)
	tupReq.SetFuncName(funcName)

	for key, value := range req {
		if err := tupReq.Put(key, value); err != nil {
			return err
		}
	}

	reqBody, err := tupReq.Encode()
	if err != nil {
		return err
	}

	rspBody, err := DoSimpleHTTPRequest(method, url, string(reqBody), nil)
	if err != nil {
		return err
	}

	tupRsp := tup.TarsUniPacket{}
	tupRsp.Init()
	tupRsp.SetVersion(3)

	err = tupRsp.Decode([]byte(rspBody))
	if err != nil {
		return err
	}

	for key, value := range rsp {
		if err := tupRsp.Get(key, value); err != nil {
			return err
		}
	}

	return nil
}

// DoSimpleTUPHTTPRequest2 TUP request using TarsUniPacket
func DoSimpleTUPHTTPRequest2(method string, url string, req *tup.TarsUniPacket, rsp *tup.TarsUniPacket) error {
	reqBody, err := req.Encode()
	if err != nil {
		return err
	}

	rspBody, err := DoSimpleHTTPRequest(method, url, string(reqBody), nil)
	if err != nil {
		return err
	}

	err = rsp.Decode([]byte(rspBody))
	if err != nil {
		return err
	}

	return nil
}

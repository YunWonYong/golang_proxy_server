package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type (
	RequestBody struct {
		Url string `json:"url"`
		Method string `json:"method"`
		Body map[string]interface{} `json:"body"`
		Header map[string]interface{} `json:"header"`
	}
)

func Start(address string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", router)
	return http.ListenAndServe(address, mux)
}

func router(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte("only post method!!!"))
		return
	}
	body := request.Body
	defer body.Close()
	buff, err := ioutil.ReadAll(body)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(fmt.Sprintf("router ioutil.ReadAll fail. err: %s", err.Error())))
		return
	}
	rb := new(RequestBody)
	if err := json.Unmarshal(buff, rb); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(fmt.Sprintf("json.Unmarshal fail. err: %s, body: %s", err.Error(), buff)))
		return
	}
	statusCode, buff, err := handler(rb)
	if err != nil {
		response.WriteHeader(statusCode)
		response.Write([]byte(fmt.Sprintf("handler fail. err: %s", err.Error())))
		return
	}
	
	response.WriteHeader(statusCode)
	response.Write(buff)
}

func handler(body *RequestBody) (int, []byte, error) {
	if len(body.Method) == 0 {
		return http.StatusInternalServerError, nil, errors.New("redirect http method data empty")
	}

	var (
		request *http.Request
		err error
		bodyByte []byte
		bodyBuff *bytes.Buffer = nil
	)
	
	if len(body.Body) > 0 {
		bodyByte, err := json.Marshal(body.Body)
		if err != nil {
			return http.StatusInternalServerError, nil, fmt.Errorf("body data marshal fail. err: %s", err.Error())
		}
		temp := bytes.NewBuffer(bodyByte)
		if temp.Len() > 0 {
			bodyBuff = temp
		}
	}

	request, err = http.NewRequest(strings.ToUpper(body.Method), body.Url, bodyBuff)
	if err != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf("http.NewRequest(%s, %s, %#v) fail. err: %s", body.Method, body.Url, bodyByte, err.Error())
	}
	defer request.Body.Close()

	if len(body.Header) > 0 {
		request.Header = getHeaders(body.Header)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf("http.DefaultClient.Do fail. err: %s", err.Error())
	}
	defer response.Body.Close()

	var (
		responseBodyBuff []byte = nil
	)
	if response.Body != nil {
		responseBodyBuff, err = ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("ioutil.ReadAll(response.Body) fail.")
		}
	}

	return response.StatusCode, responseBodyBuff, nil
}

func  getHeaders(headers map[string]interface{}) http.Header {
	header := make(http.Header)

	for key, value := range headers {
		switch t := value.(type) {
		case string:
			header.Set(key, value.(string))
		case []string:
			for _, v := range value.([]string) {
				header.Add(key, v)
			}
		default:
			fmt.Println("getHeaders", t)
			continue
		}
	}

	return header
}
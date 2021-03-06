package http

import (
	"bytes"
	"fmt"
	"git.jiayincloud.com/TestDev/pegasus.git/utils"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type RequestData struct {
	Url    string
	Method string
	// Content-Type will be auto reset if Form or File has values
	Headers map[string]string
	// Body should be nil if Form or File has values
	Body        string
	QueryParams map[string]string
	PathParams  []string
	File        map[string]string
	Form        map[string]string
}

func (r *RequestData) Request(client http.Client, results chan map[string]interface{}) {
	var (
		url        string
		err        error
		req        *http.Request
		body       []byte
		statusCode int
	)

	if r.PathParams != nil {
		url = fmt.Sprintf(r.Url, utils.UnpackString(r.PathParams)...)
	} else {
		url = r.Url
	}

	if r.Form != nil || r.File != nil {
		body, contentType := multipartForm(r.File, r.Form)
		req, err = http.NewRequest(r.Method, url, bytes.NewBuffer(body.Bytes()))
		q := req.URL.Query()
		utils.Check(err)
		req.TransferEncoding = []string{"UTF-8"}
		for k, v := range r.Headers {
			req.Header.Add(k, v)
		}
		for k, v := range r.QueryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
		req.Header.Add("Content-Type", contentType)
	} else {
		req, err = http.NewRequest(r.Method, url, bytes.NewBuffer([]byte(r.Body)))
		q := req.URL.Query()
		utils.Check(err)
		req.TransferEncoding = []string{"UTF-8"}
		for k, v := range r.Headers {
			req.Header.Add(k, v)
		}
		for k, v := range r.QueryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	startTime := int(time.Now().UnixNano() / int64(math.Pow10(6)))
	resp, err := client.Do(req)
	if err != nil {
		utils.Check(err)
	} else {
		statusCode = resp.StatusCode
		body, err = ioutil.ReadAll(resp.Body)
		utils.Check(err)
		defer resp.Body.Close()
	}
	endTime := int(time.Now().UnixNano() / int64(math.Pow10(6)))
	data := make(map[string]interface{})
	data["statusCode"] = statusCode
	data["body"] = string(body[:])
	data["endTime"] = endTime
	data["startTime"] = startTime
	data["error"] = err
	data["elapsed"] = endTime - startTime
	results <- data
}

func multipartForm(file map[string]string, form map[string]string) (body bytes.Buffer, contentType string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// add form file
	for k, v := range file {
		f, err := os.Open(v)
		utils.Check(err)
		fw, err := w.CreateFormFile(k, v)
		utils.Check(err)
		_, err = io.Copy(fw, f)
		f.Close()
		utils.Check(err)
	}
	// add form data
	for k, v := range form {
		fw, err := w.CreateFormField(k)
		utils.Check(err)
		_, err = fw.Write([]byte(v))
		utils.Check(err)
	}
	w.Close()

	return b, w.FormDataContentType()
}

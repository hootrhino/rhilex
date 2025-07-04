package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hootrhino/rhilex/glogger"
)

/*
*
* HTTP POST
*
 */
func Post(client http.Client, data any,
	url string, headers map[string]string) (string, error) {
	bites, errs1 := json.Marshal(data)
	if errs1 != nil {
		glogger.GLogger.Error(errs1)
		return "", errs1
	}
	body := strings.NewReader(string(bites))
	request, _ := http.NewRequest("POST", url, body)
	request.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	response, err2 := client.Do(request)
	if err2 != nil {
		return "", err2
	}
	if response.StatusCode != 200 {
		bytes0, err3 := io.ReadAll(response.Body)
		if err3 != nil {
			return "", err3
		}
		return "", fmt.Errorf("Error:%v", string(bytes0))
	}
	var r []byte
	response.Body.Read(r)
	bytes1, err3 := io.ReadAll(response.Body)
	if err3 != nil {
		return "", err3
	}
	return string(bytes1), nil
}

/*
*
* HTTP GET
*
 */
func Get(client http.Client, url string) string {
	var err error
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		glogger.GLogger.Error(err)
		return ""
	}

	response, err := client.Do(request)
	if err != nil {
		glogger.GLogger.Error(err)
		return ""
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		glogger.GLogger.Error(err)
		return ""
	}
	return string(body)
}

type response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (O response) String() string {
	return fmt.Sprintf("Code:%v, Error: %s", O.Code, O.Msg)
}

// Download @param: 机器码
func Download(url, param, filePath string) error {
	// 创建请求体
	body := []byte(`{"param":"` + param + `"}`)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(req.Body)
		response := response{}
		json.Unmarshal(body, &response)
		return fmt.Errorf("error:%s", response.String())
	}
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

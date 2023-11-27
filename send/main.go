package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	data := map[string]interface{}{
		"PassWord": "123",
		"Args":     []string{"123", "hah"},
	}
	PostJson("http://localhost:8080/get", data, r)

}

func PostJson(uri string, data map[string]interface{}, router *gin.Engine) {

	// 将JSON数据转换为字节切片
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// 设置HTTP请求
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发起HTTP请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	// 处理响应
	// 读取响应体内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// 将字节切片转换为字符串，并输出
	responseBodyString := string(body)
	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", responseBodyString)
	// 这里你可以根据需要读取和处理响应的内容
}

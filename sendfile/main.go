package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

const PUT_FILE_URL = "http://localhost:8080/putfile/"
const GET_FILE_URL = "http://localhost:8080/getfile/"

func main() {
	TestSendFile()
	TestGetFile()
}

func TestGetFile() {
	filename := "testfile.txt"
	// 创建目标文件
	outFile, err := os.Create(filename)
	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}
	defer outFile.Close()

	// 发送 HTTP GET 请求下载文件
	response, err := http.Get(GET_FILE_URL + filename)
	if err != nil {
		fmt.Println("Failed to download file:", err)
		return
	}
	defer response.Body.Close()

	// 检查 HTTP 状态码
	if response.StatusCode != http.StatusOK {
		fmt.Println("Server returned non-200 status code:", response.Status)
		return
	}

	// 将响应体写入文件
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		fmt.Println("Failed to save file:", err)
		return
	}

	fmt.Println("File downloaded successfully")
}

func TestSendFile() {
	file, err := os.Open("testfile.txt") // 替换为你要上传的文件路径
	if err != nil {
		fmt.Println("Failed to open file:", err)
		return
	}
	defer file.Close()

	// 创建一个 buffer 用于保存文件数据
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 创建一个文件表单字段
	fileField, err := writer.CreateFormFile("file", "testfile.txt")
	if err != nil {
		fmt.Println("Failed to create form file:", err)
		return
	}

	// 将文件内容复制到表单字段中
	_, err = io.Copy(fileField, file)
	if err != nil {
		fmt.Println("Failed to copy file to form field:", err)
		return
	}

	// 关闭 multipart writer
	writer.Close()

	// 创建 POST 请求发送文件
	request, err := http.NewRequest("POST", PUT_FILE_URL, &requestBody)
	if err != nil {
		fmt.Println("Failed to create request:", err)
		return
	}

	// 设置请求头
	request.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return
	}
	defer response.Body.Close()

	// 打印服务器响应
	fmt.Println("Server response:", response.Status)
}

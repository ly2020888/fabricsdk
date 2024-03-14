package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	loglevel     = "TAPE_LOGLEVEL"
	CHANCODENAME = "fabcar"
)

type Server struct {
	proposer *Proposer
	logger   *log.Logger
	router   *gin.Engine
}

// Define a struct to represent the success response
type PutSuccessResponse struct {
	Message  string `json:"message"`
	Playload string `json:"playload"`
}

// @Success 200 {object} PutSuccessResponse

// Define a struct to represent the failure response
type ErrorResponse struct {
	Error string `json:"error"`
}

// @Failure 400 {object} ErrorResponse
// @Failure 502 {object} ErrorResponse

func NewServer(proposer *Proposer, logger *log.Logger) *Server {
	return &Server{
		proposer: proposer,
		logger:   logger,
		router:   gin.Default(),
	}
}

func (s *Server) setupRouter() {
	gin.SetMode(gin.ReleaseMode)
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	s.logger.Infof("已经连接至Fabic网络, 等待指令...")

	s.router.POST("/put", s.handlePut)
	s.router.POST("/get", s.handleGet)
	s.router.GET("/getfile", s.handleGetFile)
	s.router.POST("/putfile", s.handlePutFile)

	s.router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, "Pong")
	})
}

func SaveFileAndGetBytes(file multipart.File, header *multipart.FileHeader) ([]byte, error) {
	// 确保关闭文件
	defer file.Close()

	// 从 header 中获取文件名
	filename := header.Filename

	// 创建文件保存路径
	filePath := filepath.Join("files", filename)

	// 创建一个文件，准备将数据写入其中
	out, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	// 将文件内容写入到文件中
	_, err = io.Copy(out, file)
	if err != nil {
		return nil, err
	}

	// 读取文件内容到字节切片中
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 返回文件内容的字节切片
	return data, nil
}

// @Summary 上传文件
// @Description 从客户端上传文件到服务器并将文件存储到本地和区块链中
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Success 200 {string} string "文件上传成功"
// @Failure 400 {string} string "请求错误"
// @Failure 500 {string} string "服务器内部错误"
// @Router /putfile [post]
func (s *Server) handlePutFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.String(400, "Bad request: "+err.Error())
		return
	}
	filename := header.Filename

	filedata, err := SaveFileAndGetBytes(file, header)
	if err != nil {
		c.String(500, "Failed to write file: "+err.Error())
		return
	}
	// blockchain
	var Args []string
	Args = append(Args, filename)
	playload, err := s.proposer.Query("KeyExists", Args)
	if err != nil {
		s.logger.Error(err)
	} else {
		if string(playload) == "true" {
			s.logger.Infof("文件已经存在:%v", Args)
		} else {

			byteSliceSlice := [][]byte{[]byte(filename), filedata}
			s.logger.Infof("文件开始上链:%v", filename)

			playload, err = s.proposer.Exec("Createhash", byteSliceSlice)
			if err != nil {
				c.String(500, "Failed to Upload file to blockchain: "+err.Error())
				return
			}
			c.String(200, "File uploaded successfully: "+filename+string(playload))

		}
	}
	c.String(200, "File uploaded successfully: "+filename)
}

// @Summary 下载文件
// @Description 从服务器下载指定文件
// @Accept json
// @Produce octet-stream
// @Param filename path string true "要下载的文件名"
// @Success 200 {file} octet-stream "文件流"
// @Failure 404 {string} string "文件未找到"
// @Router /getfile/{filename} [GET]
func (s *Server) handleGetFile(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		s.logger.Error("请求文件名为空")
		c.String(500, "请求文件名为空 ")
		return
	} else {
		s.logger.Infof("请求文件名为:%s", filename)

	}

	file, err := os.Open(filepath.Join("files", filename))
	if err != nil {
		s.logger.Infof("本地文件不存在，正在同步缓存...")

		args := []string{filename}
		playload, err := s.proposer.Query("GetFile", args)
		if err != nil {
			s.logger.Errorf("未查到到区块链上有该文件")
			c.String(500, "未查到到区块链上有该文件")
			return
		}

		file, err = os.Create(filename)
		if err != nil {
			s.logger.Infof("Failed to create file: %v\n", err)
			return
		}
		defer file.Close()

		_, err = file.Write(playload)
		if err != nil {
			s.logger.Infof("Failed to write payload to file: %v\n", err)
			return
		}
	}
	defer file.Close()
	s.logger.Infof("区块链客户端节点检测到文件，开始发送%v", filename)

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	io.Copy(c.Writer, file)

}

// @Summary 将数据上链
// @Description 将数据上链到Fabric
// @Accept json
// @Produce json
// @Param request body Message true "请求参数"
// @Success 200 {object} PutSuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 502 {object} ErrorResponse
// @Example request
//
//	{
//	  "PassWord": "这里填写内建密码",
//	  "Args": ["000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1", "bitcoin first block is here"]
//	}
//
// @Example response
//
//	{
//	  "message": "数据上链成功",
//	  "playload": ""
//	}
//
// @Router /put [post]
func (s *Server) handlePut(ctx *gin.Context) {
	var msg Message
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error binding JSON"})
		return
	}

	if Temporary == MD5(msg.PassWord) {
		byteSliceSlice := make([][]byte, len(msg.Args))
		for i, str := range msg.Args {
			byteSliceSlice[i] = []byte(str)
		}
		playload, err := s.proposer.Exec("Createhash", byteSliceSlice)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"message": "数据上链失败", "error": fmt.Sprint(err)})
		} else {
			s.logger.Infof("Fabric调用成功, 合约参数为:%v", msg.Args)
			ctx.JSON(http.StatusOK, gin.H{"message": "数据上链成功", "playload": string(playload)})
		}
	} else {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": "密码错误，拒绝访问"})
	}
}

// @Summary 数据查询
// @Description 从 Fabric blockchain 获取数据
// @Accept json
// @Produce json
// @Param request body Message true "请求参数"
// @Success 200 {object} PutSuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 502 {object} ErrorResponse
// @Example request
//
//	{
//	  "PassWord": "这里填写内建密码",
//	  "Args": ["000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1"]
//	}
//
// @Example response
//
//	{
//	  "message": "数据查询成功",
//	  "payload": "bitcoin first block is here"
//	}
//
// @Router /get [post]
func (s *Server) handleGet(ctx *gin.Context) {
	var msg Message
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error binding JSON"})
		return
	}

	if Temporary == MD5(msg.PassWord) {
		playload, err := s.proposer.Query("Queryhash", msg.Args)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"message": "数据查询失败", "error": fmt.Sprint(err)})
		} else {
			s.logger.Infof("Fabric调用成功, 合约参数为:%v", msg.Args)
			ctx.JSON(http.StatusOK, gin.H{"message": "数据查询成功", "playload": string(playload)})

		}
		s.logger.Infof("Fabric调用合约参数为:%v", msg.Args)
	} else {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": "密码错误，拒绝访问"})
	}
}

func (s *Server) Run() {
	s.setupRouter()
	go ScheduledPush(s.logger)
	s.router.Run(":8085")
}

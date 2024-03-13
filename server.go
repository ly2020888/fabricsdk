package main

import (
	"fmt"
	"io"
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
	s.router.POST("/getfile", s.handleGetFile)
	s.router.POST("/putfile", s.handlePutFile)

	s.router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, "Pong")
	})
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
	out, err := os.Create(filepath.Join("files", filename))
	if err != nil {
		c.String(500, "Failed to create file: "+err.Error())
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		c.String(500, "Failed to save file: "+err.Error())
		return
	}

	var Args []string
	Args = append(Args, filename)
	playload, err := s.proposer.Query("KeyExists", Args)
	if err != nil {
		s.logger.Error(err)
	} else {
		if string(playload) == "true" {
			s.logger.Infof("Fabric调用成功, 文件已经存在:%v", Args)
		} else {
			var Args [][]byte
			Args = append(Args, []byte(filename))
			data, err := io.ReadAll(file)
			if err != nil {
				s.logger.Errorf("Failed to read multipart file:%v", err)
				return
			}
			Args = append(Args, []byte(data))
			playload, err = s.proposer.Exec("UploadFile", Args)
			if err != nil {
				c.String(500, "Failed to Upload file to blockchain: "+err.Error())

			}

			c.String(200, "File uploaded successfully: "+filename+string(playload))
		}
	}

}

// @Summary 下载文件
// @Description 从服务器下载指定文件
// @Accept json
// @Produce octet-stream
// @Param filename path string true "要下载的文件名"
// @Success 200 {file} octet-stream "文件流"
// @Failure 404 {string} string "文件未找到"
// @Router /getfile/{filename} [POST]
func (s *Server) handleGetFile(c *gin.Context) {
	filename := c.Param("filename")
	file, err := os.Open(filepath.Join("files", filename))
	if err != nil {
		c.String(404, "File not found in local"+err.Error())
		return
	}
	defer file.Close()
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
	s.router.Run(":8080")
}

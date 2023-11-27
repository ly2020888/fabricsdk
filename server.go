package main

import (
	"net/http"

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
	s.router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, "Pong")
	})
}

// @Summary 将数据上链
// @Description 将数据上链到Fabric
// @Accept json
// @Produce json
// @Param request body Message true "请求参数"
// @Success 200 {object} PutSuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 502 {object} ErrorResponse
// @Router /put [post]
func (s *Server) handlePut(ctx *gin.Context) {
	var msg Message
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error binding JSON"})
		return
	}

	if Temporary == MD5(msg.PassWord) {
		playload, err := s.proposer.Query("Createhash", msg.Args)
		if err != nil {
			s.logger.Error(err)
		} else {
			s.logger.Infof("Fabric调用成功, 合约参数为:%v", msg.Args)
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "数据上链", "playload": string(playload)})
	} else {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": "密码错误，拒绝访问"})
	}
}

// @Summary 数据查询
// @Description 从Fabric blockchain 获取数据
// @Accept json
// @Produce json
// @Param request body Message true "请求参数"
// @Success 200 {object} PutSuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 502 {object} ErrorResponse
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
			s.logger.Error(err)
		} else {
			s.logger.Infof("Fabric调用成功, 合约参数为:%v", msg.Args)
		}
		s.logger.Infof("Fabric调用合约参数为:%v", msg.Args)
		ctx.JSON(http.StatusOK, gin.H{"message": "数据查询", "playload": string(playload)})
	} else {
		ctx.JSON(http.StatusBadGateway, gin.H{"message": "密码错误，拒绝访问"})
	}
}

func (s *Server) Run() {
	s.setupRouter()
	s.router.Run(":8080")
}

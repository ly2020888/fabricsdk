package main

import (
	"crypto/md5"
	"fmt"
	"os"

	_ "main/docs"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var Temporary string

type Message struct {
	PassWord string   `json:"PassWord" swaggertype:"string" example:"31e934ff763ae46"`
	Args     []string `json:"Args" swaggertype:"string"  example:"['000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1', 'bitcoin first block is here']"`
}

var (
	app  = kingpin.New("tapePlus", "Efficient TAPE-based client")
	run  = app.Command("run", "Start the tapePlus program").Default()
	pw   = run.Flag("password", "A memory key that needs to be set").Required().Short('p').String()
	name = run.Flag("chaincode", "chaincode name").Short('n').Default("fabcar").String()
	ch   = run.Flag("channel", "channel name").Short('c').Default("mychannel").String()
)

func main() {
	// 创建Fabric SDK实例
	var err error
	logger := log.New()
	logger.SetLevel(log.InfoLevel)
	if customerLevel, customerSet := os.LookupEnv(loglevel); customerSet {
		if lvl, err := log.ParseLevel(customerLevel); err == nil {
			logger.SetLevel(lvl)
		}
	}

	fullCmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	Temporary = MD5(*pw)
	logger.Infof("内建密码为:%s\n", *pw)

	switch fullCmd {
	case run.FullCommand():
		err = start(logger)
		if err != nil {
			logger.Error(err)
		}
	default:
		err = errors.Errorf("invalid command: %s", fullCmd)
	}

	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	os.Exit(0)

}
func start(logger *log.Logger) error {
	sdk, err := fabsdk.New(config.FromFile("./config.yaml"))
	if err != nil {
		logger.Errorf("Failed to create SDK: %v", err)
		return err
	}
	defer sdk.Close()

	clientChannelContext := sdk.ChannelContext(*ch, fabsdk.WithUser("User1"))
	client, err := channel.New(clientChannelContext)
	if err != nil {
		logger.Errorf("Failed to create channel client: %v", err)
		return err

	}

	pro := CreateProposer(*name, client, logger)
	server := NewServer(pro, logger)
	server.Run()
	return nil
}

func MD5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str
}

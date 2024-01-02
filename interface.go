package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	apiURL       = "http://120.236.202.46:18301/university/open/kt2/blockchain/performance"
	clientID     = "1"
	clientSecret = "1111_test"
)

type RequestBody struct {
	MaxTransactionsPS        int `json:"maxTransactionsPS"`
	AverageThroughput        int `json:"averageThroughput"`
	NumOfDataEntries         int `json:"numOfDataEntries"`
	NumOfBCInstructForDrones int `json:"numOfBCInstructForDrones"`
	NumOfDroneSA             int `json:"numOfDroneSA"`
}

func ScheduledPush(logger *log.Logger) {
	// Define a ticker, triggering every 3 seconds
	ticker := time.NewTicker(20 * time.Second)

	// Start an infinite loop
	for {
		// Wait for the ticker to trigger
		<-ticker.C

		// Prepare data for the POST request
		maxTransactionsPS := 180            //每秒最大交易数量
		averageThroughput := rand.Intn(100) //平均吞吐率
		numOfDataEntries := 946             // 区块链上链的数据数量
		numOfBCInstructForDrones := 269     //无人机区块链指令数量
		numOfDroneSA := 4                   //无人机安全关联的数量

		// Prepare data for the POST request
		requestData := RequestBody{
			MaxTransactionsPS:        maxTransactionsPS,
			AverageThroughput:        averageThroughput,
			NumOfDataEntries:         numOfDataEntries,
			NumOfBCInstructForDrones: numOfBCInstructForDrones,
			NumOfDroneSA:             numOfDroneSA,
		}

		// Convert data to JSON format
		jsonData, err := json.Marshal(requestData)
		if err != nil {
			fmt.Println("JSON marshaling failed:", err)
			continue
		}
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("clientId", clientID)
		req.Header.Set("clientSecret", clientSecret)

		// Send a POST request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("POST request failed:", err)
			continue
		}

		// Print response information
		logger.Infof("POST request Yunfeng push message success, status:%v\n", resp.Status)
		resp.Body.Close()
	}
}

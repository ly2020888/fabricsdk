package main

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	log "github.com/sirupsen/logrus"
)

type Proposer struct {
	worker   *channel.Client
	logger   *log.Logger
	chancode string
}

func CreateProposer(chancode string, sdk *channel.Client, logger *log.Logger) *Proposer {

	return &Proposer{chancode: chancode, worker: sdk, logger: logger}
}

func (ps *Proposer) Query(fcn string, Args []string) ([]byte, error) {
	ps.logger.Infof("Start sending transactions.")

	var argsAsBytes [][]byte
	for _, arg := range Args {
		argsAsBytes = append(argsAsBytes, []byte(arg))
	}
	response, err := ps.worker.Query(channel.Request{
		ChaincodeID: ps.chancode,
		Fcn:         fcn,
		Args:        argsAsBytes,
	})
	if err != nil {
		ps.logger.Errorf("Failed to query: %v\n", err)
		return nil, err
	}

	// 处理查询结果
	if response.ChaincodeStatus != 200 {
		ps.logger.Errorf("Chaincode query failed with status: %d - %s\n", response.ChaincodeStatus, string(response.Payload))
		return nil, err
	}
	ps.logger.Infof("Smart Contract status: %s\n", string(response.ChaincodeStatus))

	ps.logger.Infof("Smart Contract Output: %s\n", string(response.Payload))
	return response.Payload, nil
}

func (ps *Proposer) Exec(fcn string, Args []string) ([]byte, error) {
	ps.logger.Infof("Start sending transactions.")

	var argsAsBytes [][]byte
	for _, arg := range Args {
		argsAsBytes = append(argsAsBytes, []byte(arg))
	}
	response, err := ps.worker.Execute(channel.Request{
		ChaincodeID: ps.chancode,
		Fcn:         fcn,
		Args:        argsAsBytes,
	})
	if err != nil {
		ps.logger.Errorf("Failed to Execute: %v\n", err)
		return nil, err
	}

	// 处理查询结果
	if response.ChaincodeStatus != 200 {
		ps.logger.Errorf("Chaincode Execute failed with status: %d - %s\n", response.ChaincodeStatus, string(response.Payload))
		return nil, err
	}
	ps.logger.Infof(" Output: %s\n", string(response.Payload))
	return response.Payload, nil
}

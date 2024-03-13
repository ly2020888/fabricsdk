package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// KVContract 表示一个简单的 key-value 智能合约
type KVContract struct {
	contractapi.Contract
}

// UploadKV 上传一个 key-value 对到账本中
func (kvc *KVContract) Createhash(ctx contractapi.TransactionContextInterface, key string, value string) error {
	if len(key) == 0 || len(value) == 0 {
		return fmt.Errorf("key and value must not be empty")
	}

	return ctx.GetStub().PutState(key, []byte(value))
}

// UploadFile 上传一个二进制文件到账本中
func (kvc *KVContract) UploadFile(ctx contractapi.TransactionContextInterface, key string, data []byte) error {
	if len(key) == 0 || len(data) == 0 {
		return fmt.Errorf("key and data must not be empty")
	}

	return ctx.GetStub().PutState(key, data)
}

// GetKV 获取指定 key 的值
func (kvc *KVContract) Queryhash(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	value, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	if value == nil {
		return "", fmt.Errorf("the key %s does not exist", key)
	}

	return string(value), nil
}

// GetFile 获取指定 key 对应的二进制文件数据
func (kvc *KVContract) GetFile(ctx contractapi.TransactionContextInterface, key string) ([]byte, error) {
	data, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if data == nil {
		return nil, fmt.Errorf("the key %s does not exist", key)
	}

	return data, nil
}

func (kvc *KVContract) KeyExists(ctx contractapi.TransactionContextInterface, key string) (bool, error) {
	exists, err := ctx.GetStub().GetState(key)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return exists != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&KVContract{})
	if err != nil {
		fmt.Printf("Error creating KVContract chaincode: %v\n", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting KVContract chaincode: %v\n", err)
	}
}

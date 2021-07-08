package api

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

func TransferEth(client *ethclient.Client, privateKeyHex string, toAddressHex string, amount int64) (txHash [32]byte, err error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return common.Hash{},err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Hash{},errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return common.Hash{},err
	}

	value := big.NewInt(amount)
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return common.Hash{},err
	}

	toAddress := common.HexToAddress(toAddressHex)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return common.Hash{},err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return common.Hash{},err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return common.Hash{},err
	}

	return signedTx.Hash(),nil
}

func GetBalance(clientUrl string, addressHex string) (balance *big.Int, err error) {
	client, err := ethclient.Dial(clientUrl)
	if err != nil {
		return nil,err
	}

	account := common.HexToAddress(addressHex)
	b, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil,err
	}
	return b,nil
}

func GetBalanceAt(clientUrl string ,addressHex string, height int64) (balance *big.Int, err error) {
	client, err := ethclient.Dial(clientUrl)
	if err != nil {
		return nil,err
	}

	account := common.HexToAddress(addressHex)
	blockNumber := big.NewInt(height)
	b, err := client.BalanceAt(context.Background(), account, blockNumber)
	if err != nil {
		return nil,err
	}
	return b,nil
}

func GetBlockHeader(clientUrl string ,height int64) (header *types.Header, err error){
	client, err := ethclient.Dial(clientUrl)
	if err != nil {
		return nil,err
	}

	blockNumber := big.NewInt(height)
	header, err = client.HeaderByNumber(context.Background(), blockNumber)
	if err != nil {
		return nil,err
	}

	return header,nil
}
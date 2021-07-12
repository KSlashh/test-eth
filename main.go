package main

import (
	"flag"
	"math/big"
	"strconv"

	"github.com/KSlashh/test-eth/api"
	"github.com/KSlashh/test-eth/config"
	"github.com/KSlashh/test-eth/log"
	"github.com/KSlashh/test-eth/testUtils"
	"github.com/ethereum/go-ethereum/ethclient"
)

var confFile string
var function string
var defaultInitEther int64 = 10^19

func init() {
	flag.StringVar(&confFile, "conf", "./config.json", "configuration file path")
	flag.StringVar(&function, "func", "test2", "choose function to run:\n" +
		"  getBalance [address]\n" +
		"  getBalanceAt [address] [height]\n" +
		"  getHeader [height]\n" +
		"  transferEther [toAddress] [amount/(wei)]\n" +
		"  testInfinite [instanceAmount] [initEther/(ether)(default 1)]\n" +
		"  testFixedRound [instanceAmount] [round] [initEther/(ether)(default 1)]\n" +
		"  testFixedTime [instanceAmount] [duration/(second)] [initEther/(ether)(default 1)]\n+" +
		"  test2 [instanceAmount(default 200)] [initEther/(ether)(default 10)]\n+" +
		"  record [startHeight]")

	flag.Parse()

}

func main() {
	conf, err := config.LoadConfig(confFile)
	if err != nil {
		log.Fatal("LoadConfig fail", err)
	}
	client, err := ethclient.Dial(conf.Node)
	if err != nil {
		log.Fatalf("Fail to dial client")
	}

	switch function {
	case "record":
		startHeight := big.NewInt(1)
		startHeight.SetString(flag.Arg(0), 10)
		testUtils.Recorder2(client, startHeight)
	case "test2":
		instanceAmount := 200
		initEther := big.NewInt(10^19)
		args := flag.Args()
		switch len(args) {
		case 1:
			instanceAmount,err = strconv.Atoi(flag.Arg(0))
			if err != nil {
				instanceAmount = 200
			}
		case 2:
			instanceAmount,err = strconv.Atoi(flag.Arg(0))
			if err != nil {
				instanceAmount = 200
			}
			initEther,ok := initEther.SetString(flag.Arg(1), 10)
			initEther.Mul(initEther, big.NewInt(10^18))
			if !ok {
				initEther.SetInt64(10^19)
			}
		default:
		}
		testUtils.TestServer2(instanceAmount, conf.Node, conf.PrivateKey, initEther)
	case "transferEther":
		amount := big.NewInt(0)
		amount,ok := amount.SetString(flag.Arg(1), 10)
		if !ok {
			log.Fatal("Fail to parse args! Second arg must be int.", err)
		}
		hash,err := api.TransferEth(client, conf.PrivateKey, flag.Arg(0), amount)
		if err != nil {
			log.Fatal("TransferEther fail", err)
		}
		log.Infof("Success! Transfer Ether at Tx %x .", hash)
	case "getBalance":
		address := flag.Arg(0)
		balance,err := api.GetBalance(client, address)
		if err != nil {
			log.Fatal("GetBalance fail", err)
		}
		log.Infof("balance of %s is %d", address, balance)
	case "getBalanceAt":
		address := flag.Arg(0)
		blockHeight,err := strconv.Atoi(flag.Arg(1))
		if err != nil {
			log.Fatal("Fail to parse args! Second arg must be int.", err)
		}
		balance,err := api.GetBalanceAt(client, address, int64(blockHeight))
		if err != nil {
			log.Fatal("GetBalance fail", err)
		}
		log.Infof("balance of %s at height %d is %d", address, blockHeight ,balance)
	case "getHeader":
		blockHeight,err := strconv.Atoi(flag.Arg(0))
		if err != nil {
			log.Fatal("Fail to parse args! First arg must be int.", err)
		}
		header,err := api.GetBlockHeader(client, int64(blockHeight))
		if err != nil {
			log.Fatal("GetHeader fail", err)
		}
		log.Info(header)
	case "testInfinite":
		instanceAmount,err := strconv.Atoi(flag.Arg(0))
		if err != nil {
			log.Fatal("Fail to parse args! First arg must be int.", err)
		}
		initEther := big.NewInt(10^18)
		amount := big.NewInt(0)
		amount,ok := amount.SetString(flag.Arg(1), 10)
		if ok {
			initEther = amount
		}
		testUtils.TestServer(instanceAmount, conf.Node, conf.PrivateKey, initEther, 0, 0)
	case "testFixedTime":
		instanceAmount,err := strconv.Atoi(flag.Arg(0))
		if err != nil {
			log.Fatal("Fail to parse args! First arg must be int.", err)
		}
		testDuration,err := strconv.Atoi(flag.Arg(1)) // second
		if err != nil {
			log.Fatal("Fail to parse args! First arg must be int.", err)
		}
		initEther := big.NewInt(10^18)
		amount := big.NewInt(0)
		amount,ok := amount.SetString(flag.Arg(2), 10)
		if ok {
			initEther = amount
		}
		testUtils.TestServer(instanceAmount, conf.Node, conf.PrivateKey, initEther, testDuration, 0)
	case "testFixedRound":
		instanceAmount,err := strconv.Atoi(flag.Arg(0))
		if err != nil {
			log.Fatal("Fail to parse args! First arg must be int.", err)
		}
		testRound,err := strconv.Atoi(flag.Arg(1))
		if err != nil {
			log.Fatal("Fail to parse args! First arg must be int.", err)
		}
		initEther := big.NewInt(10^18)
		amount := big.NewInt(0)
		amount,ok := amount.SetString(flag.Arg(2), 10)
		if ok {
			initEther = amount
		}
		testUtils.TestServer(instanceAmount, conf.Node, conf.PrivateKey, initEther,0, testRound)
	default :
		log.Fatal("unknown function", function)
	}
}

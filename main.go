package main

import (
	"flag"
	"strconv"

	"github.com/KSlashh/test-eth/api"
	"github.com/KSlashh/test-eth/config"
	"github.com/KSlashh/test-eth/log"
	"github.com/KSlashh/test-eth/testUtils"
)

var confFile string
var function string
var defaultInitEther int64 = 10^18

func init() {
	flag.StringVar(&confFile, "conf", "./config.json", "configuration file path")
	flag.StringVar(&function, "func", "getBalance", "choose function to run:\n" +
		"  getBalance [address]\n" +
		"  getBalanceAt [address] [height]\n" +
		"  getHeader [height]\n" +
		"  testInfinite [instanceAmount] [initEther(default 1e18)]\n" +
		"  testFixedRound [instanceAmount] [round] [initEther/(wei)(default 1e18)]\n" +
		"  testFixedTime [instanceAmount] [duration/(second)] [initEther/(wei)(default 1e18)]")

	flag.Parse()

}

func main() {
	conf, err := config.LoadConfig(confFile)
	if err != nil {
		log.Fatal("LoadConfig fail", err)
	}

	switch function {
	case "getBalance":
		address := flag.Arg(0)
		balance,err := api.GetBalance(conf.Node, address)
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
		balance,err := api.GetBalanceAt(conf.Node, address, int64(blockHeight))
		if err != nil {
			log.Fatal("GetBalance fail", err)
		}
		log.Infof("balance of %s at height %d is %d", address, blockHeight ,balance)
	case "getHeader":
		blockHeight,err := strconv.Atoi(flag.Arg(0))
		if err != nil {
			log.Fatal("Fail to parse args! First arg must be int.", err)
		}
		header,err := api.GetBlockHeader(conf.Node, int64(blockHeight))
		if err != nil {
			log.Fatal("GetHeader fail", err)
		}
		log.Info(header)
	case "testInfinite":
		instanceAmount,err := strconv.Atoi(flag.Arg(0))
		if err != nil {
			log.Fatal("Fail to parse args! First arg must be int.", err)
		}
		var initEther int64 = defaultInitEther
		arg2,_ := strconv.ParseInt(flag.Arg(1), 10, 64)
		if arg2>0 {
			initEther = arg2
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
		var initEther int64 = defaultInitEther
		arg3,_ := strconv.ParseInt(flag.Arg(2), 10, 64)
		if arg3>0 {
			initEther = arg3
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
		var initEther int64 = defaultInitEther
		arg3,_ := strconv.ParseInt(flag.Arg(2), 10, 64)
		if arg3>0 {
			initEther = arg3
		}
		testUtils.TestServer(instanceAmount, conf.Node, conf.PrivateKey, initEther,0, testRound)
	default :
		log.Fatal("unknown function", function)
	}
}

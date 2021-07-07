package main

import (
	"flag"
	"github.com/test-eth/api"
	"github.com/test-eth/config"
	"github.com/test-eth/log"
	"strconv"
)

var confFile string
var function string

func init() {
	flag.StringVar(&confFile, "conf", "./config.json", "configuration file path")
	flag.StringVar(&function, "func", "deploy", "choose function to run: deploy or run")

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
	case "transferEther":
        amount,err := strconv.Atoi(flag.Arg(1))
		if err != nil {
			log.Fatal("Fail to parse args! Second arg must be int.", err)
		}
		hash,err := api.TransferEth(conf.Node, conf.PrivateKey, flag.Arg(0), int64(amount))
		if err != nil {
			log.Fatal("TransferEther fail", err)
		}
		log.Infof("Success!Tx %s is pending...", hash)
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
	}
}
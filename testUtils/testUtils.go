package testUtils

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/KSlashh/test-eth/api"
	"github.com/KSlashh/test-eth/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type instanceMsg struct {
	msgType int     // 1:successTx , 2:failedTx , 3:instanceShutDown , 4:instanceStart
	timeCost int     // millisecond
}

var recordFrequency float64 = 5    // second
var totalDataRecordFrequency float64 = 20      // second
var checkTxComfirmFrequency = time.Second * 1

func TestServer(numOfInstance int, clientUrl string, privateKeyhex string, initEther int64 , duration int, round int) {
	msgChan := make(chan instanceMsg, 10000*numOfInstance)
	if round > 0 {
		for i:=0;i<numOfInstance;i++ {
			go FixedRoundTestInstance(clientUrl, privateKeyhex, i, initEther, round, msgChan)
		}
	} else if duration >0 {
		for i:=0;i<numOfInstance;i++ {
			go FixedTimeTestInstance(clientUrl, privateKeyhex, i, initEther, duration, msgChan)
		}
	} else {
		for i:=0;i<numOfInstance;i++ {
			go InfiniteTestInstance(clientUrl, privateKeyhex, i, initEther, msgChan)
		}
	}
	client,_ := ethclient.Dial(clientUrl)
	header,_ := client.HeaderByNumber(context.Background(), nil)
	startHeight := header.Number
	log.Infof("Start test. Start at block %s.", startHeight.String())
	Recorder(msgChan)
	header,_ = client.HeaderByNumber(context.Background(), nil)
	endHeight := header.Number
	log.Infof("Done test. Started at block %s, end at block %s.", startHeight.String(), endHeight.String())
}

func Recorder(msgs chan instanceMsg) {
	goodTxTmp := big.NewInt(0)
	badTxTmp := big.NewInt(0)
	totalCostTmp := big.NewInt(0)
	var liveInstance, deadInsatance int
	goodTx := big.NewInt(0)
	badTx := big.NewInt(0)
	totalCost := big.NewInt(0)
	averageCost := big.NewInt(0)
	one := big.NewInt(1)
	start := time.Now()
	timeCache := time.Now()
	timeCache2 := time.Now()
	for msg := range msgs {
		switch msg.msgType {
		case 1:
			goodTx.Add(goodTx, one)
			totalCost.Add(totalCost, big.NewInt(int64(msg.timeCost)))
			goodTxTmp.Add(goodTxTmp, one)
			totalCostTmp.Add(totalCostTmp, big.NewInt(int64(msg.timeCost)))
		case 2:
			badTx.Add(badTx, one)
			badTxTmp.Add(badTxTmp, one)
		case 3:
			liveInstance -= 1
			deadInsatance += 1
			if liveInstance == 0 {
				log.Infof("——————————ToTal data: " +
					"Start-time: %s, " +
					"Duration: %f s, " +
					"Succeed-Txns: %d, " +
					"Failed-Txns: %d, " +
					"Running-Instance: %d, " +
					"Dead-Instance: %d, " +
					"Average-Comfirm-timeCost: %s ms, " +
					"Tps: %f",
					start.Format("2006-01-02_15:04:05"),
					time.Since(start).Seconds(),
					goodTx,
					badTx,
					liveInstance,
					deadInsatance,
					averageCost.Div(totalCost, goodTx),
					float64(goodTx.Int64())/(time.Since(start).Seconds()),
				)
				return
			}
		case 4:
			liveInstance += 1
		default:
		}
		if goodTx.Int64() == 0 {continue}
		if time.Since(timeCache2).Seconds() >= totalDataRecordFrequency {
			log.Infof("——————————ToTal data: " +
				"Start-time: %s, " +
				"Duration: %f s, " +
				"Succeed-Txns: %d, " +
				"Failed-Txns: %d, " +
				"Running-Instance: %d, " +
				"Dead-Instance: %d, " +
				"Average-Comfirm-timeCost: %s ms, " +
				"Tps: %f",
				start.Format("2006-01-02_15:04:05"),
				time.Since(start).Seconds(),
				goodTx,
				badTx,
				liveInstance,
				deadInsatance,
				averageCost.Div(totalCost, goodTx),
				float64(goodTx.Int64())/(time.Since(start).Seconds()),
			)
			timeCache2 = time.Now()
		}
		if goodTxTmp.Int64() == 0 {continue}
		if time.Since(timeCache).Seconds() >= recordFrequency {
			log.Infof("Data since last record: " +
				"Last-record-time: %s, " +
				"Duration: %f s, " +
				"Succeed-Txns: %d, " +
				"Failed-Txns: %d, " +
				"Running-Instance: %d, " +
				"Dead-Instance: %d, " +
				"Average-Comfirm-timeCost: %s ms, " +
				"Tps: %f",
				timeCache.Format("2006-01-02_15:04:05"),
				time.Since(timeCache).Seconds(),
				goodTxTmp,
				badTxTmp,
				liveInstance,
				deadInsatance,
				averageCost.Div(totalCostTmp, goodTxTmp),
				float64(goodTxTmp.Int64())/(time.Since(timeCache).Seconds()),
				)
			goodTxTmp = big.NewInt(0)
			badTxTmp = big.NewInt(0)
			totalCostTmp = big.NewInt(0)
			timeCache = time.Now()
		}
	}
}

func InfiniteTestInstance(clientUrl string, mainPrivateKeyHex string, index int, initEther int64, ch chan instanceMsg) {
	started := false
	client, err := ethclient.Dial(clientUrl)
	if err != nil {
		log.Fatalf("Instance %d fail to dial client", index)
	}

	// generate 2 accounts
	privateKeyA, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyB, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	skA := hexutil.Encode(crypto.FromECDSA(privateKeyA))[2:]
	pkA := crypto.PubkeyToAddress(*privateKeyA.Public().(*ecdsa.PublicKey)).Hex()
	skB := hexutil.Encode(crypto.FromECDSA(privateKeyB))[2:]
	pkB := crypto.PubkeyToAddress(*privateKeyB.Public().(*ecdsa.PublicKey)).Hex()

	// admin-->initEther-->A
	for {
		hash, err := api.TransferEth(clientUrl, mainPrivateKeyHex, pkA, initEther)
		if err != nil {
			continue
		}
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if isSuccess {
			break
		}
	}

	// admin-->initEther-->B
	for {
		hash, err := api.TransferEth(clientUrl, mainPrivateKeyHex, pkB, initEther)
		if err != nil {
			continue
		}
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if isSuccess {
			break
		}
	}

	ch<- instanceMsg{4,0}
	started = true
	defer func() {
		if started {
			ch<- instanceMsg{3,0}
		}
	}()
	timeCache := time.Now()

	for {
		// A-->10000wei-->B
		hash,err := api.TransferEth(clientUrl, skA, pkB, 10000)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch<- instanceMsg{2,0}
		} else {
			ch<- instanceMsg{1,int(time.Since(timeCache).Milliseconds())}
		}

		// B-->10000wei-->A
		hash,err = api.TransferEth(clientUrl, skB, pkA, 10000)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess = WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch<- instanceMsg{2,0}
		} else {
			ch<- instanceMsg{1,int(time.Since(timeCache).Milliseconds())}
		}
	}
}

func FixedTimeTestInstance(clientUrl string, mainPrivateKeyHex string, index int, initEther int64, duration int, ch chan instanceMsg) {
	started := false
	client, err := ethclient.Dial(clientUrl)
	if err != nil {
		log.Fatalf("Instance %d fail to dial client", index)
	}

	// generate 2 accounts
	privateKeyA, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyB, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	skA := hexutil.Encode(crypto.FromECDSA(privateKeyA))[2:]
	pkA := crypto.PubkeyToAddress(*privateKeyA.Public().(*ecdsa.PublicKey)).Hex()
	skB := hexutil.Encode(crypto.FromECDSA(privateKeyB))[2:]
	pkB := crypto.PubkeyToAddress(*privateKeyB.Public().(*ecdsa.PublicKey)).Hex()

	// admin-->initEther-->A
	for {
		hash, err := api.TransferEth(clientUrl, mainPrivateKeyHex, pkA, initEther)
		if err != nil {
			continue
		}
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if isSuccess {
			break
		}
	}

	// admin-->initEther-->B
	for {
		hash, err := api.TransferEth(clientUrl, mainPrivateKeyHex, pkB, initEther)
		if err != nil {
			continue
		}
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if isSuccess {
			break
		}
	}

	ch<- instanceMsg{4,0}
	started = true
	defer func() {
		if started {
			ch<- instanceMsg{3,0}
		}
	}()
	timeCache := time.Now()
	timeStart := time.Now()

	for time.Since(timeStart).Milliseconds() >= int64(duration)*1000 {
		// A-->10000wei-->B
		hash,err := api.TransferEth(clientUrl, skA, pkB, 10000)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch<- instanceMsg{2,0}
		} else {
			ch<- instanceMsg{1,int(time.Since(timeCache).Milliseconds())}
		}

		// B-->10000wei-->A
		hash,err = api.TransferEth(clientUrl, skB, pkA, 10000)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess = WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch<- instanceMsg{2,0}
		} else {
			ch<- instanceMsg{1,int(time.Since(timeCache).Milliseconds())}
		}
	}
}

func FixedRoundTestInstance(clientUrl string, mainPrivateKeyHex string, index int, initEther int64, round int, ch chan instanceMsg) {
	started := false
	client, err := ethclient.Dial(clientUrl)
	if err != nil {
		log.Fatalf("Instance %d fail to dial client", index)
	}

	// generate 2 accounts
	privateKeyA, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyB, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	skA := hexutil.Encode(crypto.FromECDSA(privateKeyA))[2:]
	pkA := crypto.PubkeyToAddress(*privateKeyA.Public().(*ecdsa.PublicKey)).Hex()
	skB := hexutil.Encode(crypto.FromECDSA(privateKeyB))[2:]
	pkB := crypto.PubkeyToAddress(*privateKeyB.Public().(*ecdsa.PublicKey)).Hex()

	// admin-->initEther-->A
	for {
		hash, err := api.TransferEth(clientUrl, mainPrivateKeyHex, pkA, initEther)
		if err != nil {
			continue
		}
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if isSuccess {
			break
		}
	}

	// admin-->initEther-->B
	for {
		hash, err := api.TransferEth(clientUrl, mainPrivateKeyHex, pkB, initEther)
		if err != nil {
			continue
		}
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if isSuccess {
			break
		}
	}

	ch<- instanceMsg{4,0}
	started = true
	defer func() {
		if started {
			ch<- instanceMsg{3,0}
		}
	}()
	timeCache := time.Now()

	for i:=0;i<round;i++ {
		// A-->10000wei-->B
		hash,err := api.TransferEth(clientUrl, skA, pkB, 10000)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch<- instanceMsg{2,0}
		} else {
			ch<- instanceMsg{1,int(time.Since(timeCache).Milliseconds())}
		}

		// B-->10000wei-->A
		hash,err = api.TransferEth(clientUrl, skB, pkA, 10000)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess = WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch<- instanceMsg{2,0}
		} else {
			ch<- instanceMsg{1,int(time.Since(timeCache).Milliseconds())}
		}
	}
}

func WaitTransactionConfirm(client *ethclient.Client,hash []byte) bool {
	for {
		time.Sleep(checkTxComfirmFrequency)
		_, isPending, err := client.TransactionByHash(context.Background(), common.BytesToHash(hash))
		if err != nil {
			continue
		}
		if isPending == true {
			continue
		} else {
			receipt, err := client.TransactionReceipt(context.Background(), common.BytesToHash(hash))
			if err != nil {
				continue
			}
			return receipt.Status == types.ReceiptStatusSuccessful
		}
	}
}

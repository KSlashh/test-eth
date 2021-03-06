package testUtils

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"
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
	msgType  int // 1:successTx , 2:failedTx , 3:instanceShutDown , 4:instanceStart32G
	timeCost int // millisecond
}

var recordFrequency float64 = 5           // second
var totalDataRecordFrequency float64 = 20 // second
var checkTxComfirmFrequency = time.Second * 1
var instanceTransferFrequency = time.Second * 1
var smapleTxnAmount = big.NewInt(10000)
var txnsPerPack = 10
var m *sync.Mutex

func TestServer2(numOfInstance int, clientUrl string, privateKeyHex string, initEther *big.Int) {
	client, err := ethclient.Dial(clientUrl)
	if err != nil {
		log.Fatal(err)
	}
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	startHeight := header.Number
	log.Infof("Start testing at height %s", startHeight.String())
	m = new(sync.Mutex)
	for i := 0; i < numOfInstance; i++ {
		Instance2(clientUrl, privateKeyHex, initEther)
	}
	// Recorder2(client, startHeight)
}

func Recorder2(client *ethclient.Client, startHeight *big.Int) {
	header, err := client.HeaderByNumber(context.Background(), startHeight)
	if err != nil {
		log.Fatal(err)
	}
	timeStamp := header.Time
	height := startHeight
	log.Infof("Start recording at height %s", height.String())
	totalTxns := big.NewInt(0)
	totalTime := big.NewInt(0)
	duration := big.NewInt(0)
	one := big.NewInt(1)
	tmp := big.NewInt(1)
	height.Add(height, one)
	for {
		header, err = client.HeaderByNumber(context.Background(), height)
		if err != nil {
			time.Sleep(time.Second * 1)
			continue
		}
		count, err := client.TransactionCount(context.Background(), header.Hash())
		if err != nil {
			log.Info(err)
			continue
		}
		height.Add(height, one)
		time := header.Time
		totalTxns.Add(totalTxns, tmp.SetUint64(uint64(count)))
		duration.SetUint64(time - timeStamp)
		totalTime.Add(totalTime, duration)
		timeStamp = time
		if count == 0 {
			log.Infof("skip empty block %d", height.Int64()-1)
			continue
		}
		log.Infof(""+
			"Now height at %s : ,"+
			"last block duration: %s s,"+
			"this txns: %d ,"+
			"total txns: %s ,"+
			"this tps: %f ,"+
			"total tps: %s ",
			height.String(),
			duration.String(),
			count,
			totalTxns.String(),
			float64(count)/float64(duration.Int64()),
			tmp.Div(totalTxns, totalTime).String(),
		)
	}
}

func Instance2(clientUrl string, mainPrivateKeyHex string, initEther *big.Int) {
	client, err := ethclient.Dial(clientUrl)
	if err != nil {
		log.Fatalf("Instance fail to dial client")
	}
    client2,_ := ethclient.Dial(clientUrl)
	// generate 2 accounts
	privateKeyA, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyB, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	// skA := hexutil.Encode(crypto.FromECDSA(privateKeyA))[2:]
	pkA := crypto.PubkeyToAddress(*privateKeyA.Public().(*ecdsa.PublicKey))
	// skB := hexutil.Encode(crypto.FromECDSA(privateKeyB))[2:]
	pkB := crypto.PubkeyToAddress(*privateKeyB.Public().(*ecdsa.PublicKey))

	// admin-->initEther-->A


	// admin-->initEther-->B
	pkabalance, pkbbalance := new(big.Int), new(big.Int)
	api.TransferEth(client, mainPrivateKeyHex, pkA.Hex(), initEther)
	for {
		balance, _ := client.BalanceAt(context.Background(), pkA, nil)
		time.Sleep(1 * time.Second)
		if balance.Cmp(initEther) >= 0 {
			pkabalance = balance
			break
		}
	}
	api.TransferEth(client, mainPrivateKeyHex, pkB.Hex(), initEther)
	for {
		balance, _ := client.BalanceAt(context.Background(), pkB, nil)
		time.Sleep(1 * time.Second)
		if balance.Cmp(initEther) >= 0 {
			pkbbalance = balance
			break
		}
	}

	log.Infof("pka %s balance %s", pkA.Hex(), pkabalance.String())
	log.Infof("pkb %s balance %s", pkB.Hex(), pkbbalance.String())

	nonceA, err := client.NonceAt(context.Background(), pkA, nil)
	nonceB, err := client.NonceAt(context.Background(), pkB, nil)
	chainID := big.NewInt(102)
	signerA := types.NewEIP155Signer(chainID)
	signerB := types.NewEIP155Signer(chainID)
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000)
	ch1 := make(chan *types.Transaction, 1000)
	ch2 := make(chan *types.Transaction, 1000)

	go func() {
		for  {
			signedTx,err := types.SignTx(
				types.NewTransaction(nonceA, pkA, smapleTxnAmount, gasLimit, gasPrice, nil),
				signerA,
				privateKeyA)
			if err != nil {
				nonceA,_ = client.PendingNonceAt(context.Background(), pkA)
			} else {
				nonceA += 1
				if nonceA%2 == 1 {
					ch1<-signedTx
				} else {
					ch2<-signedTx
				}
			}
		}
	}()

	go func() {
		for  {
			signedTx,err := types.SignTx(
				types.NewTransaction(nonceB, pkB, smapleTxnAmount, gasLimit, gasPrice, nil),
				signerB,
				privateKeyB)
			if err != nil {
				nonceB,_ = client.PendingNonceAt(context.Background(), pkB)
			} else {
				nonceB += 1
				if nonceB%2 == 1 {
					ch1<-signedTx
				} else {
					ch2<-signedTx
				}
			}
		}
	}()

	go func() {
		for  {

		}
	}()

	for {
		select{
		case tx := <-ch1:
			client.SendTransaction(context.Background(), tx)
		case tx := <- ch2:
			client2.SendTransaction(context.Background(), tx)
		}
	}
}

func sendETH(client *ethclient.Client, privateKey *ecdsa.PrivateKey, nonce uint64, toAddress common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int) error {
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, data)

	//chainID, err := client.ChainID(context.Background())
	//if err != nil {
	//	return err
	//}
	chainID := big.NewInt(102)
	signer := types.NewEIP155Signer(chainID)

	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}

	return nil
}

func TestServer(numOfInstance int, clientUrl string, privateKeyhex string, initEther *big.Int, duration int, round int) {
	msgChan := make(chan instanceMsg, 10000*numOfInstance)
	if round > 0 {
		for i := 0; i < numOfInstance; i++ {
			go FixedRoundTestInstance(clientUrl, privateKeyhex, i, initEther, round, msgChan)
		}
	} else if duration > 0 {
		for i := 0; i < numOfInstance; i++ {
			go FixedTimeTestInstance(clientUrl, privateKeyhex, i, initEther, duration, msgChan)
		}
	} else {
		for i := 0; i < numOfInstance; i++ {
			go InfiniteTestInstance(clientUrl, privateKeyhex, i, initEther, msgChan)
		}
	}
	client, _ := ethclient.Dial(clientUrl)
	header, _ := client.HeaderByNumber(context.Background(), nil)
	startHeight := header.Number
	log.Infof("Start test. Start at block %s.", startHeight.String())
	Recorder(msgChan)
	header, _ = client.HeaderByNumber(context.Background(), nil)
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
				log.Infof("??????????????????????????????ToTal data: "+
					"Start-time: %s, "+
					"Duration: %f s, "+
					"Succeed-Txns: %d, "+
					"Failed-Txns: %d, "+
					"Running-Instance: %d, "+
					"Dead-Instance: %d, "+
					"Average-Comfirm-timeCost: %s ms, "+
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
		if goodTx.Int64() == 0 {
			continue
		}
		if time.Since(timeCache2).Seconds() >= totalDataRecordFrequency {
			log.Infof("??????????????????????????????ToTal data: "+
				"Start-time: %s, "+
				"Duration: %f s, "+
				"Succeed-Txns: %d, "+
				"Failed-Txns: %d, "+
				"Running-Instance: %d, "+
				"Dead-Instance: %d, "+
				"Average-Comfirm-timeCost: %s ms, "+
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
		if goodTxTmp.Int64() == 0 {
			continue
		}
		if time.Since(timeCache).Seconds() >= recordFrequency {
			log.Infof("Data since last record: "+
				"Last-record-time: %s, "+
				"Duration: %f s, "+
				"Succeed-Txns: %d, "+
				"Failed-Txns: %d, "+
				"Running-Instance: %d, "+
				"Dead-Instance: %d, "+
				"Average-Comfirm-timeCost: %s ms, "+
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

func InfiniteTestInstance(clientUrl string, mainPrivateKeyHex string, index int, initEther *big.Int, ch chan instanceMsg) {
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
		hash, err := api.TransferEth(client, mainPrivateKeyHex, pkA, initEther)
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
		hash, err := api.TransferEth(client, mainPrivateKeyHex, pkB, initEther)
		if err != nil {
			continue
		}
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if isSuccess {
			break
		}
	}

	ch <- instanceMsg{4, 0}
	started = true
	defer func() {
		if started {
			ch <- instanceMsg{3, 0}
		}
	}()
	timeCache := time.Now()

	for {
		// A-->10000wei-->B
		hash, err := api.TransferEth(client, skA, pkB, smapleTxnAmount)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch <- instanceMsg{2, 0}
		} else {
			ch <- instanceMsg{1, int(time.Since(timeCache).Milliseconds())}
		}

		// B-->10000wei-->A
		hash, err = api.TransferEth(client, skB, pkA, smapleTxnAmount)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess = WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch <- instanceMsg{2, 0}
		} else {
			ch <- instanceMsg{1, int(time.Since(timeCache).Milliseconds())}
		}
	}
}

func FixedTimeTestInstance(clientUrl string, mainPrivateKeyHex string, index int, initEther *big.Int, duration int, ch chan instanceMsg) {
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
		hash, err := api.TransferEth(client, mainPrivateKeyHex, pkA, initEther)
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
		hash, err := api.TransferEth(client, mainPrivateKeyHex, pkB, initEther)
		if err != nil {
			continue
		}
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if isSuccess {
			break
		}
	}

	ch <- instanceMsg{4, 0}
	started = true
	defer func() {
		if started {
			ch <- instanceMsg{3, 0}
		}
	}()
	timeCache := time.Now()
	timeStart := time.Now()

	for time.Since(timeStart).Milliseconds() >= int64(duration)*1000 {
		// A-->10000wei-->B
		hash, err := api.TransferEth(client, skA, pkB, smapleTxnAmount)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch <- instanceMsg{2, 0}
		} else {
			ch <- instanceMsg{1, int(time.Since(timeCache).Milliseconds())}
		}

		// B-->10000wei-->A
		hash, err = api.TransferEth(client, skB, pkA, smapleTxnAmount)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess = WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch <- instanceMsg{2, 0}
		} else {
			ch <- instanceMsg{1, int(time.Since(timeCache).Milliseconds())}
		}
	}
}

func FixedRoundTestInstance(clientUrl string, mainPrivateKeyHex string, index int, initEther *big.Int, round int, ch chan instanceMsg) {
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
		hash, err := api.TransferEth(client, mainPrivateKeyHex, pkA, initEther)
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
		hash, err := api.TransferEth(client, mainPrivateKeyHex, pkB, initEther)
		if err != nil {
			continue
		}
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if isSuccess {
			break
		}
	}

	ch <- instanceMsg{4, 0}
	started = true
	defer func() {
		if started {
			ch <- instanceMsg{3, 0}
		}
	}()
	timeCache := time.Now()

	for i := 0; i < round; i++ {
		// A-->10000wei-->B
		hash, err := api.TransferEth(client, skA, pkB, smapleTxnAmount)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess := WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch <- instanceMsg{2, 0}
		} else {
			ch <- instanceMsg{1, int(time.Since(timeCache).Milliseconds())}
		}

		// B-->10000wei-->A
		hash, err = api.TransferEth(client, skB, pkA, smapleTxnAmount)
		if err != nil {
			continue
		}
		timeCache = time.Now()
		isSuccess = WaitTransactionConfirm(client, hash[:])
		if !isSuccess {
			ch <- instanceMsg{2, 0}
		} else {
			ch <- instanceMsg{1, int(time.Since(timeCache).Milliseconds())}
		}
	}
}

func WaitTransactionConfirm(client *ethclient.Client, hash []byte) bool {
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

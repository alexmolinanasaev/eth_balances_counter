package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

const BLOCKS_STEP = 100
const REQUESTS_DELAY = 40 * time.Millisecond // 25 запросов в секунду

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Info("Starting application")
	apiKey := os.Getenv("API_KEY")
	nodeURL := os.Getenv("NODE_URL")
	logrus.Infof("\nAPI_KEY=%s\nNODE_URL=%s\n", apiKey, nodeURL)

	client, err := ethclient.Dial(fmt.Sprintf("%s?api_key=%s", nodeURL, apiKey))
	if err != nil {
		panic(err)
	}

	logrus.Info("Querying latest block number")
	latestBlockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}

	balances := make(map[common.Address]*big.Int)
	wg := &sync.WaitGroup{}
	mux := sync.RWMutex{}

	ticker := time.Ticker{}

	startBlockNumber := latestBlockNumber - BLOCKS_STEP
	for i := startBlockNumber; i <= latestBlockNumber; i++ {
		logrus.Infof("Querying block number: %d", i)
		block, err := client.BlockByNumber(context.Background(), big.NewInt(0).SetUint64(i))
		if err != nil {
			logrus.Error("cannot request block: %s", err)
			continue
		}

		for _, transaction := range block.Transactions() {
			wg.Add(1)
			go func(tx *types.Transaction, blockNumber uint64) {
				defer wg.Done()

				var fromAddress common.Address
				msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), tx.GasPrice())
				if err != nil {
					logrus.Errorf("cannot get sender at tx[%s]", tx.Hash())
					return
				}
				fromAddress = msg.From()

				receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
				if err != nil {
					log.Fatal(err)
				}

				gasUsed := big.NewInt(0).SetUint64(receipt.GasUsed)

				mux.Lock()
				fromBalance := big.NewInt(0)
				if _, ok := balances[fromAddress]; ok {
					fromBalance = balances[fromAddress]
				}
				fromBalance = big.NewInt(0).Sub(fromBalance, big.NewInt(0).Add(gasUsed, tx.Value()))
				balances[fromAddress] = fromBalance

				to := tx.To()
				// при создании контракта получателя может не быть
				if to == nil {
					logrus.Infof("contract creation at tx[%s]", tx.Hash())
					to = &common.Address{}
				}
				toBalance := big.NewInt(0)
				if _, ok := balances[*to]; ok {
					toBalance = balances[*to]
				}
				toBalance = big.NewInt(0).Add(toBalance, tx.Value())
				balances[*to] = toBalance
				mux.Unlock()

				logrus.Infof("\nBlock:  %d\nTxHash: %s\nFrom:   %s\nTo:     %s\nCost:   %s\n",
					blockNumber, tx.Hash(), fromAddress, to.String(), tx.Cost().String())

			}(transaction, i)

			<-ticker.C
		}
	}
	wg.Wait()
	ticker.Stop()

	biggestBalanceDelta := big.NewInt(0)
	var addressWithBiggestBalanceDelta common.Address

	for address, balance := range balances {
		if balance.CmpAbs(biggestBalanceDelta) == 1 {
			biggestBalanceDelta = balance
			addressWithBiggestBalanceDelta = address
		}
	}

	fmt.Println("")
	logrus.Infof("Address with biggest balance delta: %s\n", addressWithBiggestBalanceDelta.Hex())
	logrus.Infof("Biggest balance delta [wei]: %s\n", biggestBalanceDelta.String())
}

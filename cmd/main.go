package main

import (
	"btcgo/internal/core"
	"btcgo/internal/domain"
	"btcgo/internal/utils"
	"log"
	"math/big"
	"sync"
)

func main() {
	ranges, wallets := utils.LoadData()
	params := utils.GetParameters(*wallets)
	run(params, ranges, wallets)
}

func run(params domain.Parameters, ranges *domain.Ranges, wallets *domain.Wallets) {
	inputChannel := make(chan *big.Int, params.WorkerCount*2)
	outputChannel := make(chan *big.Int, params.WorkerCount)
	var workerGroup, outputGroup sync.WaitGroup

	workerGroup.Add(1)
	outputGroup.Add(1)
	go core.WorkersStartUp(params, wallets, inputChannel, outputChannel, &workerGroup)
	go core.OutputHandler(outputChannel, wallets, params, &outputGroup)

	for i := 0; i < params.BatchCount || params.BatchCount == -1; i++ {
		batchCounter := i + 1
		start, end := getStartAndEnd(ranges, params)
		startClone := utils.Clone(start)

		if params.Rng {
			start, _ = utils.GenerateRandomNumber(start, end)
		} else if params.BatchSize != -1 {
			startAdd := new(big.Int).Mul(
				big.NewInt(params.BatchSize),
				new(big.Int).Sub(big.NewInt(int64(batchCounter)), big.NewInt(1)))
			start = new(big.Int).Add(start, startAdd)
		}

		if start.Cmp(end) > 0 {
			break
		}

		if params.VerboseSummary {
			if batchCounter <= 1 {
				core.PrintSummary(startClone, utils.Clone(end), utils.Clone(start), params, batchCounter)
			} else {
				core.PrintTinySummary(startClone, utils.Clone(end), utils.Clone(start), params, batchCounter)
			}
		}

		core.Scheduler(start, end, params, inputChannel)
	}

	close(inputChannel)
	workerGroup.Wait()
	close(outputChannel)
	outputGroup.Wait()
}

func getStartAndEnd(ranges *domain.Ranges, params domain.Parameters) (*big.Int, *big.Int) {
	start, ok := new(big.Int).SetString(ranges.Ranges[params.TargetWallet].Min[2:], 16)
	if !ok {
		log.Fatal("Erro ao converter o valor de início")
	}
	end, ok := new(big.Int).SetString(ranges.Ranges[params.TargetWallet].Max[2:], 16)
	if !ok {
		log.Fatal("Erro ao converter o valor de fim")
	}
	return start, end
}

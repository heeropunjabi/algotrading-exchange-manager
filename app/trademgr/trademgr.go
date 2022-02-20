// trademgr - executes and manages trades.
// Read strategies from db, spwans threads for each strategy.
// Remains active till the trade is closed
package trademgr

import (
	"fmt"
	"goTicker/app/data"
	"goTicker/app/db"
	"goTicker/app/srv"
	"os"
	"strings"
	"sync"
	"time"
)

// tradeStrategies - list of all strategies to be executed. Read once from db at start of day
var (
	tradeStrategies        []*data.Strategies
	terminateTradeOperator bool = false
)

// Scan DB for all strategies with strategy_en = 1. Each funtion is executed in a separate thread and remains active till the trade is complete.
// TODO: recovery logic for server restarts
func Trader() {

	var wgTrademgr sync.WaitGroup
	terminateTradeOperator = false

	srv.InitTradeLogger()

	// 1. Read trading strategies from dB
	tradeStrategies = db.ReadStrategiesFromDb()

	// 2. Setup time intervals for each strategy (loop for each)
	for each := range tradeStrategies {
		wgTrademgr.Add(1)
		go tradeOperator(tradeStrategies[each], &wgTrademgr)
	}

	// 3. wait till all trades are completed
	wgTrademgr.Wait()
	os.Exit(0)

}

func StopTrader() {
	terminateTradeOperator = true
}

func tradeOperator(tradeStrategies *data.Strategies, wg *sync.WaitGroup) {
	defer wg.Done()

	srv.TradesLogger.Println("\n\ntradeOperator ", tradeStrategies)

	if checkTriggerDays(tradeStrategies) {
		// 1. wait till trigger time
		awaitTriggerTime(tradeStrategies)

		// 2. execute signal api

		// 3. on signal, execute trade

		// 4. on trade completion, update db

		// 5. montitor trade positions

		// 6. check exit conditions

		// 7. on signal, exit trade

		// 8. on exit, update db
	}
}

// Check if the current day is a trading day. Valid syntax "Monday,Tuesday,Wednesday,Thursday,Friday". For day selection to trade - Every day must be explicitly listed in dB.
func checkTriggerDays(tradeStrategies *data.Strategies) bool {

	triggerdays := strings.Split(tradeStrategies.P_trigger_days, ",")
	currentday := time.Now().Weekday().String()

	for each := range triggerdays {
		if triggerdays[each] == currentday {
			srv.TradesLogger.Println(tradeStrategies.Strategy_id, " : Trade signal registered")
			return true
		}
	}
	srv.TradesLogger.Println(tradeStrategies.Strategy_id, " : Trade signal skipped due to no valid day trigger present")
	return false
}

// Wait till the current time is greater than the trigger time.
// TODO: master exit condition & EoD termniation
func awaitTriggerTime(tradeStrategies *data.Strategies) {

	// const layoutTime = "15:04:05"
	for {
		curTime := time.Now()
		triggerTime := tradeStrategies.P_trigger_time
		// fmt.Println(triggerTime, " : ", curTime)

		if curTime.Hour() == triggerTime.Hour() {
			if curTime.Minute() == triggerTime.Minute() {
				fmt.Println("intiating trade check")
				return
			}
		}

		// termination requested
		if terminateTradeOperator {
			return
		}

		time.Sleep(1 * time.Second * 10)
		fmt.Println("sleeping ", tradeStrategies.Strategy_id)
	}
}

package kite

import (
	"goTicker/app/data"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

func PlaceOrder(order *data.TradeSignal) bool {

	if order.Instr != "" {
		return true
	} else {
		return false
	}
}

func CalOrderMargin(order data.TradeSignal, ts data.Strategies) bool {

	var marginParam kiteconnect.GetMarginParams

	// default params
	marginParam.Compact = false
	marginParam.OrderParams[0].Exchange = "NSE"
	marginParam.OrderParams[0].OrderType = "MARKET"
	marginParam.OrderParams[0].Quantity = 1
	marginParam.OrderParams[0].Price = 0
	marginParam.OrderParams[0].TriggerPrice = 0
	// specific params
	marginParam.OrderParams[0].Variety = ts.CtrlParam.KiteSettings.Varieties
	marginParam.OrderParams[0].Product = ts.CtrlParam.KiteSettings.Products
	if strings.ToLower(order.Dir) == "bullish" {
		marginParam.OrderParams[0].TransactionType = "BUY"
	} else {
		marginParam.OrderParams[0].TransactionType = "SELL"
	}

	switch ts.CtrlParam.TradeSettings.OrderRoute {

	default:
		fallthrough
	case "stock":
		marginParam.OrderParams[0].Tradingsymbol = order.Instr

	case "option":
		marginParam.OrderParams[0].Tradingsymbol = deriveOptionName(order, ts, time.Now())

	case "futures":
		marginParam.OrderParams[0].Tradingsymbol = deriveFuturesName(order, ts, time.Now())

	}
	OrderMargins, err := kc.GetOrderMargins(marginParam)

	print(OrderMargins, err)
	return true

}

// The format is BANKNIFTY<YY><M><DD>strike<PE/CE>
// The month format is 1 for JAN, 2 for FEB, 3, 4, 5, 6, 7, 8, 9, O(capital o) for October, N for November, D for December.
// var symbolFutStr string = "FAILED"
// BANKNIFTY2232435000CE - 24th Mar 2022
// BANKNIFTY22MAR31000CE - 31st Mar 2022
// Last week of Month - will be monthly expiry
func deriveOptionName(order data.TradeSignal, ts data.Strategies, selDate time.Time) string {

	var (
		lvl             float64
		mth             string
		strikePriceStep int
		optn            string
	)

	// ---------------------------------------------------------------------- READ STRIKE PRICE LEVELS FOR INSTRUMENTS
	f, err := os.Open("./../zfiles/config/OptionsStrikePriceSteps.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	envMap, err := godotenv.Parse(f) // load inst, strikeprice into a map
	if err != nil {
		panic(err)
	}

	// ---------------------------------------------------------------------- HOUSEKEEPING JOBS
	Instr := strings.ReplaceAll(order.Instr, "-FUT", "") // remove -FUT suffix
	wkday := selDate.Weekday()
	currThu := time.Now() // dummy initialisation

	// ---------------------------------------------------------------------- COMPUTE EXPIRY THU AND NEXT THU
	// select upcoming thu
	if wkday <= time.Thursday {
		currThu = selDate.AddDate(0, 0, int(time.Thursday-wkday)) //  upcoming Thu
	} else {
		currThu = selDate.AddDate(0, 0, int(7-(wkday-time.Thursday))) //  recent passed Thu + 7 days
	}
	// select thu based on offset requested
	currThu = currThu.AddDate(0, 0, ts.CtrlParam.TradeSettings.OptionExpiryWeek*7)
	nextThu := currThu.AddDate(0, 0, 7)
	year, month, day := currThu.Date()

	// ---------------------------------------------------------------------- COMPUTE YEAR AND DAY
	yr := strconv.Itoa(year - 2000) // year in 2 digit format
	dy := strconv.Itoa(day)

	// ---------------------------------------------------------------------- COMPUTE MONTH
	switch month {
	case 10:
		mth = "O"
	case 11:
		mth = "N"
	case 12:
		mth = "D"
	default:
		mth = strconv.Itoa(int(month))
	}

	mnt3ltr := strings.ToUpper(currThu.Month().String()[:3])

	// ---------------------------------------------------------------------- ROUNDOFF STRIKE PRICE
	strikePriceStep, _ = strconv.Atoi(envMap[Instr])
	rnd := math.Round((order.Entry / float64(strikePriceStep))) * float64(strikePriceStep)

	// ---------------------------------------------------------------------- COMPUTE CE/PE and ITM/ATM/OTM Value
	if ts.CtrlParam.TradeSettings.OrderRoute == "option-buy" {
		if strings.ToLower(order.Dir) == "bullish" {
			optn = "CE"
			lvl = rnd + (float64(strikePriceStep) * float64(ts.CtrlParam.TradeSettings.OptionLevel))
		} else {
			optn = "PE"
			lvl = rnd - (float64(strikePriceStep) * float64(ts.CtrlParam.TradeSettings.OptionLevel))
		}
	} else if ts.CtrlParam.TradeSettings.OrderRoute == "option-sell" {
		if strings.ToLower(order.Dir) == "bearish" {
			optn = "PE"
			lvl = rnd - (float64(strikePriceStep) * float64(ts.CtrlParam.TradeSettings.OptionLevel))
		} else {
			optn = "CE"
			lvl = rnd + (float64(strikePriceStep) * float64(ts.CtrlParam.TradeSettings.OptionLevel))
		}
	} else {

	}

	// ---------------------------------------------------------------------- COMPUTE STRING
	// if expiry last in month use montly expiry
	if nextThu.Month() == currThu.Month() { // curr and next thu in same month?
		symbolFutStr = (Instr + yr + mth + dy + strconv.Itoa(int(lvl)) + optn)
	} else {
		symbolFutStr = (Instr + yr + mnt3ltr + strconv.Itoa(int(lvl)) + optn)
	}

	return symbolFutStr
}

// NIFTY21DECFUT
func deriveFuturesName(order data.TradeSignal, ts data.Strategies, selDate time.Time) string {

	var symbolFutStr string = "FAILED"

	Instr := strings.ReplaceAll(order.Instr, "-FUT", "") // remove -FUT suffix

	wkday := selDate.Weekday()
	currThu := time.Now() // dummy initialisation

	if wkday <= time.Thursday {
		currThu = selDate.AddDate(0, ts.CtrlParam.TradeSettings.FuturesExpiryMonth, int(time.Thursday-wkday)) //  upcoming Thu
	} else {
		currThu = selDate.AddDate(0, ts.CtrlParam.TradeSettings.FuturesExpiryMonth, int(7-(wkday-time.Thursday))) //  recent passed Thu + 7 days
	}
	nextThu := currThu.AddDate(0, 0, 7)

	if nextThu.Month() == currThu.Month() { // curr and next thu in same month?

		symbolFutStr = Instr + currThu.Format("06-Jan") + "FUT"

	} else {
		if ts.CtrlParam.TradeSettings.SkipExipryWeekFutures {
			symbolFutStr = Instr + nextThu.Format("06-Jan") + "FUT"
		} else {
			symbolFutStr = Instr + currThu.Format("06-Jan") + "FUT"
		}
	}

	symbolFutStr = strings.ReplaceAll(symbolFutStr, "-", "")
	symbolFutStr = strings.ToUpper(symbolFutStr)

	return symbolFutStr
}

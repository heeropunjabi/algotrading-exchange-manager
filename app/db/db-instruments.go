package db

import (
	"algo-ex-mgr/app/appdata"
	"algo-ex-mgr/app/srv"
	"context"
	"strconv"
)

func GetInstrumentsToken() map[string]string {

	var tokensMap = make(map[string]string)

	tknEq := getNseEqTokens()
	tknFut := getFuturesTokens()

	for k, v := range tknEq {
		tokensMap[k] = v
	}
	for k, v := range tknFut {
		tokensMap[k] = v
	}
	return tokensMap
}

func getFuturesTokens() map[string]string {

	var tokensMap = make(map[string]string)

	ctx := context.Background()
	myCon, _ := dbPool.Acquire(ctx)
	defer myCon.Release()

	rows, err := myCon.Query(ctx, sqlQueryFutures)

	if err != nil {
		srv.ErrorLogger.Printf("Cannot read list of tokens for ticker %v\n", err)
		return tokensMap
	}

	for rows.Next() {

		var itoken int64
		var symbol string

		err = rows.Scan(&itoken, &symbol)
		if err != nil {
			srv.ErrorLogger.Printf("Cannot parse list of tokens for ticker %v\n", err)
			return tokensMap
		}

		if rows.Err() != nil {
			srv.ErrorLogger.Println("Cannot parse list of tokens for ticker: ", rows.Err())

			return tokensMap
		}
		tokensMap[strconv.FormatInt(itoken, 10)] = symbol
	}
	defer rows.Close()

	return tokensMap
}

// TODO: Check for MCX, BSE and other symbols
func getNseEqTokens() map[string]string {

	var tokensMap = make(map[string]string)

	ctx := context.Background()
	myCon, _ := dbPool.Acquire(ctx)
	defer myCon.Release()

	rows, err := myCon.Query(ctx, sqlQueryNseEqTokens)

	if err != nil {
		srv.ErrorLogger.Printf("Cannot read list of tokens for ticker %v\n", err)
		return tokensMap
	}

	for rows.Next() {

		var itoken int64
		var symbol string

		err = rows.Scan(&itoken, &symbol)
		if err != nil {
			srv.ErrorLogger.Printf("Cannot parse list of tokens for ticker %v\n", err)
			return tokensMap
		}

		if rows.Err() != nil {
			srv.ErrorLogger.Println("Cannot parse list of tokens for ticker: ", rows.Err())

			return tokensMap
		}
		tokensMap[strconv.FormatInt(itoken, 10)] = symbol
	}
	defer rows.Close()

	return tokensMap
}

func FetchInstrData(instrument string, strikelevel uint64, opdepth int, instrtype string, startdate string, enddate string) (instrname string, lotsize float64) {

	lock.Lock()
	defer lock.Unlock()
	var size int64
	var name string

	ctx := context.Background()
	myCon, _ := dbPool.Acquire(ctx)
	defer myCon.Release()

	var err error
	if instrtype == "EQ" {

		sqlquery := DbQueryNameUpdate("DB_TBL_USER_SYMBOLS", appdata.Env["DB_TBL_USER_SYMBOLS"], sqlInstrDataQueryEQ)
		sqlquery = DbQueryNameUpdate("DB_TBL_INSTRUMENTS", appdata.Env["DB_TBL_INSTRUMENTS"], sqlquery)
		err = myCon.QueryRow(ctx, sqlquery,
			instrument).Scan(&name, &size)

	} else if instrtype == "FUT" {

		sqlquery := DbQueryNameUpdate("DB_TBL_USER_SYMBOLS", appdata.Env["DB_TBL_USER_SYMBOLS"], sqlInstrDataQueryFUT)
		sqlquery = DbQueryNameUpdate("DB_TBL_INSTRUMENTS", appdata.Env["DB_TBL_INSTRUMENTS"], sqlquery)

		err = myCon.QueryRow(ctx, sqlquery, instrument,
			startdate, enddate).Scan(&name, &size)

	} else {

		sqlquery := DbQueryNameUpdate("DB_TBL_USER_SYMBOLS", appdata.Env["DB_TBL_USER_SYMBOLS"], sqlInstrDataQueryOptn)
		sqlquery = DbQueryNameUpdate("DB_TBL_INSTRUMENTS", appdata.Env["DB_TBL_INSTRUMENTS"], sqlquery)

		err = myCon.QueryRow(ctx, sqlquery,
			instrument, strikelevel,
			opdepth, instrtype,
			startdate, enddate).Scan(&name, &size)
	}

	if err != nil {
		srv.ErrorLogger.Printf("FetchOrderData error %v\n", err.Error())
		return "", 0
	}

	return name, float64(size)
}

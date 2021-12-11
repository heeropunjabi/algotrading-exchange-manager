package db

import (
	"context"
	"goTicker/app/kite"
	"goTicker/app/srv"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

var dbPool *pgxpool.Pool

func DbInit() bool {
	// urlExample := "postgres://username:password@localhost:5432/database_name"

	ctx := context.Background()
	var err error

	dbUrl := os.Getenv("DATABASE_URL")
	dbPool, err = pgxpool.Connect(ctx, dbUrl)

	if err != nil {
		srv.ErrorLogger.Printf("Unable to connect to database: %v\n", err)
		return false
	}

	var greeting string
	err = dbPool.QueryRow(ctx, "select 'Hello, Timescale!'").Scan(&greeting)

	if err != nil {
		srv.ErrorLogger.Printf("QueryRow failed: %v\n", err)
		return false
	}
	srv.InfoLogger.Printf("connected to DB : " + greeting)

	// check if table exist, else create it
	queryCreateTicksTable := `CREATE TABLE dailyTicks (
                                                time TIMESTAMPTZ NOT NULL,
                                                symbol integer NULL,
                                                last_price double precision NULL,
                                                open double precision NULL,
                                                close double precision NULL,
                                                low double precision NULL,
                                                high double precision NULL,
                                                volume int NULL
                                            );
						SELECT create_hypertable('dailyTicks', 'time');
						`

	//execute statement, fails if table already exists
	_, _ = dbPool.Exec(ctx, queryCreateTicksTable)

	return true

}

func StoreTickInDb() {

	for v := range kite.ChTick {
		// fmt.Println("\nkite ch data rx ", v)
		//fmt.Println("Timestamp: ", v.Timestamp)
		ctx := context.Background()
		// kite.ChTick <- kite.TickData{Timestamp: "2021-11-30 22:12:10", Insttoken: 1, Lastprice: 1, Open: 1.1, High: 1.2, Low: 1.3, Close: 1.4, Volume: 9}

		queryInsertMetadata := `INSERT INTO dailyTicks (time, symbol, last_price, open, close, low, high, volume) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`
		_, err := dbPool.Exec(ctx, queryInsertMetadata, v.Timestamp, v.Insttoken, v.Lastprice, v.Open, v.Close, v.Low, v.High, v.Volume)
		if err != nil {
			srv.ErrorLogger.Printf("Unable to insert data into database: %v\n", err)
			//os.Exit(1)
		}
	}

}

func CloseDBPool() {
	dbPool.Close()
}

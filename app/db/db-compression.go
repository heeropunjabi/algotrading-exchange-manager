package db

import "context"

func setupDbCompression() {

	ctx := context.Background()
	myCon, _ := dbPool.Acquire(ctx)
	defer myCon.Release()

	_, _ = myCon.Exec(ctx, DB_COMPRESSION_QUERY)
	// if err != nil {
	// 	srv.WarningLogger.Printf("Error setting up DB Compression: %v\n", err)
	// }
}
package db

var DB_EXISTS_QUERY = "SELECT datname FROM pg_catalog.pg_database  WHERE lower(datname) = lower('algotrading');"
var DB_CREATE_QUERY = "CREATE DATABASE algotrading;"
var DB_TABLE_ID_DECODED_NAME = `token_id_decoded`
var DB_CREATE_TABLE_ID_DECODED = `CREATE TABLE token_id_decoded
								(
									time TIMESTAMP NOT NULL,
									nse_symbol VARCHAR(30),
									mcx_symbol VARCHAR(30)
								);
						`
var DB_TABLE_TICKER_NAME = `ticks_data`
var DB_CREATE_TABLE_TICKER = `CREATE TABLE 
								ticks_data
							 (
									time TIMESTAMP NOT NULL,
									symbol VARCHAR(30) NOT NULL,
									last_traded_price double precision NOT NULL DEFAULT 0,
									buy_demand bigint NOT NULL DEFAULT 0,
									sell_demand bigint NOT NULL DEFAULT 0,
									trades_till_now bigint NOT NULL DEFAULT 0,
									open_interest bigint NOT NULL DEFAULT 0
								);
							SELECT create_hypertable('ticks_data', 'time');
							SELECT set_chunk_time_interval('ticks_data', INTERVAL '7 days');
						`

var DB_COMPRESSION_QUERY = `ALTER TABLE ticks_data SET 
							(
								timescaledb.compress,
								timescaledb.compress_segmentby = 'symbol'
							); 
							SELECT add_compression_policy('ticks_data ', INTERVAL '30 days');
						`

var DB_VIEW_EXISTS_QUERY = `
					SELECT view_name 
						FROM timescaledb_information.continuous_aggregates
						WHERE view_name = $1;`

// $1 candles_5min
// $2 5
var DB_CREATE_VIEW = `
					CREATE MATERIALIZED VIEW candles_$1min
						WITH (timescaledb.continuous) AS
						SELECT time_bucket('$1 minutes', time) AS candle, 
							symbol,
							FIRST(last_traded_price, time) as open,
							MAX(last_traded_price) as high,
							MIN(last_traded_price) as low,
							LAST(last_traded_price, time) as close,
							LAST(trades_till_now, time) - FIRST(trades_till_now, time) as volume
						FROM
							ticks_data
						
						GROUP by
							symbol, candle
						WITH NO DATA;
						`

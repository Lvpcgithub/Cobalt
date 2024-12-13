package dao

import (
	"Cobalt/config"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"log"
)

// 连接数据库
func ConnectToDB() *sql.DB {
	dsn := config.Mysqldb
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("Error ConnectToDB:", err)
		return nil
	}
	return db
}

// 连接redis
func ConnRedis() redis.Conn {
	// 连接 Redis
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	return conn
}
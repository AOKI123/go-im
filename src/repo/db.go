package repo

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DatabaseOps *gorm.DB = Connect(DataSource{
	Host:     "localhost",
	Port:     3306,
	Database: "go_im",
	User:     "root",
	Password: "12345678",
})

type DataSource struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

func Connect(ds DataSource) *gorm.DB {
	var (
		db    *gorm.DB
		sqlDB *sql.DB
		err   error
	)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		ds.User, ds.Password, ds.Host, ds.Port, ds.Database)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, err = db.DB()
	if err != nil {
		panic(err)
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetConnMaxIdleTime(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置了连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}

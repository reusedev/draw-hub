package mysql

import (
	"fmt"
	"github.com/reusedev/draw-hub/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitMySQL(config config.MySQL) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database, config.Charset)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	DB = db
}

func CreateDataBase(config config.MySQL) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=True&loc=Local", config.Username, config.Password, config.Host, config.Port)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		panic(err)
	}
	err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET %s", config.Database, config.Charset)).Error
	if err != nil {
		panic(err)
	}
}

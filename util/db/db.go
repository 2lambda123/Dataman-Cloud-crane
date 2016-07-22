package db

import (
	"github.com/Dataman-Cloud/rolex/plugins/auth"
	"github.com/Dataman-Cloud/rolex/util/config"

	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/mattes/migrate/driver/mysql"
)

var db *gorm.DB

func DB() *gorm.DB {
	if db == nil {
		InitDB()
	}
	return db
}

func InitDB() {
	var err error
	conf := config.GetConfig()
	log.Infof("mysql connection uri: %s", conf.DbDSN)

	db, err = gorm.Open(conf.DbDriver, conf.DbDSN)
	if err != nil {
		log.Fatalf("init mysql error: %v", err)
	}
	db.DB().SetMaxIdleConns(5)
	db.DB().SetMaxOpenConns(50)
	db.SetLogger(log.StandardLogger())
	db.LogMode(true)
	MigriateTable()
}

func MigriateTable() {
	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&auth.Account{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&auth.Group{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&auth.AccountGroup{})
}

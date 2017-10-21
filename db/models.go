package db

import (
	"time"
	"os"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type MCompany struct {
	Code int `gorm:"primary_key"`
	MTokyoScId int
	Name string
	WhoisRegistName string
	WebSite string
	CreatedAt time.Time
}

type DomainList struct {
	MCompanyCode int
	Domain string
	ReportingDate time.Time
	CreatedAt time.Time
}

func DbOpen() *gorm.DB {
	rdbms := "mysql";
	dbName := os.Getenv("DB_DATABASE");
	user := os.Getenv("DB_USER");
	pass := os.Getenv("DB_PASSWORD");

	db, err := gorm.Open(rdbms,  user+":"+pass+"@/"+dbName+"?parseTime=true&loc=Local")
	if err != nil {
		panic(err.Error())
	}

	db.SingularTable(true)

	return db
}

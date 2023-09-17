package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// sql.go
// sql まわりの操作関数

func openDB(dbName string) (db *gorm.DB, err error) {
	// DBファイルのオープン
	db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	return db, err
}

func migrateDB(db *gorm.DB) (err error) {
	// Sqlite3 DB の テーブルを struct から 作成 or マイグレートする
	if err := db.AutoMigrate(&PersonalScoreRow{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&DayRankingTableRow{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&MonthryRankingTableRow{}); err != nil {
		panic(err)
	}
	return nil
}

// getMaxId 最大のIDを取得する
func getMaxId(db *gorm.DB, table interface{}) (maxId int64, err error) {
	stmt := &gorm.Statement{DB: db}
	_ = stmt.Parse(table)
	fmt.Println(stmt.Schema.Table) // Output: users
	tableName := stmt.Schema.Table
	// 最大のIDを取得
	var count int64
	db.Model(table).Count(&count)
	if count == 0 {
		return 0, nil
	}
	maxId = 0
	err = db.Raw("SELECT MAX(id) FROM ?", tableName).Scan(&maxId).Error
	return maxId, err
}

// truncateTable 指定テーブルのレコードを全削除する
func truncateTable(db *gorm.DB, tableName interface{}) error {
	db.Unscoped().Where("1 = 1").Delete(tableName)
	return db.Error
}

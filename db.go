package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"os"
	"strings"
)

type DataBaseConfig struct {
	DbName   string `json:"db_name"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

func connectionToMysqlServer() (*sql.DB, error) {
	config, err := readConfig()
	if err != nil {
		return nil, err
	}

	connectionString := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", config.User, config.Password, config.Protocol, config.Host, config.Port, config.DbName)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func readConfig() (*DataBaseConfig, error) {
	var config DataBaseConfig

	fileConfig, err := os.Open("./dbConfig.json")
	if err != nil {
		return nil, err
	}
	defer fileConfig.Close()

	jsonConfig, err := ioutil.ReadAll(fileConfig)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func createDbSchema() error {
	file, err := ioutil.ReadFile("./schema.sql")
	if err != nil {
		return err
	}

	queries := strings.SplitAfter(string(file), ";")

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	for _, query := range queries[:len(queries)-1] {
		_, err := tx.Exec(query)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("\ncreateDbSchema: Error query: %s\nError text:%s", query, err)
		}
	}

	if tx.Commit() != nil {
		return err
	}

	return nil
}

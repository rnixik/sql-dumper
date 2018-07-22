package main

import (
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"os"
)

func dbConnect(conset *ConnectionSettings) (db *sqlx.DB, err error) {
	return sqlx.Open(conset.driver, conset.dsn())
}

func main() {
	configFile := flag.String("config", ".env", "source label file")
	format := flag.String("format", "sql", "Output format: sql, csv, simple")
	dstFile := flag.String("file", "result.sql", "Filename for single output file")
	dstDir := flag.String("dir", "./", "Output directory for multiple output files")
	flag.Parse()

	err := Run(dbConnect, flag.Args(), *configFile, *format, *dstFile, *dstDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

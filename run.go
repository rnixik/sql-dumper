package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"os"
)

func Run(dbConnect dbConnector, argsTail []string, configFile string, format string, fw FileWriter, dstFile string, dstDir string, csvDelimiter string) (err error) {
	if len(argsTail) != 2 && len(argsTail) != 3 {
		showHelp()
		return
	}

	conset, err := getConnectionSettings(configFile)
	if err != nil {
		return err
	}

	tablesPart := argsTail[0]
	intervalPart := argsTail[1]
	relationsPart := ""
	if len(argsTail) == 3 {
		relationsPart = argsTail[2]
	}

	query, err := ParseRequest(tablesPart, intervalPart, relationsPart)
	if err != nil {
		return err
	}

	combined := true
	var writer DataWriter
	writer = &SimpleWriter{}
	if format == "sql" {
		combined = false
		if dstFile == "" && dstDir == "" {
			dstFile = "result.sql"
		}
		writer = NewSqlWriter(fw, dstFile, dstDir)
	} else if format == "csv" {
		if dstFile == "" && dstDir == "" {
			dstFile = "result.csv"
		}
		if dstDir != "" {
			combined = false
		}
		writer = NewCsvWriter(fw, dstFile, dstDir, csvDelimiter)
	}

	err = query.QueryResult(dbConnect, conset, writer, combined)
	if err != nil {
		return err
	}
	return nil
}

func getConnectionSettings(configFile string) (*ConnectionSettings, error) {
	if os.Getenv("DB_NAME") == "" {
		cfg, err := ini.Load(configFile)
		if err != nil {
			return nil, fmt.Errorf("Empty DB_NAME in environment and fail to read config file: %v", err)
		}
		cfgSection := cfg.Section("")
		os.Setenv("DB_USER", cfgSection.Key("DB_USER").String())
		os.Setenv("DB_PASSWORD", cfgSection.Key("DB_PASSWORD").String())
		os.Setenv("DB_NAME", cfgSection.Key("DB_NAME").String())
		os.Setenv("DB_HOST", cfgSection.Key("DB_HOST").String())
	}

	return &ConnectionSettings{
		driver:   "mysql",
		user:     os.Getenv("DB_USER"),
		password: os.Getenv("DB_PASSWORD"),
		dbname:   os.Getenv("DB_NAME"),
		dbhost:   os.Getenv("DB_HOST"),
	}, nil
}

func showHelp() {
	usage := "Dumps data from DB.\n"
	usage += "\n"
	usage += "Usage: sql-dumper [OPTIONS] <tables> <interval> [relations]\n"
	usage += "\n"
	usage += "Options:\n"
	usage += "  --config <filename>        File with settings of connection to DB.\n"
	usage += "                             It will be used if environment variable DB_NAME is not defined (default .env)\n"
	usage += "  --format {sql|csv|simple}  Format of output format (default sql)\n"
	usage += "  --csv-delimiter            Sets delimiter of values in CSV (default ,)\n"
	usage += "  --file <filename>          Specify file to save combined result from all tables. Can't be used with --dir (default result.sql)\n"
	usage += "  --dir <directory>          Specify directory to save the result in a separete file for every table\n"
	usage += "\n"
	usage += "Arguments:\n"
	usage += "\n"
	usage += "  tables     List of tables and columns to dump: table1:column11,column12,...,column1N;table2:column21;...\n"
	usage += "  interval   Interval of values for the first column in the first table to select from DB: int-int\n"
	usage += "  relations  List of relations between chosen tables and columns:\n"
	usage += "             table1.column11=table2.column21;table2.column22=table3.column31\n"
	usage += "\n"
	usage += "Example:\n"
	usage += "\n"
	usage += "  sql-dumper \"routes:id,name;stations:id,name;stations_for_routes:station_id,route_id,ord\" \\\n"
	usage += "     2000-2200 \\\n"
	usage += "     \"routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id\"\n"
	fmt.Fprintln(os.Stderr, usage)
}

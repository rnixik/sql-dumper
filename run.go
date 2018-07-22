package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"os"
)

func Run(dbConnect dbConnector, argsTail []string, configFile string, format string, fw FileWriter, dstFile string, dstDir string) (err error) {
	if len(argsTail) != 2 && len(argsTail) != 3 {
		showHelp()
		return
	}

	if os.Getenv("DB_NAME") == "" {
		cfg, err := ini.Load(configFile)
		if err != nil {
			return fmt.Errorf("Empty DB_NAME in environment and fail to read config file: %v", err)
		}
		cfgSection := cfg.Section("")
		os.Setenv("DB_USER", cfgSection.Key("DB_USER").String())
		os.Setenv("DB_PASSWORD", cfgSection.Key("DB_PASSWORD").String())
		os.Setenv("DB_NAME", cfgSection.Key("DB_NAME").String())
		os.Setenv("DB_HOST", cfgSection.Key("DB_HOST").String())
	}

	tablesPart := argsTail[0]
	intervalPart := argsTail[1]
	relationsPart := ""
	if len(argsTail) == 3 {
		relationsPart = argsTail[2]
	}

	conset := &ConnectionSettings{
		driver:   "mysql",
		user:     os.Getenv("DB_USER"),
		password: os.Getenv("DB_PASSWORD"),
		dbname:   os.Getenv("DB_NAME"),
		dbhost:   os.Getenv("DB_HOST"),
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
		writer = NewSqlWriter(fw, dstFile, dstDir)
	}

	err = query.QueryResult(dbConnect, conset, writer, combined)
	if err != nil {
		return err
	}
	return nil
}

func showHelp() {
	fmt.Println("Dumps data from DB.")
	fmt.Println("Usage:")
	fmt.Println("    sql-dumper \"tables definitions\" \"primary interval\" [\"relations definitions\"]")
	fmt.Println("")
	fmt.Println("Format of tables definitions: table1:column11,column12,...,column1N;table2:column21;...")
	fmt.Println("Format of primary interval: int-int")
	fmt.Println("Format of relations: table1.column11=table2.column21;table2.column22=table3.column31")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("    sql-dumper \"routes:id,name;stations:id,name;stations_for_routes:station_id,route_id,ord\" \\")
	fmt.Println("        2000-2200 \\")
	fmt.Println("        \"routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id\"")
}

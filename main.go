package main

import (
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
)

func main() {
	configFile := flag.String("config", ".env", "source label file")
	flag.Parse()

	tail := flag.Args()
	if len(tail) != 2 && len(tail) != 3 {
		showHelp()
		return
	}

	if os.Getenv("DB_NAME") == "" {
		cfg, err := ini.Load(*configFile)
		if err != nil {
			fmt.Printf("Empty DB_NAME in environment and fail to read config file: %v", err)
			os.Exit(1)
		}
		cfgSection := cfg.Section("")
		os.Setenv("DB_USER", cfgSection.Key("DB_USER").String())
		os.Setenv("DB_PASSWORD", cfgSection.Key("DB_PASSWORD").String())
		os.Setenv("DB_NAME", cfgSection.Key("DB_NAME").String())
		os.Setenv("DB_HOST", cfgSection.Key("DB_HOST").String())
	}

	tablesPart := tail[0]
	intervalPart := tail[1]
	relationsPart := ""
	if len(tail) == 3 {
		relationsPart = tail[2]
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
		fmt.Print(err)
		os.Exit(1)
	}
	err = query.QueryResult(conset)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Dumps data from DB into file.")
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

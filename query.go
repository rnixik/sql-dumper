package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
)

type QueryTable struct {
	name    string
	columns []string
}

type QueryRelation struct {
	table1  string
	column1 string
	table2  string
	column2 string
}

type Query struct {
	tables          []*QueryTable
	relations       []*QueryRelation
	primaryInterval []int64
}

type ConnectionSettings struct {
	driver   string
	user     string
	password string
	dbname   string
	dbhost   string
}

func (q *Query) QueryResult(conset *ConnectionSettings) (err error) {
	db, err := getDb(conset)
	if err != nil {
		log.Fatal(err)
	}

	var query string
	for i, qt := range q.tables {
		if i == 0 {
			query, err = q.toSqlForSingleTable(qt)
		} else {
			query, err = q.toSqlForRelation(qt)
		}
		if err != nil {
			return
		}
		fmt.Println(query)
		resultsMaps, err := dbSelect(db, query, q.primaryInterval[0], q.primaryInterval[1])
		if err != nil {
			return err
		}
		DumpResults(resultsMaps)
	}
	return
}

func dbSelect(db *sqlx.DB, query string, args ...interface{}) (resultsMaps []map[string]interface{}, err error) {
	rows, err := db.Queryx(query, args...)
	if err != nil {
		return
	}

	resultsMaps = make([]map[string]interface{}, 0)
	for rows.Next() {
		results := make(map[string]interface{})
		err = rows.MapScan(results)
		if err != nil {
			return
		}
		resultsMaps = append(resultsMaps, results)
	}
	return resultsMaps, nil
}

func getDb(conset *ConnectionSettings) (db *sqlx.DB, err error) {
	dsn := conset.user + ":" + conset.password + "@tcp(" + conset.dbhost + ")/" + conset.dbname
	return sqlx.Open(conset.driver, dsn)
}

func (q *Query) toSqlForSingleTable(qt *QueryTable) (str string, err error) {
	str = "SELECT " + qt.sqlPartForSelectColumns() + "\n"
	str += "FROM " + sqlTable(qt.name) + "\n"
	str += "WHERE " + sqlTableAndColumn(qt.name, qt.columns[0]) + " BETWEEN ? AND ?"
	return str, nil
}

func (q *Query) toSqlForRelation(qt *QueryTable) (str string, err error) {
	subquery, leftTableColumn, err := q.toSqlSubQueryForRelation(qt)
	if err != nil {
		return
	}
	str = "SELECT " + qt.sqlPartForSelectColumns() + "\n"
	str += "FROM " + sqlTable(qt.name) + "\n"
	str += "WHERE " + leftTableColumn + " IN \n"
	str += "(\n" + subquery + "\n)"
	return
}

func (qt *QueryTable) sqlPartForSelectColumns() string {
	selectFields := make([]string, 0)
	for _, qtcol := range qt.columns {
		selectFields = append(selectFields, "`"+qt.name+"`.`"+qtcol+"`")
	}
	return strings.Join(selectFields, ", ")
}

func (qt *QueryTable) sqlPartForTableName() string {
	return "`" + qt.name + "`"
}

func sqlTable(name string) string {
	return "`" + name + "`"
}

func sqlColumn(name string) string {
	return "`" + name + "`"
}

func sqlTableAndColumn(table string, column string) string {
	return sqlTable(table) + "." + sqlColumn(column)
}

func (q *Query) toSqlSubQueryForRelation(mainTable *QueryTable) (subquery string, leftTableColumn string, err error) {
	if mainTable.name == q.tables[0].name {
		return "", "", fmt.Errorf("Cannot build subquery for first table in list")
	}

	rightTableColumn := ""
	for _, qr := range q.relations {
		if qr.table1 == mainTable.name {
			leftTableColumn = sqlTableAndColumn(qr.table1, qr.column1)
			rightTableColumn = sqlTableAndColumn(qr.table2, qr.column2)
			break
		}
		if qr.table2 == mainTable.name {
			leftTableColumn = sqlTableAndColumn(qr.table2, qr.column2)
			rightTableColumn = sqlTableAndColumn(qr.table1, qr.column1)
			break
		}
	}
	if rightTableColumn == "" {
		return "", "", fmt.Errorf("Cannot find relation for table '%s'. Relations: %v", mainTable.name, q.relations)
	}

	subquery = "SELECT " + rightTableColumn + "\n"
	subquery += "FROM "
	selectTables := make([]string, 0)
	for _, qt := range q.tables {
		if mainTable == qt {
			continue
		}
		selectTables = append(selectTables, sqlTable(qt.name))
	}
	subquery += strings.Join(selectTables, ", ")
	subquery += "\n"
	subquery += "WHERE "
	whereConditions := make([]string, 0)
	firstTable := q.tables[0]
	firstCondition := sqlTableAndColumn(firstTable.name, firstTable.columns[0]) + " BETWEEN ? AND ?"
	whereConditions = append(whereConditions, firstCondition)
	for _, qr := range q.relations {
		if qr.table1 == mainTable.name || qr.table2 == mainTable.name {
			continue
		}
		condition := sqlTableAndColumn(qr.table1, qr.column1) + " = " + sqlTableAndColumn(qr.table2, qr.column2)
		whereConditions = append(whereConditions, condition)
	}
	subquery += "(" + strings.Join(whereConditions, ") AND (") + ")"
	return
}

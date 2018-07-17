package main

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

// QueryTable represents definition of one table for sql query
type QueryTable struct {
	name    string
	columns []string
}

// QueryRelation represents definition of relations between tables for sql query
type QueryRelation struct {
	table1  string
	column1 string
	table2  string
	column2 string
}

// Query represents information for building final sql queries
type Query struct {
	tables          []*QueryTable
	relations       []*QueryRelation
	primaryInterval []int64
}

// ConnectionSettings contains settings for DB connection
type ConnectionSettings struct {
	driver   string
	user     string
	password string
	dbname   string
	dbhost   string
}

type TableColumnDDL struct {
	Field   string         `db:"Field"`
	Type    string         `db:"Type"`
	Null    string         `db:"Null"`
	Key     string         `db:"Key"`
	Default sql.NullString `db:"Default"`
	Extra   string         `db:"Extra"`
}

type dbConnector func(conset *ConnectionSettings) (db *sqlx.DB, err error)

// QueryResult returns rows of data from DB
func (q *Query) QueryResult(dbConnect dbConnector, conset *ConnectionSettings, writer DataWriter) (err error) {
	if len(q.primaryInterval) != 2 {
		return fmt.Errorf("primaryInterval should contain two values")
	}

	db, err := dbConnect(conset)
	if err != nil {
		return err
	}

	var query string
	for i, qt := range q.tables {
		if i == 0 {
			query = q.toSqlForSingleTable(qt)
		} else {
			query, err = q.toSqlForRelation(qt)
		}
		if err != nil {
			return
		}
		//fmt.Println(query)
		resultsMaps, err := dbSelect(db, query, q.primaryInterval[0], q.primaryInterval[1])
		if err != nil {
			return err
		}
		writer.Write(resultsMaps)
	}

	_, err = q.toDDL(db)
	if err != nil {
		return
	}

	return
}

func dbSelect(db *sqlx.DB, query string, args ...interface{}) (resultsMaps []*map[string]interface{}, err error) {
	rows, err := db.Queryx(query, args...)
	if err != nil {
		return
	}

	resultsMaps = make([]*map[string]interface{}, 0)
	for rows.Next() {
		results := make(map[string]interface{})
		err = rows.MapScan(results)
		if err != nil {
			return
		}
		resultsMaps = append(resultsMaps, &results)
	}
	return resultsMaps, nil
}

func (q *Query) toSqlForSingleTable(qt *QueryTable) (str string) {
	str = "SELECT " + qt.sqlPartForSelectColumns() + "\n"
	str += "FROM " + sqlTable(qt.name) + "\n"
	str += "WHERE " + sqlTableAndColumn(qt.name, qt.columns[0]) + " BETWEEN ? AND ?"
	return str
}

func (q *Query) toSqlForRelation(qt *QueryTable) (str string, err error) {
	subquery, leftTableColumn, err := q.toSqlSubQueryForRelation(qt)
	if err != nil {
		return
	}
	str = "SELECT " + qt.sqlPartForSelectColumns() + "\n"
	str += "FROM " + sqlTable(qt.name) + "\n"
	str += "WHERE " + leftTableColumn + " IN\n"
	str += "(\n" + subquery + "\n)"
	return
}

func (q *Query) toDDL(db *sqlx.DB) (ddl string, err error) {
	for _, qt := range q.tables {
		tableDescribtion, err := getTableDescription(db, qt.name)
		if err != nil {
			return "", err
		}
		tableDDL, err := makeDDLFromTableDescription(qt.name, tableDescribtion, qt.columns, q.relations)
		if err != nil {
			return "", err
		}
		//fmt.Printf("%s\n", tableDDL)
		ddl += tableDDL
	}
	return ddl, nil
}

func getTableDescription(db *sqlx.DB, tableName string) (tableDescribtion []TableColumnDDL, err error) {
	columnsDDL := []TableColumnDDL{}
	err = db.Select(&columnsDDL, "DESCRIBE "+sqlTable(tableName))
	return columnsDDL, err
}

func makeDDLFromTableDescription(tableName string, tableDescribtion []TableColumnDDL, columnsOnly []string, relations []*QueryRelation) (tableDDL string, err error) {
	columnsDDLs := []string{}
	primaryKeys := []string{}
	indexColumns := []string{}
	uniqueColumns := []string{}
	possibleFKDefs := map[string]string{}
	for _, columnDescr := range tableDescribtion {
		if !contains(columnsOnly, columnDescr.Field) {
			continue
		}
		columnDDL := sqlColumn(columnDescr.Field) + " " + columnDescr.Type + " "
		if columnDescr.Null == "YES" {
			columnDDL += "NULL"
		} else {
			columnDDL += "NOT NULL"
		}
		if columnDescr.Default.Valid {
			columnDDL += " "
			columnDDL += "DEFAULT '" + columnDescr.Default.String + "'"
		}
		columnsDDLs = append(columnsDDLs, columnDDL)
		if columnDescr.Key == "PRI" {
			primaryKeys = append(primaryKeys, sqlColumn(columnDescr.Field))
		}
		if columnDescr.Key == "MUL" {
			indexColumns = append(indexColumns, sqlColumn(columnDescr.Field))
		}
		if columnDescr.Key == "UNI" {
			uniqueColumns = append(uniqueColumns, sqlColumn(columnDescr.Field))
		}
		rTable, rColumn, _ := findRelation(relations, tableName, columnDescr.Field)
		if rColumn != "" {
			possibleFKDefs[sqlColumn(columnDescr.Field)] = "CONSTRAINT " + sqlColumn(columnDescr.Field) + " FOREIGN KEY (" + sqlColumn(rColumn) + ") REFERENCES " + sqlTable(rTable) + " (`id`) ON DELETE CASCADE"
		}
	}

	if len(columnsDDLs) == 0 {
		return "", fmt.Errorf("Table '%s' contains 0 of specified fields", tableName)
	}

	rows := columnsDDLs
	if len(primaryKeys) > 0 {
		rows = append(rows, "PRIMARY KEY ("+strings.Join(primaryKeys, ", ")+")")
	}
	for _, ind := range indexColumns {
		rows = append(rows, "INDEX "+ind+" ("+ind+")")
	}
	for _, unq := range uniqueColumns {
		rows = append(rows, "UNIQUE INDEX "+unq+" ("+unq+")")
	}
	for column, fk := range possibleFKDefs {
		if !(len(primaryKeys) == 1 && primaryKeys[0] == column) {
			rows = append(rows, fk)
		}
	}

	for i, row := range rows {
		rows[i] = "    " + row
	}

	tableDDL = "CREATE TABLE " + sqlTable(tableName) + " (\n"
	tableDDL += strings.Join(rows, ",\n")
	tableDDL += "\n"
	tableDDL += ");"
	return tableDDL, nil
}

func (qt *QueryTable) sqlPartForSelectColumns() string {
	selectFields := make([]string, 0)
	for _, qtcol := range qt.columns {
		selectFields = append(selectFields, "`"+qt.name+"`.`"+qtcol+"`")
	}
	return strings.Join(selectFields, ", ")
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

func findRelation(relations []*QueryRelation, tableName string, tableColumn string) (rightTableName string, rightTableColumn string, err error) {
	for _, qr := range relations {
		if qr.table1 == tableName && qr.column1 == tableColumn {
			rightTableName = qr.table2
			rightTableColumn = qr.column2
			break
		}
		if qr.table2 == tableName && qr.column2 == tableColumn {
			rightTableName = qr.table1
			rightTableColumn = qr.column1
			break
		}
	}
	if rightTableColumn == "" {
		return "", "", fmt.Errorf("Cannot find relation for column '%s' of table '%s'", tableColumn, tableName)
	}
	return
}

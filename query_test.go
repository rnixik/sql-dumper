package main

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"reflect"
	"testing"
)

var typicalQuery = &Query{
	tables: []*QueryTable{
		{"routes", []string{"id", "name"}},
		{"stations", []string{"id", "sname"}},
		{"stations_for_routes", []string{"station_id", "route_id", "ord"}},
	},
	relations: []*QueryRelation{
		{"routes", "id", "stations_for_routes", "route_id"},
		{"stations", "id", "stations_for_routes", "station_id"},
	},
	primaryInterval: []int64{1000, 2000},
}

type EmptyWriter struct {
}

func (w *EmptyWriter) WriteRows(_ string, _ []string, _ []*map[string]interface{}) (err error) {
	return nil
}

func (w *EmptyWriter) WriteDDL(_ string, _ string) (err error) {
	return nil
}

func TestQueryResult(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	dbConnectMock := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return sqlxDB, nil
	}

	mock.ExpectQuery("DESCRIBE `routes`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("DESCRIBE `stations`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("DESCRIBE `stations_for_routes`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("station_id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("SELECT (.+) FROM `routes` WHERE `routes`.`id` BETWEEN \\? AND \\?").
		WithArgs(1000, 2000).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	mock.ExpectQuery("SELECT (.+) FROM `stations` WHERE `stations`.`id` IN (.+)").
		WithArgs(1000, 2000).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	mock.ExpectQuery("SELECT (.+) FROM `stations_for_routes` WHERE `stations_for_routes`.`route_id` IN (.+)").
		WithArgs(1000, 2000).
		WillReturnRows(sqlmock.NewRows([]string{"station_id", "route_id", "ord"}))

	err = typicalQuery.QueryResult(dbConnectMock, &ConnectionSettings{}, &EmptyWriter{}, false)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
}

func TestQueryResultIntervalError(t *testing.T) {
	dbConnect := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return nil, nil
	}

	var simpleQuery = &Query{
		tables: []*QueryTable{
			{"some_table", []string{"id"}},
		},
		relations:       []*QueryRelation{},
		primaryInterval: []int64{},
	}

	err := simpleQuery.QueryResult(dbConnect, &ConnectionSettings{}, &SimpleWriter{}, false)
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestQueryResultConnectionError(t *testing.T) {
	dbConnect := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return nil, fmt.Errorf("Some DB error")
	}
	err := typicalQuery.QueryResult(dbConnect, &ConnectionSettings{}, &SimpleWriter{}, false)
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestQueryResultSingleQueryError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	dbConnectMock := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return sqlxDB, nil
	}

	mock.ExpectQuery("DESCRIBE `routes`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("DESCRIBE `stations`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("DESCRIBE `stations_for_routes`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("station_id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("SELECT (.+) FROM `routes` WHERE `routes`.`id` BETWEEN \\? AND \\?").
		WithArgs(1000, 2000).
		WillReturnError(fmt.Errorf("Some error"))

	err = typicalQuery.QueryResult(dbConnectMock, &ConnectionSettings{}, &EmptyWriter{}, false)
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestQueryResultRelationError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	dbConnectMock := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return sqlxDB, nil
	}

	queryWithBadRelations := &Query{
		tables: []*QueryTable{
			{"routes", []string{"id", "name"}},
			{"stations", []string{"id", "sname"}},
		},
		relations: []*QueryRelation{
			{"routes", "id", "stations_for_routes", "route_id"},
		},
		primaryInterval: []int64{1000, 2000},
	}

	mock.ExpectQuery("DESCRIBE `routes`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("DESCRIBE `stations`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("SELECT (.+) FROM `routes` WHERE `routes`.`id` BETWEEN \\? AND \\?").
		WithArgs(1000, 2000).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	err = queryWithBadRelations.QueryResult(dbConnectMock, &ConnectionSettings{}, &EmptyWriter{}, false)
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestQueryResultDDLError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	dbConnectMock := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return sqlxDB, nil
	}

	var simpleQuery = &Query{
		tables: []*QueryTable{
			{"some_table", []string{"id"}},
		},
		relations:       []*QueryRelation{},
		primaryInterval: []int64{1, 2},
	}

	mock.ExpectQuery("SELECT (.+) FROM `some_table`").
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	mock.ExpectQuery("DESCRIBE `some_table`").
		WillReturnError(fmt.Errorf("Some error"))

	err = simpleQuery.QueryResult(dbConnectMock, &ConnectionSettings{}, &EmptyWriter{}, false)
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestQueryResultCombined(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	dbConnectMock := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return sqlxDB, nil
	}

	mock.ExpectQuery("DESCRIBE `routes`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("DESCRIBE `stations`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("DESCRIBE `stations_for_routes`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("station_id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("SELECT (.+)").
		WithArgs(1000, 2000).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	err = typicalQuery.QueryResult(dbConnectMock, &ConnectionSettings{}, &EmptyWriter{}, true)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
}

func TestQueryResultFileError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	dbConnectMock := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return sqlxDB, nil
	}

	var simpleQuery = &Query{
		tables: []*QueryTable{
			{"some_table", []string{"id"}},
		},
		relations:       []*QueryRelation{},
		primaryInterval: []int64{1, 2},
	}

	mock.ExpectQuery("DESCRIBE `some_table`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	fw := &TestFileErrorWriter{}
	writer := NewSqlWriter(fw, "result.sql", "")

	err = simpleQuery.QueryResult(dbConnectMock, &ConnectionSettings{}, writer, true)
	if err == nil {
		t.Errorf("Expected error by file writer, but got nil")
		return
	}
}

func TestSelectAndWriteCombinedFileError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	var simpleQuery = &Query{
		tables: []*QueryTable{
			{"some_table", []string{"id"}},
		},
		relations:       []*QueryRelation{},
		primaryInterval: []int64{1, 2},
	}

	mock.ExpectQuery("SELECT (.+)").
		WithArgs(1, 2).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "some"),
		)

	fw := &TestFileErrorWriter{}
	writer := NewSqlWriter(fw, "result.sql", "")

	err = simpleQuery.selectAndWrite(sqlxDB, writer, true)
	if err == nil {
		t.Errorf("Expected error by file writer, but got nil")
		return
	}
}

func TestSelectAndWriteFileError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	var simpleQuery = &Query{
		tables: []*QueryTable{
			{"some_table", []string{"id"}},
		},
		relations:       []*QueryRelation{},
		primaryInterval: []int64{1, 2},
	}

	mock.ExpectQuery("SELECT (.+)").
		WithArgs(1, 2).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "some"),
		)

	fw := &TestFileErrorWriter{}
	writer := NewSqlWriter(fw, "result.sql", "")

	err = simpleQuery.selectAndWrite(sqlxDB, writer, false)
	if err == nil {
		t.Errorf("Expected error by file writer, but got nil")
		return
	}
}

func TestQueryResultCombinedError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	dbConnectMock := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return sqlxDB, nil
	}

	var simpleQuery = &Query{
		tables: []*QueryTable{
			{"some_table", []string{"id"}},
		},
		relations:       []*QueryRelation{},
		primaryInterval: []int64{1, 2},
	}

	mock.ExpectQuery("DESCRIBE `some_table`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("SELECT (.+)").
		WithArgs(1000, 2000).
		WillReturnError(fmt.Errorf("Some error"))

	err = simpleQuery.QueryResult(dbConnectMock, &ConnectionSettings{}, &EmptyWriter{}, true)
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestDbSelect(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	columns := []string{"id", "name"}

	mock.ExpectQuery("SELECT (.+) FROM some_table WHERE id BETWEEN \\? AND \\?").
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "name1").AddRow(2, "name2"))

	resultsMaps, err := dbSelect(sqlxDB, "SELECT id, name FROM some_table WHERE id BETWEEN ? AND ?", 10, 20)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	expectedResultsMaps := []*map[string]interface{}{
		{"id": 1, "name": "name1"},
		{"id": 2, "name": "name2"},
	}

	if !reflect.DeepEqual(expectedResultsMaps, resultsMaps) {
		t.Errorf("Unexpected results")
		return
	}

	mock.ExpectQuery("SELECT (.+) FROM some_table WHERE id BETWEEN \\? AND \\?").
		WithArgs(10, 20).
		WillReturnError(fmt.Errorf("Some error"))

	_, err = dbSelect(sqlxDB, "SELECT id, name FROM some_table WHERE id BETWEEN ? AND ?", 10, 20)
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestToSqlForSingleTable(t *testing.T) {
	sql := typicalQuery.toSqlForSingleTable(typicalQuery.tables[0])
	expected := "SELECT `routes`.`id`, `routes`.`name`\n" +
		"FROM `routes`\n" +
		"WHERE `routes`.`id` BETWEEN ? AND ?"
	if sql != expected {
		t.Errorf("EXP:\n%s\nGOT:\n%s\n", expected, sql)
	}
}

func TestToSqlForRelation(t *testing.T) {
	sql, _ := typicalQuery.toSqlForRelation(typicalQuery.tables[1])
	expected := "SELECT `stations`.`id`, `stations`.`sname`\n" +
		"FROM `stations`\n" +
		"WHERE `stations`.`id` IN\n" +
		"(\n" +
		"SELECT `stations_for_routes`.`station_id`\n" +
		"FROM `routes`, `stations_for_routes`\n" +
		"WHERE (`routes`.`id` BETWEEN ? AND ?) AND (`routes`.`id` = `stations_for_routes`.`route_id`)\n" +
		")"
	if sql != expected {
		t.Errorf("EXP:\n%s\nGOT:\n%s\n", expected, sql)
	}

	sql2, err := typicalQuery.toSqlForRelation(typicalQuery.tables[0])
	if err == nil {
		t.Errorf("Expected error, but got query: %s", sql2)
	}
}

func TestToSqlForCombinedRows(t *testing.T) {
	sql := typicalQuery.toSqlForCombinedRows()
	expected := "SELECT `routes`.`id` AS `routes.id`, `routes`.`name` AS `routes.name`, "
	expected += "`stations`.`id` AS `stations.id`, `stations`.`sname` AS `stations.sname`, "
	expected += "`stations_for_routes`.`station_id` AS `stations_for_routes.station_id`, "
	expected += "`stations_for_routes`.`route_id` AS `stations_for_routes.route_id`, "
	expected += "`stations_for_routes`.`ord` AS `stations_for_routes.ord`\n"
	expected += "FROM `routes`, `stations`, `stations_for_routes`\n"
	expected += "WHERE (`routes`.`id` = `stations_for_routes`.`route_id`) AND "
	expected += "(`stations`.`id` = `stations_for_routes`.`station_id`) AND (`routes`.`id` BETWEEN ? AND ?)"
	if sql != expected {
		t.Errorf("EXP:\n%s\nGOT:\n%s\n", expected, sql)
	}
}

func TestToDDLError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	mock.ExpectQuery("DESCRIBE `routes`").
		WillReturnError(fmt.Errorf("Some error"))

	_, err = typicalQuery.toDDL(sqlxDB)
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestToDDLError2(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	mock.ExpectQuery("DESCRIBE `routes`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("OTHER_FIELD", "bigint(20)", "NO", "PRI", nil, ""),
		)

	_, err = typicalQuery.toDDL(sqlxDB)
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestMakeDDLFromTableDescriptionError(t *testing.T) {
	_, err := makeDDLFromTableDescription("", []TableColumnDDL{}, []string{}, []*QueryRelation{})
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestMakeDDLFromTableDescription(t *testing.T) {
	tableDescribtion := []TableColumnDDL{
		{"id", "bigint(20)", "NO", "PRI", sql.NullString{String: "", Valid: false}, ""},
		{"id2", "bigint(20)", "YES", "MUL", sql.NullString{String: "", Valid: true}, ""},
		{"id3", "bigint(20)", "YES", "UNI", sql.NullString{String: "0", Valid: true}, ""},
		{"id4", "varchar(255)", "YES", "PRI", sql.NullString{String: "", Valid: false}, ""},
	}

	relations := []*QueryRelation{
		{"some_table", "id2", "other_table", "id"},
	}

	ddl, err := makeDDLFromTableDescription("some_table", tableDescribtion, []string{"id", "id2", "id3", "no"}, relations)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	expectedDDL := "CREATE TABLE `some_table` (\n" +
		"    `id` bigint(20) NOT NULL,\n" +
		"    `id2` bigint(20) NULL DEFAULT '',\n" +
		"    `id3` bigint(20) NULL DEFAULT '0',\n" +
		"    PRIMARY KEY (`id`),\n" +
		"    INDEX `id2` (`id2`),\n" +
		"    UNIQUE INDEX `id3` (`id3`),\n" +
		"    CONSTRAINT `fk_id2` FOREIGN KEY (`id2`) REFERENCES `other_table` (`id`) ON DELETE CASCADE\n" +
		");"
	if ddl != expectedDDL {
		t.Errorf("Expected DDL\n%s\nGOT:\n%s\n", expectedDDL, ddl)
		return
	}
}

func TestTqlPartForSelectColumns(t *testing.T) {
	sql := typicalQuery.tables[0].sqlPartForSelectColumns()
	expected := "`routes`.`id`, `routes`.`name`"
	if sql != expected {
		t.Errorf("EXPECTED '%s' GOT '%s'", expected, sql)
	}
}

func TestSqlTable(t *testing.T) {
	sql := sqlTable("some_table")
	expected := "`some_table`"
	if sql != expected {
		t.Errorf("EXPECTED '%s' GOT '%s'", expected, sql)
	}
}

func TestSqlColumn(t *testing.T) {
	sql := sqlColumn("some_column")
	expected := "`some_column`"
	if sql != expected {
		t.Errorf("EXPECTED '%s' GOT '%s'", expected, sql)
	}
}

func TestSqlTableAndColumn(t *testing.T) {
	sql := sqlTableAndColumn("some_table", "some_column")
	expected := "`some_table`.`some_column`"
	if sql != expected {
		t.Errorf("EXPECTED '%s' GOT '%s'", expected, sql)
	}
}

type testToSqlSubQueryForRelationInput struct {
	query                   *Query
	mainTable               *QueryTable
	expectedSubquery        string
	expectedLeftTableColumn string
	expectedErr             bool
}

var testsToSqlSubQueryForRelation = []testToSqlSubQueryForRelationInput{
	{
		query:     typicalQuery,
		mainTable: &QueryTable{"stations", []string{"id", "sname"}},
		expectedSubquery: "SELECT `stations_for_routes`.`station_id`\n" +
			"FROM `routes`, `stations`, `stations_for_routes`\n" +
			"WHERE (`routes`.`id` BETWEEN ? AND ?) AND (`routes`.`id` = `stations_for_routes`.`route_id`)",
		expectedLeftTableColumn: "`stations`.`id`",
	},
	{
		query: &Query{
			tables: []*QueryTable{
				{"routes", []string{"id", "name"}},
				{"stations", []string{"id", "sname"}},
				{"stations_for_routes", []string{"station_id", "route_id", "ord"}},
			},
			relations: []*QueryRelation{
				{"routes", "id", "stations_for_routes", "route_id"},
				{"stations_for_routes", "station_id", "stations", "id"}, // inverted
			},
			primaryInterval: []int64{1000, 2000},
		},
		mainTable: &QueryTable{"stations", []string{"id", "sname"}},
		expectedSubquery: "SELECT `stations_for_routes`.`station_id`\n" +
			"FROM `routes`, `stations`, `stations_for_routes`\n" +
			"WHERE (`routes`.`id` BETWEEN ? AND ?) AND (`routes`.`id` = `stations_for_routes`.`route_id`)",
		expectedLeftTableColumn: "`stations`.`id`",
	},
	{
		query:       typicalQuery,
		mainTable:   &QueryTable{"routes", []string{"id", "name"}},
		expectedErr: true,
	},
	{
		query:       typicalQuery,
		mainTable:   &QueryTable{"people", []string{"id"}},
		expectedErr: true,
	},
}

func TestToSqlSubQueryForRelation(t *testing.T) {
	for _, input := range testsToSqlSubQueryForRelation {
		subquery, leftTableColumn, err := input.query.toSqlSubQueryForRelation(input.mainTable)
		if err != nil && !input.expectedErr {
			t.Errorf("Unexpected error: %s", err)
			continue
		}
		if subquery != input.expectedSubquery || leftTableColumn != input.expectedLeftTableColumn {
			t.Errorf("FOR %s WITH MAIN TABLE %s \n EXP %s WITH %s\n GOT %s WITH %s\n",
				convertQueryToString(input.query),
				input.mainTable.name,
				input.expectedSubquery,
				input.expectedLeftTableColumn,
				subquery,
				leftTableColumn,
			)
		}
	}
}

func TestFindRelationError(t *testing.T) {
	relations := []*QueryRelation{
		{"some_table", "id2", "other_table", "id"},
	}
	_, _, err := findRelation(relations, "other_table", "other_column")
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestConSetDns(t *testing.T) {
	conset := &ConnectionSettings{
		driver:   "mysql",
		user:     "user",
		password: "pass",
		dbname:   "dbname",
		dbhost:   "host",
	}
	dsn := conset.dsn()
	expected := "user:pass@tcp(host)/dbname"
	if dsn != expected {
		t.Errorf("EXPECTED '%s' GOT '%s'", expected, dsn)
	}
}

func convertQueryToString(q *Query) string {
	return fmt.Sprintf("%s\n%v\n%s\n",
		convertQtsToString(q.tables),
		q.primaryInterval,
		convertQrsToString(q.relations),
	)
}

package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"reflect"
	"testing"
)

var typicalQuery = &Query{
	tables: []*QueryTable{
		&QueryTable{"routes", []string{"id", "name"}},
		&QueryTable{"stations", []string{"id", "sname"}},
		&QueryTable{"stations_for_routes", []string{"station_id", "route_id", "ord"}},
	},
	relations: []*QueryRelation{
		{"routes", "id", "stations_for_routes", "route_id"},
		{"stations", "id", "stations_for_routes", "station_id"},
	},
	primaryInterval: []int64{1000, 2000},
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

	mock.ExpectQuery("SELECT (.+) FROM `routes` WHERE `routes`.`id` BETWEEN \\? AND \\?").
		WithArgs(1000, 2000).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	mock.ExpectQuery("SELECT (.+) FROM `stations` WHERE `stations`.`id` IN (.+)").
		WithArgs(1000, 2000).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	mock.ExpectQuery("SELECT (.+) FROM `stations_for_routes` WHERE `stations_for_routes`.`route_id` IN (.+)").
		WithArgs(1000, 2000).
		WillReturnRows(sqlmock.NewRows([]string{"station_id", "route_id", "ord"}))

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

	err = typicalQuery.QueryResult(dbConnectMock, &ConnectionSettings{}, &SimpleWriter{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
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
		&map[string]interface{}{"id": 1, "name": "name1"},
		&map[string]interface{}{"id": 2, "name": "name2"},
	}

	if !reflect.DeepEqual(expectedResultsMaps, resultsMaps) {
		t.Errorf("Unexpected results")
		return
	}

	mock.ExpectQuery("SELECT (.+) FROM some_table WHERE id BETWEEN \\? AND \\?").
		WithArgs(10, 20).WillReturnError(fmt.Errorf("Some error"))

	_, err = dbSelect(sqlxDB, "SELECT id, name FROM some_table WHERE id BETWEEN ? AND ?", 10, 20)
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
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
				&QueryTable{"routes", []string{"id", "name"}},
				&QueryTable{"stations", []string{"id", "sname"}},
				&QueryTable{"stations_for_routes", []string{"station_id", "route_id", "ord"}},
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

func convertQueryToString(q *Query) string {
	return fmt.Sprintf("%s\n%v\n%s\n",
		convertQtsToString(q.tables),
		q.primaryInterval,
		convertQrsToString(q.relations),
	)
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

func TestTqlPartForSelectColumns(t *testing.T) {
	sql := typicalQuery.tables[0].sqlPartForSelectColumns()
	expected := "`routes`.`id`, `routes`.`name`"
	if sql != expected {
		t.Errorf("EXPECTED '%s' GOT '%s'", expected, sql)
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

func TestToSqlForSingleTable(t *testing.T) {
	sql, _ := typicalQuery.toSqlForSingleTable(typicalQuery.tables[0])
	expected := "SELECT `routes`.`id`, `routes`.`name`\n" +
		"FROM `routes`\n" +
		"WHERE `routes`.`id` BETWEEN ? AND ?"
	if sql != expected {
		t.Errorf("EXP:\n%s\nGOT:\n%s\n", expected, sql)
	}
}

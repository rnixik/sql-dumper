package main

import (
	"fmt"
	"testing"
)

type testToSqlSubQueryForRelationInput struct {
	query                   *Query
	mainTable               *QueryTable
	expectedSubquery        string
	expectedLeftTableColumn string
	expectedErr             bool
}

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

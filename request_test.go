package main

import (
	"fmt"
	"reflect"
	"testing"
)

// Tables

type testTableTripl struct {
	tablesPart  string
	expected    []*QueryTable
	expectedErr bool
}

var testsTable = []testTableTripl{
	{
		tablesPart: "routes:id,name;stations:id,sname;stations_for_routes:station_id,route_id,ord",
		expected: []*QueryTable{
			&QueryTable{"routes", []string{"id", "name"}},
			&QueryTable{"stations", []string{"id", "sname"}},
			&QueryTable{"stations_for_routes", []string{"station_id", "route_id", "ord"}},
		},
	},
	{
		tablesPart: "routes:id,name",
		expected: []*QueryTable{
			&QueryTable{"routes", []string{"id", "name"}},
		},
	},
	{
		tablesPart:  "routes",
		expectedErr: true,
	},
	{
		tablesPart:  "",
		expectedErr: true,
	},
}

func TestParseTablesPart(t *testing.T) {
	for _, tripl := range testsTable {
		tables, err := parseTablesPart(tripl.tablesPart)
		if err != nil && !tripl.expectedErr {
			t.Errorf("Unexpected error: %s", err)
			continue
		}
		if !reflect.DeepEqual(tables, tripl.expected) {
			t.Errorf("FOR %s\n EXP %s\n GOT %s\n",
				tripl.tablesPart,
				convertQtsToString(tripl.expected),
				convertQtsToString(tables),
			)
		}
	}
}

func convertQtsToString(qts []*QueryTable) string {
	str := ""
	for _, qt := range qts {
		str += fmt.Sprintf("%+v", qt)
	}
	return str
}

// Intervals

type testIntervalsTripl struct {
	intervalsPart string
	expected      []int64
	expectedErr   bool
}

var testsIntervals = []testIntervalsTripl{
	{
		intervalsPart: "704293046165300-704293046165399",
		expected:      []int64{704293046165300, 704293046165399},
	},
	{
		intervalsPart: "1-2",
		expected:      []int64{1, 2},
	},
	{
		intervalsPart: "1-a",
		expectedErr:   true,
	},
	{
		intervalsPart: "1",
		expectedErr:   true,
	},
	{
		intervalsPart: "",
		expectedErr:   true,
	},
}

func TestParseIntervalPart(t *testing.T) {
	for _, tripl := range testsIntervals {
		intervals, err := parseIntervalPart(tripl.intervalsPart)
		if err != nil && !tripl.expectedErr {
			t.Errorf("Unexpected error: %s", err)
			continue
		}
		if !reflect.DeepEqual(intervals, tripl.expected) {
			t.Errorf("FOR %s EXP %v GOT %v\n",
				tripl.intervalsPart,
				tripl.expected,
				intervals,
			)
		}
	}
}

// Relations

type testRelationsTripl struct {
	relationsPart string
	expected      []*QueryRelation
	expectedErr   bool
}

var testsRelations = []testRelationsTripl{
	{
		relationsPart: "routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id",
		expected: []*QueryRelation{
			{"routes", "id", "stations_for_routes", "route_id"},
			{"stations", "id", "stations_for_routes", "station_id"},
		},
	},
	{
		relationsPart: "routes.id=stations_for_routes.route_id",
		expected: []*QueryRelation{
			{"routes", "id", "stations_for_routes", "route_id"},
		},
	},
	{
		relationsPart: "",
		expected:      []*QueryRelation{},
	},
	{
		relationsPart: "routes.id=stations_for_routes.",
		expectedErr:   true,
	},
	{
		relationsPart: "asd",
		expectedErr:   true,
	},
}

func TestParseRelationsPart(t *testing.T) {
	for _, tripl := range testsRelations {
		relations, err := parseRelationsPart(tripl.relationsPart)
		if err != nil && !tripl.expectedErr {
			t.Errorf("Unexpected error: %s", err)
			continue
		}
		if !reflect.DeepEqual(relations, tripl.expected) {
			t.Errorf("FOR %s\n EXP %s\n GOT %s\n",
				tripl.relationsPart,
				convertQrsToString(tripl.expected),
				convertQrsToString(relations),
			)
		}
	}
}

func convertQrsToString(qrs []*QueryRelation) string {
	str := ""
	for _, qt := range qrs {
		str += fmt.Sprintf("%+v", qt)
	}
	return str
}

// Query

type testQueryInput struct {
	tablesPart    string
	relationsPart string
	intervalsPart string
	expected      *Query
	expectedErr   bool
}

var testsQuery = []testQueryInput{
	{
		tablesPart:    "routes:id,name;stations:id,sname;stations_for_routes:station_id,route_id,ord",
		intervalsPart: "154293032165394-154293032165399",
		relationsPart: "routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id",
		expected: &Query{
			tables: []*QueryTable{
				&QueryTable{"routes", []string{"id", "name"}},
				&QueryTable{"stations", []string{"id", "sname"}},
				&QueryTable{"stations_for_routes", []string{"station_id", "route_id", "ord"}},
			},
			relations: []*QueryRelation{
				{"routes", "id", "stations_for_routes", "route_id"},
				{"stations", "id", "stations_for_routes", "station_id"},
			},
			primaryInterval: []int64{154293032165394, 154293032165399},
		},
	},
	{
		tablesPart:    "routes:id,name;stations:id,sname;stations_for_routes:station_id,route_id,ord",
		intervalsPart: "154293032165394-154293032165399",
		relationsPart: "asd",
		expectedErr:   true,
	},
	{
		tablesPart:    "routes:id,name;stations:id,sname;stations_for_routes:station_id,route_id,ord",
		intervalsPart: "asd",
		relationsPart: "routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id",
		expectedErr:   true,
	},
	{
		tablesPart:    "asd",
		intervalsPart: "154293032165394-154293032165399",
		relationsPart: "routes.id=stations_for_routes.route_id;stations.id=stations_for_routes.station_id",
		expectedErr:   true,
	},
}

func TestParseRequest(t *testing.T) {
	for _, input := range testsQuery {
		query, err := ParseRequest(input.tablesPart, input.intervalsPart, input.relationsPart)
		if err != nil && !input.expectedErr {
			t.Errorf("Unexpected error: %s", err)
			continue
		}
		if !reflect.DeepEqual(query, input.expected) {
			t.Errorf("FOR %s\n%s\n%s\n EXP %s\n%v\n%s\n GOT %s\n%v\n%s\n",
				input.tablesPart,
				input.intervalsPart,
				input.relationsPart,
				convertQtsToString(input.expected.tables),
				input.expected.primaryInterval,
				convertQrsToString(input.expected.relations),
				convertQtsToString(query.tables),
				query.primaryInterval,
				convertQrsToString(query.relations),
			)
		}
	}
}

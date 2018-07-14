package main

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseRequest parses input strings into query definitions
func ParseRequest(tablesPart string, intervalPart string, relationsPart string) (query *Query, err error) {
	tables, err := parseTablesPart(tablesPart)
	if err != nil {
		return nil, err
	}
	interval, err := parseIntervalPart(intervalPart)
	if err != nil {
		return nil, err
	}
	relations, err := parseRelationsPart(relationsPart)
	if err != nil {
		return nil, err
	}
	return &Query{
		tables:          tables,
		relations:       relations,
		primaryInterval: interval,
	}, nil
}

func parseTablesPart(tablesPart string) (tables []*QueryTable, err error) {
	if tablesPart == "" {
		return nil, fmt.Errorf("Tables part is empty")
	}
	tables = make([]*QueryTable, 0)
	tablesDefinitions := strings.Split(tablesPart, ";")
	for _, tableDefinition := range tablesDefinitions {
		tableDefinitionParts := strings.Split(tableDefinition, ":")
		if len(tableDefinitionParts) != 2 {
			return nil, fmt.Errorf("Table definition should be in format 'table:column1,column2,...' Got %s", tableDefinition)
		}
		if tableDefinitionParts[1] == "" {
			return nil, fmt.Errorf("Table definition should contain one column at least. Got %s", tableDefinition)
		}
		tableName := tableDefinitionParts[0]
		columns := strings.Split(tableDefinitionParts[1], ",")
		queryTable := &QueryTable{tableName, columns}
		tables = append(tables, queryTable)
	}
	return tables, nil
}

func parseIntervalPart(intervalPart string) (interval []int64, err error) {
	interval = make([]int64, 2)
	intervalParts := strings.Split(intervalPart, "-")
	if len(intervalParts) != 2 {
		return nil, fmt.Errorf("Interval definition should be in format 'start-end'. Got: %s", intervalPart)
	}
	startValue, err := strconv.ParseInt(intervalParts[0], 10, 64)
	if err != nil {
		return nil, err
	}
	endValue, err := strconv.ParseInt(intervalParts[1], 10, 64)
	if err != nil {
		return nil, err
	}
	interval[0] = startValue
	interval[1] = endValue
	return interval, err
}

func parseRelationsPart(relationsPart string) (relations []*QueryRelation, err error) {
	relations = make([]*QueryRelation, 0)
	if len(relationsPart) == 0 {
		return relations, nil
	}
	relationsDefinitions := strings.Split(relationsPart, ";")
	for _, relationDefinition := range relationsDefinitions {
		bothSides := strings.Split(relationDefinition, "=")
		if len(bothSides) != 2 {
			return nil, fmt.Errorf("Relation definition should in format 'table1.column1=table2.column2'. Got %s", relationDefinition)
		}
		side1 := bothSides[0]
		side2 := bothSides[1]
		side1Parts := strings.Split(side1, ".")
		side2Parts := strings.Split(side2, ".")
		if len(side1Parts) != 2 || len(side2Parts) != 2 {
			return nil, fmt.Errorf("Relation definition should in format 'table1.column1=table2.column2'. Got %s", relationDefinition)
		}
		if side1Parts[0] == "" || side1Parts[1] == "" || side2Parts[0] == "" || side2Parts[1] == "" {
			return nil, fmt.Errorf("Found empty relation part: table or column")
		}
		queryRelation := QueryRelation{side1Parts[0], side1Parts[1], side2Parts[0], side2Parts[1]}
		relations = append(relations, &queryRelation)
	}
	return relations, nil
}

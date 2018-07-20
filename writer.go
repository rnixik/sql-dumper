package main

import (
	"fmt"
)

type DataWriter interface {
	WriteRows(tableName string, columns []string, results []*map[string]interface{}) (err error)
	WriteDDL(tableName string, ddl string) (err error)
}

type SimpleWriter struct {
}

func (w *SimpleWriter) WriteRows(tableName string, _ []string, rows []*map[string]interface{}) (err error) {
	fmt.Println(tableName)
	for _, row := range rows {
		for field, v := range *row {
			value := ""
			switch typedValue := v.(type) {
			case int:
				value = fmt.Sprintf("%d", typedValue)
				break
			case int64:
				value = fmt.Sprintf("%d", typedValue)
				break
			case float64:
				value = fmt.Sprintf("%f", typedValue)
				break
			case string:
				value = typedValue
				break
			case []uint8:
				value = fmt.Sprintf("%s", typedValue)
			case nil:
				value = "NULL"
			default:
				value = "UNDEFINED"
			}
			fmt.Printf("%s = %s;||", field, value)
		}
		fmt.Println("")
	}
	return nil
}

func (w *SimpleWriter) WriteDDL(tableName string, ddl string) (err error) {
	fmt.Println(ddl)
	return nil
}

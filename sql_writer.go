package main

import (
	"fmt"
	"strings"
)

type SqlWriter struct {
	dstFile string
	dstDir  string
}

func (w *SqlWriter) WriteDDL(tableName string, ddl string) (err error) {
	fmt.Println("SET FOREIGN_KEY_CHECKS=0;")
	fmt.Println(ddl)
	fmt.Println("SET FOREIGN_KEY_CHECKS=1;")
	return nil
}

func (w *SqlWriter) WriteRows(tableName string, columns []string, rows []*map[string]interface{}) (err error) {
	columnsNames := make([]string, 0)
	for _, column := range columns {
		columnsNames = append(columnsNames, "`"+column+"`")
	}
	for _, row := range rows {
		values := make([]string, 0)
		for _, field := range columns {
			v := (*row)[field]
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
				value = escapeString(typedValue)
				break
			case []uint8:
				value = escapeString(fmt.Sprintf("%s", typedValue))
			case nil:
				value = "NULL"
			default:
				value = "UNDEFINED"
			}
			values = append(values, value)
		}
		insert := "INSERT INTO `" + tableName + "` (" + strings.Join(columnsNames, ", ") + ") " +
			"VALUES (" + strings.Join(values, ", ") + ");"
		fmt.Println(insert)
	}
	return nil
}

func escapeString(str string) string {
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "'", "\\'", -1)
	return "'" + str + "'"
}

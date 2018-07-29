package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"os"
	"reflect"
	"testing"
)

func TestRunErrorArguments(t *testing.T) {
	dbConnect := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return nil, nil
	}
	err := Run(dbConnect, []string{}, "", "", NewTestFileWriter(), "", "", ",")
	if err != nil {
		t.Errorf("Expected help, but got error: %s", err)
		return
	}
}

func TestRun(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	dbConnectMock := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return sqlxDB, nil
	}

	mock.ExpectQuery("DESCRIBE `some_table`").
		WillReturnRows(
			sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).AddRow("id", "bigint(20)", "NO", "PRI", nil, ""),
		)

	mock.ExpectQuery("SELECT (.+) FROM `some_table` (.+)").
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	err = Run(dbConnectMock, []string{"some_table:id", "1-2"}, ".env.example", "sql", NewTestFileWriter(), "test_example.sql", "", ",")
	os.Setenv("DB_NAME", "")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
}

func TestRunDbError(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	dbConnectMock := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return sqlxDB, nil
	}

	mock.ExpectQuery("DESCRIBE `some_table`").WillReturnError(fmt.Errorf("Some DB error"))

	err = Run(dbConnectMock, []string{"some_table:id", "1-2"}, ".env.example", "sql", NewTestFileWriter(), "test_example.sql", "", ",")
	os.Setenv("DB_NAME", "")
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestRunConfigReadError(t *testing.T) {
	dbConnect := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return nil, nil
	}
	os.Setenv("DB_NAME", "")
	err := Run(dbConnect, []string{"some_table:id", "1-2"}, "not_existing_file", "", NewTestFileWriter(), "", "", ",")
	os.Setenv("DB_NAME", "")
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestRunParseError(t *testing.T) {
	dbConnect := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return nil, nil
	}
	os.Setenv("DB_NAME", "")
	err := Run(dbConnect, []string{"", "", ""}, ".env.example", "", NewTestFileWriter(), "", "", ",")
	os.Setenv("DB_NAME", "")
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestGetConnectionSettingsFileError(t *testing.T) {
	os.Setenv("DB_NAME", "")
	_, err := getConnectionSettings("not_existing_file")
	if err == nil {
		t.Errorf("Expected error, but got nil")
		return
	}
}

func TestGetWriterAndCombinedMode_sqlDefault(t *testing.T) {
	format := "sql"
	fw := NewTestFileWriter()
	dstFile := ""
	dstDir := ""
	csvDelimiter := ","
	writer, combined := getWriterAndCombinedMode(format, fw, dstFile, dstDir, csvDelimiter)
	assert.IsType(t, &SqlWriter{}, writer)
	sqlWriter := writer.(*SqlWriter)
	assert.Equal(t, fw, sqlWriter.fw)
	assert.Equal(t, "result.sql", sqlWriter.dstFile)
	assert.Equal(t, "", sqlWriter.dstDir)
	assert.False(t, combined)
}

func TestGetWriterAndCombinedMode_sqlDstFile(t *testing.T) {
	format := "sql"
	fw := NewTestFileWriter()
	dstFile := "custom.sql"
	dstDir := ""
	csvDelimiter := ","
	writer, combined := getWriterAndCombinedMode(format, fw, dstFile, dstDir, csvDelimiter)
	assert.IsType(t, &SqlWriter{}, writer)
	sqlWriter := writer.(*SqlWriter)
	assert.Equal(t, "custom.sql", sqlWriter.dstFile)
	assert.Equal(t, "", sqlWriter.dstDir)
	assert.False(t, combined)
}

func TestGetWriterAndCombinedMode_sqlDstDir(t *testing.T) {
	format := "sql"
	fw := NewTestFileWriter()
	dstFile := ""
	dstDir := "./"
	csvDelimiter := ","
	writer, combined := getWriterAndCombinedMode(format, fw, dstFile, dstDir, csvDelimiter)
	assert.IsType(t, &SqlWriter{}, writer)
	sqlWriter := writer.(*SqlWriter)
	assert.Equal(t, "", sqlWriter.dstFile)
	assert.Equal(t, "./", sqlWriter.dstDir)
	assert.False(t, combined)
}

func TestGetWriterAndCombinedMode_csvDefault(t *testing.T) {
	format := "csv"
	fw := NewTestFileWriter()
	dstFile := ""
	dstDir := ""
	csvDelimiter := ","
	writer, combined := getWriterAndCombinedMode(format, fw, dstFile, dstDir, csvDelimiter)
	assert.IsType(t, &CsvWriter{}, writer)
	csvWriter := writer.(*CsvWriter)
	assert.Equal(t, fw, csvWriter.fw)
	assert.Equal(t, "result.csv", csvWriter.dstFile)
	assert.Equal(t, "", csvWriter.dstDir)
	assert.True(t, combined)
}

func TestGetWriterAndCombinedMode_csvDstFile(t *testing.T) {
	format := "csv"
	fw := NewTestFileWriter()
	dstFile := "custom.csv"
	dstDir := ""
	csvDelimiter := ","
	writer, combined := getWriterAndCombinedMode(format, fw, dstFile, dstDir, csvDelimiter)
	assert.IsType(t, &CsvWriter{}, writer)
	csvWriter := writer.(*CsvWriter)
	assert.Equal(t, fw, csvWriter.fw)
	assert.Equal(t, "custom.csv", csvWriter.dstFile)
	assert.Equal(t, "", csvWriter.dstDir)
	assert.True(t, combined)
}

func TestGetWriterAndCombinedMode_csvDstDir(t *testing.T) {
	format := "csv"
	fw := NewTestFileWriter()
	dstFile := ""
	dstDir := "./"
	csvDelimiter := ","
	writer, combined := getWriterAndCombinedMode(format, fw, dstFile, dstDir, csvDelimiter)
	assert.IsType(t, &CsvWriter{}, writer)
	csvWriter := writer.(*CsvWriter)
	assert.Equal(t, fw, csvWriter.fw)
	assert.Equal(t, "", csvWriter.dstFile)
	assert.Equal(t, "./", csvWriter.dstDir)
	assert.False(t, combined)
}

func TestGetWriterAndCombinedMode_simple(t *testing.T) {
	format := "simple"
	fw := NewTestFileWriter()
	dstFile := ""
	dstDir := ""
	csvDelimiter := ","
	writer, combined := getWriterAndCombinedMode(format, fw, dstFile, dstDir, csvDelimiter)
	assert.IsType(t, &SimpleWriter{}, writer)
	assert.True(t, combined)
}

func TestGetConnectionSettings(t *testing.T) {
	os.Setenv("DB_NAME", "")
	conset, err := getConnectionSettings(".env.example")
	os.Setenv("DB_NAME", "")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	fields := map[string]string{
		"driver":   "mysql",
		"user":     "root",
		"password": "root",
		"dbname":   "test",
		"dbhost":   "127.0.0.1",
	}
	for key, val := range fields {
		r := reflect.ValueOf(conset)
		f := reflect.Indirect(r).FieldByName(key)
		actualValue := f.String()
		if actualValue != val {
			t.Errorf("Expected for %s: %s, got: %s", key, val, actualValue)
			return
		}
	}
}

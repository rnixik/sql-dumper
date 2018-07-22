package main

import (
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestRunErrorArguments(t *testing.T) {
	dbConnect := func(conset *ConnectionSettings) (db *sqlx.DB, err error) {
		return nil, nil
	}
	err := Run(dbConnect, []string{}, "", "", NewTestFileWriter(), "", "")
	if err != nil {
		t.Errorf("Expected help, but got err: %s", err)
		return
	}
}

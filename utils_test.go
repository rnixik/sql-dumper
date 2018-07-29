package main

import (
	"testing"
)

type TestContainsDef struct {
	haystack []string
	needle   string
	result   bool
}

var testsContains = []TestContainsDef{
	{[]string{"abc", "def", "ghi"}, "def", true},
	{[]string{"abc", "def", "ghi"}, "jkl", false},
	{[]string{"abc", "def", "ghi"}, "", false},
	{[]string{}, "abc", false},
	{[]string{}, "", false},
}

func TestContains(t *testing.T) {
	for _, testDef := range testsContains {
		if contains(testDef.haystack, testDef.needle) != testDef.result {
			t.Errorf("FOR %s\n IN %v EXPECTED %v\n",
				testDef.needle,
				testDef.haystack,
				testDef.result,
			)
		}
	}
}

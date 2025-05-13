package main

import (
	"fmt"
	"testing"
)

func Test_getAnalyzers(t *testing.T) {
	analyzers := getAnalyzers()

	expected := []string{
		"SA1012", "SA4006", "SA5000", "SA6000", "SA9004",
		"printf", "shadow", "structtag", "unreachable", "unusedresult",
		"noosexit",
	}

	fmt.Println(analyzers)

	found := make(map[string]bool)
	for _, a := range analyzers {
		found[a.Name] = true
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("analyzer %s not found in the list", name)
		}
	}
}

package foldercleaner

import (
	"os"
	"testing"

	"gopkg.in/yaml.v2"
)

func createFolders(t *testing.T, folderNames []string) {
	for _, folderName := range folderNames {
		err := os.Mkdir(folderName, 0755)
		if err != nil && !os.IsExist(err) {
			t.Fatalf("Unable to create folder: %s: %s", folderName, err)
		}
	}
	t.Cleanup(func() {
		for _, folderName := range folderNames {
			os.RemoveAll(folderName)
		}
	})
}

func TestCleaner(t *testing.T) {
	// Setting up the test
	createFolders(t, []string{
		"/tmp/cleaner_test",
		"/tmp/cleaner_test/folderA",
		"/tmp/cleaner_test/folderB",
	})

	yamlConfig := `---
tasks:
  - path: /tmp/cleaner_test/folderA
    pattern: .*
    ttl: 10s
  - path: /tmp/cleaner_test/folderB
    ttl: 1m
    interval: 10s`
	config := Configuration{}
	if err := yaml.Unmarshal([]byte(yamlConfig), &config); err != nil {
		t.Fatalf("Unable to unmarshal the yaml config: %s", err)
	}
	// TODO: Actually create tests
}

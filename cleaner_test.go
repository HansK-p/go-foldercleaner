package foldercleaner

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
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
	yamlConfig := `---
tasks:
  - path: /tmp/cleaner_test/folderA
    pattern: .*
    ttl: 1s
  - path: /tmp/cleaner_test/folderB
    ttl: 1m
    interval: 2s`

	t.Log("Parsing the configuration")
	config := Configuration{}
	if err := yaml.Unmarshal([]byte(yamlConfig), &config); err != nil {
		t.Fatalf("Unable to unmarshal the yaml config: %s", err)
	}

	if len(config.Tasks) != 2 {
		t.Fatalf("Number of configured tasks should be 2, but was %d", len(config.Tasks))
	}

	t.Log("Creating the folder cleaner")
	logger := log.New().WithFields(log.Fields{})
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	folderCleaner, err := GetFolderCleaner(logger, ctx, wg, &config)
	if err != nil {
		t.Fatalf("Unable to get a the Folder Cleaner")
	}

	if alive, message := folderCleaner.IsAlive(); alive {
		t.Errorf("The folder cleaner is alive with message '%s' before it has been scheduled", message)
	}
	if ready, message := folderCleaner.IsReady(); ready {
		t.Errorf("The folder cleaner is ready with message '%s' before it has been scheduled", message)
	}

	t.Log("Scheduler folder cleaner tasks")
	if err := folderCleaner.Schedule(); err != nil {
		t.Fatalf("When scheduling cleaner jobs: %s", err)
	}

	if alive, message := folderCleaner.IsAlive(); !alive {
		t.Errorf("The folder cleaner is not alive with message '%s' after it has been scheduled", message)
	}
	if ready, message := folderCleaner.IsReady(); !ready {
		t.Errorf("The folder cleaner is not ready with message '%s' after it has been scheduled", message)
	}

	sleepTimeMS := 1500
	t.Logf("Wait %d ms for the folder cleaner to run and fail as target folders does not exist", sleepTimeMS)
	time.Sleep(time.Millisecond * time.Duration(sleepTimeMS))
	if alive, message := folderCleaner.IsAlive(); alive {
		t.Errorf("The folder cleaner is alive with message '%s' after a cleanup task has failed", message)
	}
	if ready, message := folderCleaner.IsReady(); ready {
		t.Errorf("The folder cleaner is ready with message '%s' after a cleanup task has failed", message)
	}

	t.Logf("Creating folders needed for cleanup tasks to succeed")
	createFolders(t, []string{
		"/tmp/cleaner_test",
		"/tmp/cleaner_test/folderA",
		"/tmp/cleaner_test/folderB",
	})

	t.Logf("Wait %d ms for the folder cleaner to run and succeed as target folders have been created", sleepTimeMS)
	time.Sleep(time.Millisecond * time.Duration(sleepTimeMS))
	if alive, message := folderCleaner.IsAlive(); !alive {
		t.Errorf("The folder cleaner is not alive with message '%s' after all cleanup tasks have succeeded", message)
	}
	if ready, message := folderCleaner.IsReady(); !ready {
		t.Errorf("The folder cleaner is not ready with message '%s' after all cleanup tasks have succeeded", message)
	}

	t.Logf("Controlled stop of cleaner jobs")
	cancel()
	wg.Wait()

	t.Logf("All done")
}

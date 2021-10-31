package foldercleaner

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	Tasks []CfgCleanerTask `yaml:"tasks"`
}

type FolderCleaner struct {
	logger *log.Entry
	ctx    context.Context
	wg     *sync.WaitGroup
	config *Configuration
	tasks  []*FolderCleanerTask
}

func GetFolderCleaner(logger *log.Entry, ctx context.Context, wg *sync.WaitGroup, config *Configuration) (*FolderCleaner, error) {
	tasks := []*FolderCleanerTask{}

	for idx := range config.Tasks {
		task := &config.Tasks[idx]
		logger := logger.WithFields(log.Fields{"CleanerTask": task})

		if cleanerTask, err := GetFolderCleanerTask(logger, ctx, wg, task); err != nil {
			logger.Errorf("When scheduling cleaner task: %s", err)
		} else {
			if err := cleanerTask.Schedule(); err != nil {
				logger.Errorf("When scheduling the cleaner: %s", err)
			}
			logger.Infof("Cleaner task scheduled")
		}
	}

	return &FolderCleaner{
		logger: logger.WithFields(log.Fields{"Package": "cleaner", "Module": "FolderCleaner", "TaskConfig": config}),
		ctx:    ctx,
		wg:     wg,
		config: config,
		tasks:  tasks,
	}, nil
}

func (fc *FolderCleaner) Schedule() error {
	logger := fc.logger.WithFields(log.Fields{"Function": "Schedule"})
	logger.Infof("Scheduling the folder cleaner")

	for _, task := range fc.tasks {
		logger := logger.WithFields(log.Fields{"CleanerTask": task})
		if err := task.Schedule(); err != nil {
			return fmt.Errorf("when scheduling task %#v: %s", *task, err)
		}
		logger.Info("Scheduled task")
	}
	return nil
}

package cleaner

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	Jobs []CfgCleanerJob `yaml:"jobs"`
}

type FolderCleaner struct {
	logger *log.Entry
	ctx    context.Context
	wg     *sync.WaitGroup
	config *Configuration
	jobs   []*FolderCleanerJob
}

func GetFolderCleaner(logger *log.Entry, ctx context.Context, wg *sync.WaitGroup, config *Configuration) (*FolderCleaner, error) {
	jobs := []*FolderCleanerJob{}

	for idx := range config.Jobs {
		job := &config.Jobs[idx]
		logger := logger.WithFields(log.Fields{"CleanerJob": job})

		if cleanerJob, err := GetFolderCleanerJob(logger, ctx, wg, job); err != nil {
			logger.Errorf("When scheduling cleaner job: %s", err)
		} else {
			if err := cleanerJob.Schedule(); err != nil {
				logger.Errorf("When scheduling the cleaner: %s", err)
			}
			logger.Infof("Cleaner job scheduled")
		}
	}

	return &FolderCleaner{
		logger: logger.WithFields(log.Fields{"Package": "cleaner", "Module": "FolderCleaner", "JobConfig": config}),
		ctx:    ctx,
		wg:     wg,
		config: config,
		jobs:   jobs,
	}, nil
}

func (fc *FolderCleaner) Schedule() error {
	logger := fc.logger.WithFields(log.Fields{"Function": "Schedule"})
	logger.Infof("Scheduling the folder cleaner")

	for _, job := range fc.jobs {
		logger := logger.WithFields(log.Fields{"CleanerJob": job})
		if err := job.Schedule(); err != nil {
			return fmt.Errorf("when scheduling job %#v: %s", *job, err)
		}
		logger.Info("Scheduled job")
	}
	return nil
}

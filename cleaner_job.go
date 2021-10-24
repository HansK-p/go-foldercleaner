package foldercleaner

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type CfgCleanerJob struct {
	Path    string        `yaml:"path"`
	Pattern string        `yaml:"pattern,omitempty"`
	TTL     time.Duration `yaml:"ttl,omitempty"`
}

type FolderCleanerJob struct {
	logger    *log.Entry
	ctx       context.Context
	wg        *sync.WaitGroup
	config    *CfgCleanerJob
	patternRE *regexp.Regexp `yaml:"pattern_re,omitempty"`
	interval  time.Duration
}

func GetFolderCleanerJob(logger *log.Entry, ctx context.Context, wg *sync.WaitGroup, config *CfgCleanerJob) (*FolderCleanerJob, error) {
	var patternRE *regexp.Regexp
	var err error

	if config.Pattern != "" {
		patternRE, err = regexp.Compile(config.Pattern)
		if err != nil {
			return nil, fmt.Errorf("converting the file pattern '%s' to a regular expression: %w", config.Pattern, err)
		}
	}

	return &FolderCleanerJob{
		logger:    logger.WithFields(log.Fields{"Package": "cleaner", "Module": "FolderCleanerJob", "JobConfig": config}),
		ctx:       ctx,
		wg:        wg,
		config:    config,
		patternRE: patternRE,
		interval:  config.TTL / 4,
	}, nil
}

func (fcj *FolderCleanerJob) clean() error {
	logger := fcj.logger.WithFields(log.Fields{"Function": "clean"})
	logger.Infof("Cleaning")

	files, err := ioutil.ReadDir(fcj.config.Path)
	if err != nil {
		return fmt.Errorf("when reading files in folder %s: %w", fcj.config.Path, err)
	}

	for _, file := range files {
		if file.Mode().IsRegular() && time.Since(file.ModTime()) > fcj.config.TTL && (fcj.patternRE == nil || fcj.patternRE.Match([]byte(file.Name()))) {
			logger := logger.WithFields(log.Fields{"Name": file.Name(), "Modified": file.ModTime()})
			logger.Infof("Deleting file as ttl was reached")
			path := filepath.Join(fcj.config.Path, file.Name())
			err = os.Remove(path)
			if err != nil {
				return fmt.Errorf("when deleting file %s: %w", path, err)
			}
		}
	}
	return nil
}

func (fcj *FolderCleanerJob) scheduleCleaner() {
	logger := fcj.logger.WithFields(log.Fields{"Function": "scheduleCleaner"})
	ticker := time.NewTicker(fcj.interval)
	defer ticker.Stop()
	fcj.wg.Add(1)
	defer fcj.wg.Done()
	for {
		err := fcj.clean()
		if err != nil {
			logger.Errorf("While cleaning folder: %s", err)
		}
		select {
		case <-ticker.C:
		case <-fcj.ctx.Done():
			logger.Info("Cleaner stopped")
			return
		}
	}
}

func (fcj *FolderCleanerJob) Schedule() error {
	fcj.logger.Infof("Scheduling the folder cleaner")
	go fcj.scheduleCleaner()
	return nil
}

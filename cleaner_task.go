package foldercleaner

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/HansK-p/go-customtypes"
	log "github.com/sirupsen/logrus"
)

type CfgCleanerTask struct {
	Path     string              `yaml:"path"`
	Pattern  *customtypes.Regexp `yaml:"pattern,omitempty"`
	TTL      time.Duration       `yaml:"ttl,omitempty"`
	Interval time.Duration       `yaml:"interval,omitempty"`
}

type FolderCleanerTask struct {
	logger *log.Entry
	ctx    context.Context
	wg     *sync.WaitGroup
	config *CfgCleanerTask

	healthDescription string
	isHealthy         bool
}

func GetFolderCleanerTask(logger *log.Entry, ctx context.Context, wg *sync.WaitGroup, config *CfgCleanerTask) (*FolderCleanerTask, error) {
	return &FolderCleanerTask{
		logger: logger.WithFields(log.Fields{"Package": "cleaner", "Module": "FolderCleanerTask", "TaskConfig": config}),
		ctx:    ctx,
		wg:     wg,
		config: config,

		healthDescription: "Not yet started",
		isHealthy:         false,
	}, nil
}

func (fct *FolderCleanerTask) clean() error {
	logger := fct.logger.WithFields(log.Fields{"Function": "clean"})
	logger.Infof("Cleaning")

	files, err := ioutil.ReadDir(fct.config.Path)
	if err != nil {
		return fmt.Errorf("when reading files in folder %s: %w", fct.config.Path, err)
	}

	for _, file := range files {
		if file.Mode().IsRegular() && time.Since(file.ModTime()) > fct.config.TTL && (fct.config.Pattern == nil || fct.config.Pattern.Match([]byte(file.Name()))) {
			logger := logger.WithFields(log.Fields{"Name": file.Name(), "Modified": file.ModTime()})
			logger.Infof("Deleting file as ttl was reached")
			path := filepath.Join(fct.config.Path, file.Name())
			err = os.Remove(path)
			if err != nil {
				return fmt.Errorf("when deleting file %s: %w", path, err)
			}
		}
	}
	return nil
}

func (fct *FolderCleanerTask) scheduleCleaner() {
	logger := fct.logger.WithFields(log.Fields{"Function": "scheduleCleaner"})
	interval := fct.config.Interval
	if interval.Milliseconds() == 0 {
		interval = fct.config.TTL / 4
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	fct.wg.Add(1)
	defer fct.wg.Done()
	for {
		err := fct.clean()
		if err != nil {
			logger.Errorf("While cleaning folder: %s", err)
			fct.setHealthStatus(false, "Error cleaning folder")
		} else {
			fct.setHealthStatus(true, "Folder cleaned")
		}
		select {
		case <-ticker.C:
		case <-fct.ctx.Done():
			logger.Info("Cleaner stopped")
			return
		}
	}
}

func (fct *FolderCleanerTask) Schedule() error {
	fct.logger.Infof("Scheduling the folder cleaner")
	go fct.scheduleCleaner()
	fct.setHealthStatus(true, "Cleaner task started")
	return nil
}

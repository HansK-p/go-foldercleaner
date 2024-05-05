package foldercleaner

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/HansK-p/go-customtypes"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type CfgCleanerTask struct {
	Path      string              `yaml:"path"`
	Pattern   *customtypes.Regexp `yaml:"pattern,omitempty"`
	TTL       time.Duration       `yaml:"ttl"`
	Interval  time.Duration       `yaml:"interval,omitempty"`
	Recursive bool                `yaml:"recursive"`
}

type FolderCleanerTask struct {
	logger *log.Entry
	ctx    context.Context
	wg     *sync.WaitGroup
	config *CfgCleanerTask

	labels prometheus.Labels

	healthDescription string
	isHealthy         bool
}

func NewFolderCleanerTask(logger *log.Entry, ctx context.Context, wg *sync.WaitGroup, config *CfgCleanerTask) (folderCleanerTask *FolderCleanerTask, err error) {
	folderCleanerTask = &FolderCleanerTask{
		logger:            logger.WithFields(log.Fields{"Package": "cleaner", "Module": "FolderCleanerTask", "TaskConfig": config}),
		ctx:               ctx,
		wg:                wg,
		config:            config,
		healthDescription: "Not yet started",
		isHealthy:         false,
	}

	if config.Pattern != nil {
		folderCleanerTask.labels = prometheus.Labels{"path": config.Path, "pattern": config.Pattern.String()}
	} else {
		folderCleanerTask.labels = prometheus.Labels{"path": config.Path, "pattern": ""}
	}

	promFilesRemoved.With(folderCleanerTask.labels)
	promFileCleanFailures.With(folderCleanerTask.labels)
	promFileRemoveFailures.With(folderCleanerTask.labels)
	return
}

func (fct *FolderCleanerTask) conditionallyRemove(fileInfo fs.FileInfo, path string) (bool, error) {
	logger := fct.logger.WithFields(log.Fields{"Function": "conditionallyRemove", "Name": fileInfo.Name(), "Modified": fileInfo.ModTime()})
	if fileInfo.Mode().IsRegular() && time.Since(fileInfo.ModTime()) > fct.config.TTL && (fct.config.Pattern == nil || fct.config.Pattern.Match([]byte(fileInfo.Name()))) {
		logger.Infof("Deleting file as ttl was reached")
		if err := os.Remove(path); err != nil {
			promFileRemoveFailures.With(fct.labels).Add(1)
			return true, fmt.Errorf("when deleting file %s: %w", path, err)
		}
		promFilesRemoved.With(fct.labels).Add(1)
		return true, nil
	}
	return false, nil
}
func (fct *FolderCleanerTask) clean() error {
	logger := fct.logger.WithFields(log.Fields{"Function": "clean"})
	logger.Infof("Cleaning")

	return filepath.WalkDir(fct.config.Path, func(path string, d fs.DirEntry, err error) error {
		logger := logger.WithFields(log.Fields{"InnerFunction": "WalkDirFunc", "Path": path, "Error": err})
		logger.Debugf("In WalkDir with DirEntry: %#v", d)
		if err != nil {
			return err
		}
		fileInfo, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("when running stat on %s: %w", path, err)
		}
		if fileInfo.IsDir() {
			if fct.config.Recursive {
				return nil // Traverse into the folder
			} else if path == fct.config.Path {
				return nil // Traverse into the folder
			} else {
				logger.Infof("Skip folder as the recursive flag isn't set")
				return filepath.SkipDir
			}
		}
		logger = logger.WithFields(log.Fields{"FileName": fileInfo.Name(), "Modified": fileInfo.ModTime()})
		deleted, err := fct.conditionallyRemove(fileInfo, path)
		if err != nil {
			if deleted {
				return fmt.Errorf("error trying to delete the file: %w", err)
			} else {
				return fmt.Errorf("there was an error investigating the file: %w", err)
			}
		}
		if deleted {
			logger.Info("deleted the file")
		}
		return nil
	})
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
			promFileCleanFailures.With(fct.labels).Add(1)
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

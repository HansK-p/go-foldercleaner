package foldercleaner

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const monitorNamespace = "foldercleaner"

var (
	promFilesRemoved = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: monitorNamespace,
			Name:      "removed_files_count",
			Help:      "Number of files Removed",
		},
		[]string{"path", "pattern"},
	)
	promFileCleanFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: monitorNamespace,
			Name:      "clean_folder_failures_count",
			Help:      "Number of times the clean folders operation has failed",
		},
		[]string{"path", "pattern"},
	)
	promFileRemoveFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: monitorNamespace,
			Name:      "remove_file_failures_count",
			Help:      "Number of file remove failures",
		},
		[]string{"path", "pattern"},
	)
)

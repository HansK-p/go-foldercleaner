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
	promFileRemoveFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: monitorNamespace,
			Name:      "remove_file_failuers_count",
			Help:      "Number of file remove failures",
		},
		[]string{"path", "pattern"},
	)
)

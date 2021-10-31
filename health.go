package foldercleaner

func (fc FolderCleaner) IsAlive() (bool, string) {
	for _, fct := range fc.tasks {
		isHealthy, healthDescription := fct.IsAlive()
		if !isHealthy {
			return isHealthy, healthDescription
		}
	}
	return true, "All cleaner tasks are alive"
}

func (fc FolderCleaner) IsReady() (bool, string) {
	for _, fct := range fc.tasks {
		isHealthy, healthDescription := fct.IsReady()
		if !isHealthy {
			return isHealthy, healthDescription
		}
	}
	return true, "All cleaner tasks are ready"
}

func (fct FolderCleanerTask) IsAlive() (bool, string) {
	return fct.isHealthy, fct.healthDescription
}

func (fct FolderCleanerTask) IsReady() (bool, string) {
	return fct.IsAlive()
}

func (fct *FolderCleanerTask) setHealthStatus(isHealthy bool, healthDescription string) {
	fct.healthDescription = healthDescription
	fct.isHealthy = isHealthy
}

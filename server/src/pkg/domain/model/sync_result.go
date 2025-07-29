package model

// TimeEntrySyncResult contains the time entries that have been created, updated, or deleted
type TimeEntrySyncResult struct {
	Created []TimeEntry
	Updated []TimeEntry
	Deleted []TimeEntry
}

// ProjectSyncResult contains the projects that have been created, updated, or deleted
type ProjectSyncResult struct {
	Created []Project
	Updated []Project
	Deleted []Project
}

package model

// TimeEntrySyncResult contains separate lists for created, updated, and deleted time entries
type TimeEntrySyncResult struct {
	Created []TimeEntry
	Updated []TimeEntry
	Deleted []TimeEntry
}

// ProjectSyncResult contains separate lists for created, updated, and deleted projects
type ProjectSyncResult struct {
	Created []Project
	Updated []Project
	Deleted []Project
}

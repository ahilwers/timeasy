package model

type SyncData struct {
	TimeEntriesToBeUpdated []TimeEntry
	TimeEntriesToBeDeleted []TimeEntry
	ProjectsToBeUpdated    []Project
	ProjectsToBeDeleted    []Project
}

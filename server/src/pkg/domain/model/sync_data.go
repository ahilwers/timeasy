package model

type SyncData struct {
	TimeEntriesToBeCreated []TimeEntry
	TimeEntriesToBeUpdated []TimeEntry
	TimeEntriesToBeDeleted []TimeEntry
	ProjectsToBeCreated    []Project
	ProjectsToBeUpdated    []Project
	ProjectsToBeDeleted    []Project
}

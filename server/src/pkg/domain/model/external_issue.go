package model

import (
	"time"

	"github.com/gofrs/uuid"
)

type ExternalIssue struct {
	ID           uuid.UUID `json:"id"`
	ProjectID    uuid.UUID `json:"projectId"`
	Provider     string    `json:"provider"` // "github", "gitlab", "jira"
	KeyOrNumber  string    `json:"key"` // "#123", "ABC-123"
	Title        string    `json:"title"`
	State        string    `json:"state"` // "open", "closed", etc.
	URL          string    `json:"url"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// IssueState represents the possible states of an issue
type IssueState string

const (
	IssueStateOpen     IssueState = "open"
	IssueStateClosed   IssueState = "closed"
	IssueStatePending  IssueState = "pending"
)

// IssueResolveResult represents the result of resolving an issue reference
type IssueResolveResult struct {
	IssueID *uuid.UUID `json:"issueId,omitempty"`
	Key     string     `json:"key"`
	Title   string     `json:"title,omitempty"`
	State   string     `json:"state,omitempty"`
	URL     string     `json:"url,omitempty"`
	Status  string     `json:"status"` // "resolved" or "pending"
}
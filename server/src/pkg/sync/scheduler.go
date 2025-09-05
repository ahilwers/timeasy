package sync

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"timeasy-server/pkg/usecase"

	"github.com/gofrs/uuid"
)

type SyncConfig struct {
	ActiveProjectInterval  time.Duration // 15 minutes for projects with recent activity
	RecentProjectInterval  time.Duration // 1 hour for projects with activity in last 24h
	DormantProjectInterval time.Duration // 24 hours for dormant projects
	MaxConcurrentSyncs     int           // Maximum concurrent sync operations
	ActivityWindowActive   time.Duration // 4 hours - window for "active" classification
	ActivityWindowRecent   time.Duration // 24 hours - window for "recent" classification
}

func DefaultSyncConfig() SyncConfig {
	return SyncConfig{
		ActiveProjectInterval:  15 * time.Minute,
		RecentProjectInterval:  1 * time.Hour,
		DormantProjectInterval: 24 * time.Hour,
		MaxConcurrentSyncs:     3,
		ActivityWindowActive:   4 * time.Hour,
		ActivityWindowRecent:   24 * time.Hour,
	}
}

// ProjectSyncInfo tracks sync state for a project
type ProjectSyncInfo struct {
	ProjectID      uuid.UUID
	LastSyncTime   time.Time
	LastActivity   time.Time
	NextSyncTime   time.Time
	SyncInProgress bool
	FailCount      int
}

// SyncScheduler manages automated issue synchronization
type SyncScheduler struct {
	config                     SyncConfig
	externalIntegrationUsecase *usecase.ExternalIntegrationUseCase
	timeEntryUsecase           usecase.TimeEntryUsecase
	projectUsecase             usecase.ProjectUsecase

	projectSyncInfo map[uuid.UUID]*ProjectSyncInfo
	syncSemaphore   chan struct{}
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewSyncScheduler(
	config SyncConfig,
	externalIntegrationUsecase *usecase.ExternalIntegrationUseCase,
	timeEntryUsecase usecase.TimeEntryUsecase,
	projectUsecase usecase.ProjectUsecase,
) *SyncScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &SyncScheduler{
		config:                     config,
		externalIntegrationUsecase: externalIntegrationUsecase,
		timeEntryUsecase:           timeEntryUsecase,
		projectUsecase:             projectUsecase,
		projectSyncInfo:            make(map[uuid.UUID]*ProjectSyncInfo),
		syncSemaphore:              make(chan struct{}, config.MaxConcurrentSyncs),
		ctx:                        ctx,
		cancel:                     cancel,
	}
}

func (s *SyncScheduler) Start() {
	slog.Info("Starting sync scheduler",
		"active_interval", s.config.ActiveProjectInterval,
		"recent_interval", s.config.RecentProjectInterval,
		"dormant_interval", s.config.DormantProjectInterval,
		"max_concurrent", s.config.MaxConcurrentSyncs)

	go s.discoverConnectedProjects()
	go s.schedulerLoop()
	go s.activityTrackerLoop()
}

func (s *SyncScheduler) Stop() {
	slog.Info("Stopping sync scheduler")
	s.cancel()
}

func (s *SyncScheduler) schedulerLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkAndScheduleSyncs()
		}
	}
}

func (s *SyncScheduler) activityTrackerLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateProjectActivity()
		}
	}
}

func (s *SyncScheduler) discoverConnectedProjects() {
	slog.Info("Discovering connected projects")
	s.updateProjectActivity()
}

func (s *SyncScheduler) updateProjectActivity() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	recentProjects, err := s.timeEntryUsecase.GetProjectsWithRecentActivity(now.Add(-s.config.ActivityWindowRecent))
	if err != nil {
		slog.Error("Error getting projects with recent activity", "error", err)
		return
	}

	for _, projectID := range recentProjects {
		lastActivity, err := s.timeEntryUsecase.GetLastActivityTimeForProject(projectID)
		if err != nil {
			slog.Error("Error getting last activity for project", "project_id", projectID, "error", err)
			continue
		}

		if info, exists := s.projectSyncInfo[projectID]; exists {
			info.LastActivity = lastActivity
		} else {
			// New project discovered
			s.projectSyncInfo[projectID] = &ProjectSyncInfo{
				ProjectID:    projectID,
				LastActivity: lastActivity,
				NextSyncTime: now.Add(s.config.ActiveProjectInterval),
			}
			slog.Info("Discovered new project with external connection", "project_id", projectID)
		}
	}
}

func (s *SyncScheduler) checkAndScheduleSyncs() {
	s.mu.RLock()
	projectsToSync := make([]*ProjectSyncInfo, 0)
	now := time.Now()

	for _, info := range s.projectSyncInfo {
		if !info.SyncInProgress && now.After(info.NextSyncTime) {
			projectsToSync = append(projectsToSync, info)
		}
	}
	s.mu.RUnlock()

	for _, info := range projectsToSync {
		go s.syncProject(info)
	}
}

func (s *SyncScheduler) syncProject(info *ProjectSyncInfo) {
	// Acquire semaphore to limit concurrent syncs
	select {
	case s.syncSemaphore <- struct{}{}:
		defer func() { <-s.syncSemaphore }()
	case <-s.ctx.Done():
		return
	}

	s.mu.Lock()
	info.SyncInProgress = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		info.SyncInProgress = false
		info.LastSyncTime = time.Now()
		info.NextSyncTime = s.calculateNextSyncTime(info)
		s.mu.Unlock()
	}()

	slog.Info("Syncing project", "project_id", info.ProjectID)

	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	err := s.externalIntegrationUsecase.SyncProjectIssues(ctx, info.ProjectID)
	if err != nil {
		info.FailCount++
		slog.Error("Failed to sync project", 
			"project_id", info.ProjectID, 
			"error", err,
			"fail_count", info.FailCount)

		// Apply exponential backoff for failed syncs
		backoffMultiplier := 1 << min(info.FailCount-1, 4) // Max 16x backoff
		info.NextSyncTime = time.Now().Add(s.config.ActiveProjectInterval * time.Duration(backoffMultiplier))
	} else {
		info.FailCount = 0
		slog.Info("Successfully synced project", "project_id", info.ProjectID)
	}
}

func (s *SyncScheduler) calculateNextSyncTime(info *ProjectSyncInfo) time.Time {
	now := time.Now()
	timeSinceActivity := now.Sub(info.LastActivity)

	var interval time.Duration

	switch {
	case timeSinceActivity <= s.config.ActivityWindowActive:
		interval = s.config.ActiveProjectInterval
	case timeSinceActivity <= s.config.ActivityWindowRecent:
		interval = s.config.RecentProjectInterval
	default:
		interval = s.config.DormantProjectInterval
	}

	return now.Add(interval)
}

// RegisterProject adds a project to the sync scheduler
func (s *SyncScheduler) RegisterProject(projectID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projectSyncInfo[projectID]; !exists {
		now := time.Now()
		info := &ProjectSyncInfo{
			ProjectID:    projectID,
			LastActivity: now,
			NextSyncTime: now.Add(s.config.ActiveProjectInterval), // Sync soon after registration
		}
		s.projectSyncInfo[projectID] = info
		slog.Info("Registered project for sync scheduling", "project_id", projectID)
	}
}

// UnregisterProject removes a project from the sync scheduler
func (s *SyncScheduler) UnregisterProject(projectID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.projectSyncInfo, projectID)
	slog.Info("Unregistered project from sync scheduling", "project_id", projectID)
}

func (s *SyncScheduler) TriggerImmediateSync(projectID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if info, exists := s.projectSyncInfo[projectID]; exists {
		info.NextSyncTime = time.Now()
		info.LastActivity = time.Now()
		slog.Info("Triggered immediate sync for project", "project_id", projectID)
	}
}

func (s *SyncScheduler) GetSyncStatus() map[uuid.UUID]*ProjectSyncInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[uuid.UUID]*ProjectSyncInfo)
	for id, info := range s.projectSyncInfo {
		infoCopy := *info
		result[id] = &infoCopy
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

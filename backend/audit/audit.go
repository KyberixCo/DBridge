package audit

import (
	"sync"
	"time"

	"oramcp/backend/models"
)

type EventListener func(entry models.AuditLogEntry)

// AuditManager manages in-memory audit logs of all MCP interactions.
type AuditManager struct {
	mu        sync.RWMutex
	maxLogs   int
	entries   []models.AuditLogEntry
	listeners []EventListener
}

// NewAuditManager creates an audit manager with a maximum history size.
func NewAuditManager(maxLogs int) *AuditManager {
	if maxLogs <= 0 {
		maxLogs = 500
	}
	return &AuditManager{
		maxLogs:   maxLogs,
		entries:   make([]models.AuditLogEntry, 0, maxLogs),
		listeners: make([]EventListener, 0),
	}
}

// AddListener registers a listener to be notified when new audit entries are recorded.
func (am *AuditManager) AddListener(listener EventListener) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.listeners = append(am.listeners, listener)
}

// Record saves an audit entry and notifies listeners.
func (am *AuditManager) Record(entry models.AuditLogEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	am.mu.Lock()
	if len(am.entries) >= am.maxLogs {
		// Drop oldest entry
		am.entries = am.entries[1:]
	}
	am.entries = append(am.entries, entry)
	listeners := make([]EventListener, len(am.listeners))
	copy(listeners, am.listeners)
	am.mu.Unlock()

	// Notify outside lock
	for _, l := range listeners {
		l(entry)
	}
}

// GetEntries returns a slice of all recorded entries (newest first).
func (am *AuditManager) GetEntries() []models.AuditLogEntry {
	am.mu.RLock()
	defer am.mu.RUnlock()

	n := len(am.entries)
	res := make([]models.AuditLogEntry, n)
	for i := 0; i < n; i++ {
		res[i] = am.entries[n-1-i] // reverse order (newest first)
	}
	return res
}

// Clear clears the audit log.
func (am *AuditManager) Clear() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.entries = make([]models.AuditLogEntry, 0, am.maxLogs)
}

// Stats returns total count and blocked count.
func (am *AuditManager) Stats() (total int64, blocked int64) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, e := range am.entries {
		total++
		if !e.Allowed {
			blocked++
		}
	}
	return total, blocked
}

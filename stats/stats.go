package stats

import (
	"fmt"
	"sync"
)

const (
	UpProxy   = "outbound>>>proxy>>>traffic>>>uplink"
	DownProxy = "outbound>>>proxy>>>traffic>>>downlink"
)

var DefaultManager *Manager = nil

func init() {
	DefaultManager = &Manager{
		counters: make(map[string]*Counter, 2),
	}
	DefaultManager.counters[UpProxy] = new(Counter)
	DefaultManager.counters[DownProxy] = new(Counter)
}

type Manager struct {
	access   sync.RWMutex
	counters map[string]*Counter
	running  bool
}

// NewManager creates an instance of Statistics Manager.
func NewManager() (*Manager, error) {
	m := &Manager{
		counters: make(map[string]*Counter),
	}

	return m, nil
}

// RegisterCounter implements stats.Manager.
func (m *Manager) RegisterCounter(name string) (*Counter, error) {
	m.access.Lock()
	defer m.access.Unlock()

	if _, found := m.counters[name]; found {
		return nil, fmt.Errorf("counter %s already registered", name)
	}
	// newError("create new counter ", name).AtDebug().WriteToLog()
	c := new(Counter)
	m.counters[name] = c
	return c, nil
}

func (m *Manager) UnregisterCounter(name string) error {
	m.access.Lock()
	defer m.access.Unlock()

	if _, found := m.counters[name]; found {
		// newError("remove counter ", name).AtDebug().WriteToLog()
		delete(m.counters, name)
	}
	return nil
}

func (m *Manager) GetCounter(name string) *Counter {
	m.access.RLock()
	defer m.access.RUnlock()

	if c, found := m.counters[name]; found {
		return c
	}
	return nil
}

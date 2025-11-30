package links

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Storage struct {
	store   map[int]Cache
	ttl     time.Duration
	groupID int
	mu      sync.RWMutex
}

type Cache struct {
	Data      []LinkInformation `json:"data"`
	ExpiresAt time.Time         `json:"expires_at"`
}

func NewStorage(ttl time.Duration) *Storage {
	return &Storage{
		store:   make(map[int]Cache),
		ttl:     ttl,
		groupID: 1,
	}
}

func (s *Storage) Get(id int) (links []LinkInformation, isExists bool, expired bool) {
	
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.store[id]
	if !ok {
		return nil, false, false
	}

	isExpired := time.Now().After(item.ExpiresAt)
	return item.Data, true, isExpired
}

func (s *Storage) Set(links []LinkInformation) int {
	
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.groupID
	s.store[id] = Cache{
		Data:      links,
		ExpiresAt: time.Now().Add(s.ttl),
	}

	s.groupID++

	return id
}

func (s *Storage) Update(id int, links []LinkInformation) {
	
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.store[id]; ok {
		s.store[id] = Cache{
			Data:      links,
			ExpiresAt: time.Now().Add(s.ttl),
		}
	}
}

func (s *Storage) SaveToFile(filename string) error {
	
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadFromFile loads the storage state from a JSON file.
func (s *Storage) LoadFromFile(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, start empty
		}
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &s.store); err != nil {
		return err
	}

	// Update groupID to avoid collisions
	maxID := 0
	for id := range s.store {
		if id > maxID {
			maxID = id
		}
	}
	s.groupID = maxID + 1

	return nil
}

package links

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Cache struct {
	Data      []LinkInformation `json:"data"`
	ExpiresAt time.Time         `json:"expires_at"`
}

type Storage struct {
	store   map[int]Cache
	ttl     time.Duration
	groupID int
	mu      sync.RWMutex
}

func NewStorage(ttl time.Duration) *Storage {
	return &Storage{
		store:   make(map[int]Cache),
		ttl:     ttl,
		groupID: 1,
	}
}

func (s *Storage) Get(id int) ([]LinkInformation, bool, bool) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	var isExists bool
	data, ok := s.store[id]
	if !ok {
		return nil, isExists, false
	}

	isExpired := time.Now().After(data.ExpiresAt)
	isExists = true

	return data.Data, isExists, isExpired
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

func (s *Storage) SaveToJSONFile(filename string) error {

	s.mu.RLock()
	defer s.mu.RUnlock()

	jsonData, err := json.MarshalIndent(s.store, "", "  ")
	if err != nil {
		return fmt.Errorf("error serialization in SaveToJSONFile: %w", err)
	}

	tmpFile := filename + ".tmp"
	if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing into tmp-file: %w", err)
	}

	return os.Rename(tmpFile, filename)
}

func (s *Storage) ReadFromJSONFile(filename string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) || len(data) == 0 {
			return nil
		}
		return fmt.Errorf("file not exist or empty: %w", err)
	}

	if err := json.Unmarshal(data, &s.store); err != nil {
		return fmt.Errorf("error deserialization in ReadFromJSONFile: %w", err)
	}

	max := 0
	for id := range s.store {
		if id > max {
			max = id
		}
	}
	s.groupID = max + 1

	return nil
}

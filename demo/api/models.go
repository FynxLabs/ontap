package main

import (
	"sync"
	"time"
)

// Item represents an item in our API
type Item struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ItemInput represents the input for creating or updating an item
type ItemInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// InMemoryStore is a simple in-memory data store for items
type InMemoryStore struct {
	items     map[int]Item
	nextID    int
	itemsLock sync.RWMutex
}

// NewInMemoryStore creates a new in-memory store with some sample data
func NewInMemoryStore() *InMemoryStore {
	store := &InMemoryStore{
		items:  make(map[int]Item),
		nextID: 1,
	}

	// Add some sample items
	store.AddItem(Item{
		Name:        "Sample Item 1",
		Description: "This is the first sample item",
		CreatedAt:   time.Now(),
	})
	store.AddItem(Item{
		Name:        "Sample Item 2",
		Description: "This is the second sample item",
		CreatedAt:   time.Now(),
	})
	store.AddItem(Item{
		Name:        "Sample Item 3",
		Description: "This is the third sample item",
		CreatedAt:   time.Now(),
	})

	return store
}

// GetItems returns all items in the store
func (s *InMemoryStore) GetItems(limit, offset int) []Item {
	s.itemsLock.RLock()
	defer s.itemsLock.RUnlock()

	// Convert map to slice
	items := make([]Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}

	// Apply offset and limit
	if offset >= len(items) {
		return []Item{}
	}

	end := offset + limit
	if end > len(items) {
		end = len(items)
	}

	return items[offset:end]
}

// GetItem returns an item by ID
func (s *InMemoryStore) GetItem(id int) (Item, bool) {
	s.itemsLock.RLock()
	defer s.itemsLock.RUnlock()

	item, exists := s.items[id]
	return item, exists
}

// AddItem adds a new item to the store
func (s *InMemoryStore) AddItem(item Item) Item {
	s.itemsLock.Lock()
	defer s.itemsLock.Unlock()

	// Set the ID and created time
	item.ID = s.nextID
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}

	// Add to the store
	s.items[item.ID] = item
	s.nextID++

	return item
}

// UpdateItem updates an existing item
func (s *InMemoryStore) UpdateItem(id int, item Item) (Item, bool) {
	s.itemsLock.Lock()
	defer s.itemsLock.Unlock()

	// Check if the item exists
	existingItem, exists := s.items[id]
	if !exists {
		return Item{}, false
	}

	// Update the item
	item.ID = id
	item.CreatedAt = existingItem.CreatedAt
	s.items[id] = item

	return item, true
}

// DeleteItem deletes an item by ID
func (s *InMemoryStore) DeleteItem(id int) bool {
	s.itemsLock.Lock()
	defer s.itemsLock.Unlock()

	// Check if the item exists
	_, exists := s.items[id]
	if !exists {
		return false
	}

	// Delete the item
	delete(s.items, id)
	return true
}

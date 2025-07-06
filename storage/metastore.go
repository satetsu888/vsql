package storage

import (
	"sync"
)

type MetaStore struct {
	tableColumns map[string]map[string]bool
	columnOrder  map[string][]string // Maintains column order for each table
	mu           sync.RWMutex
}

func NewMetaStore() *MetaStore {
	return &MetaStore{
		tableColumns: make(map[string]map[string]bool),
		columnOrder:  make(map[string][]string),
	}
}

func (ms *MetaStore) AddColumn(tableName, columnName string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if _, exists := ms.tableColumns[tableName]; !exists {
		ms.tableColumns[tableName] = make(map[string]bool)
	}
	ms.tableColumns[tableName][columnName] = true
}

func (ms *MetaStore) AddColumns(tableName string, columns []string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if _, exists := ms.tableColumns[tableName]; !exists {
		ms.tableColumns[tableName] = make(map[string]bool)
	}
	
	// Store column order
	ms.columnOrder[tableName] = columns
	
	for _, col := range columns {
		ms.tableColumns[tableName][col] = true
	}
}

func (ms *MetaStore) GetTableColumns(tableName string) []string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	// Return columns in the order they were defined
	if order, exists := ms.columnOrder[tableName]; exists {
		return append([]string{}, order...) // Return a copy
	}
	
	// Fallback to unordered columns if no order is stored
	cols, exists := ms.tableColumns[tableName]
	if !exists {
		return []string{}
	}
	
	result := make([]string, 0, len(cols))
	for col := range cols {
		result = append(result, col)
	}
	return result
}

func (ms *MetaStore) DropTable(tableName string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.tableColumns, tableName)
	delete(ms.columnOrder, tableName)
}

func (ms *MetaStore) UpdateFromRow(tableName string, row Row) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if _, exists := ms.tableColumns[tableName]; !exists {
		ms.tableColumns[tableName] = make(map[string]bool)
	}
	
	// Only add new columns, preserving existing order
	for col := range row {
		if !ms.tableColumns[tableName][col] {
			ms.tableColumns[tableName][col] = true
			// Append new columns to the order
			ms.columnOrder[tableName] = append(ms.columnOrder[tableName], col)
		}
	}
}
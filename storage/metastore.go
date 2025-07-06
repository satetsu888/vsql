package storage

import (
	"sync"
)

type MetaStore struct {
	tableColumns map[string]map[string]bool
	mu           sync.RWMutex
}

func NewMetaStore() *MetaStore {
	return &MetaStore{
		tableColumns: make(map[string]map[string]bool),
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
	
	for _, col := range columns {
		ms.tableColumns[tableName][col] = true
	}
}

func (ms *MetaStore) GetTableColumns(tableName string) []string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
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
}

func (ms *MetaStore) UpdateFromRow(tableName string, row Row) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if _, exists := ms.tableColumns[tableName]; !exists {
		ms.tableColumns[tableName] = make(map[string]bool)
	}
	
	for col := range row {
		ms.tableColumns[tableName][col] = true
	}
}
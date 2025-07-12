package storage

import (
	"sync"
	"time"
)

type MetaStore struct {
	tableColumns map[string]map[string]bool
	columnOrder  map[string][]string // Maintains column order for each table
	columnTypes  map[string]map[string]*ColumnTypeInfo // Column type information
	mu           sync.RWMutex
}

func NewMetaStore() *MetaStore {
	return &MetaStore{
		tableColumns: make(map[string]map[string]bool),
		columnOrder:  make(map[string][]string),
		columnTypes:  make(map[string]map[string]*ColumnTypeInfo),
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
	delete(ms.columnTypes, tableName)
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

// GetColumnType returns the type of a column
func (ms *MetaStore) GetColumnType(tableName, columnName string) ColumnType {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if tableTypes, exists := ms.columnTypes[tableName]; exists {
		if typeInfo, exists := tableTypes[columnName]; exists {
			return typeInfo.CurrentType
		}
	}
	return TypeUnknown
}

// SetColumnType sets or updates the type of a column based on a value
func (ms *MetaStore) SetColumnType(tableName, columnName string, value interface{}) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	// Initialize table type map if needed
	if _, exists := ms.columnTypes[tableName]; !exists {
		ms.columnTypes[tableName] = make(map[string]*ColumnTypeInfo)
	}
	
	// Get or create column type info
	typeInfo, exists := ms.columnTypes[tableName][columnName]
	if !exists {
		typeInfo = &ColumnTypeInfo{
			CurrentType: TypeUnknown,
			IsConfirmed: false,
		}
		ms.columnTypes[tableName][columnName] = typeInfo
	}
	
	// Skip NULL values - they don't affect type
	if value == nil {
		return nil
	}
	
	// Infer type from value
	newType := InferTypeFromValue(value)
	
	// Check compatibility
	if !IsTypeCompatible(typeInfo.CurrentType, newType) {
		return TypeMismatchError{
			Table:    tableName,
			Column:   columnName,
			Expected: typeInfo.CurrentType,
			Actual:   newType,
		}
	}
	
	// Update type if needed
	if typeInfo.CurrentType != newType {
		typeInfo.CurrentType = newType
		typeInfo.IsConfirmed = true
		typeInfo.LastUpdateTime = time.Now()
	}
	
	return nil
}

// ValidateValueType validates that a value is compatible with the column's type
func (ms *MetaStore) ValidateValueType(tableName, columnName string, value interface{}) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	// NULL is always valid
	if value == nil {
		return nil
	}
	
	// Get column type info
	if tableTypes, exists := ms.columnTypes[tableName]; exists {
		if typeInfo, exists := tableTypes[columnName]; exists && typeInfo.IsConfirmed {
			valueType := InferTypeFromValue(value)
			if !IsTypeCompatible(typeInfo.CurrentType, valueType) {
				return TypeMismatchError{
					Table:    tableName,
					Column:   columnName,
					Expected: typeInfo.CurrentType,
					Actual:   valueType,
				}
			}
		}
	}
	
	return nil
}
package storage

import (
	"sync"
)

type Row map[string]interface{}

type Table struct {
	Name string
	Rows []Row
	mu   sync.RWMutex
}

func (t *Table) Insert(row Row) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Rows = append(t.Rows, row)
}

func (t *Table) GetRows() []Row {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]Row, len(t.Rows))
	copy(result, t.Rows)
	return result
}

type DataStore struct {
	tables map[string]*Table
	mu     sync.RWMutex
}

func NewDataStore() *DataStore {
	return &DataStore{
		tables: make(map[string]*Table),
	}
}

func (ds *DataStore) CreateTable(name string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	
	if _, exists := ds.tables[name]; exists {
		return nil
	}
	
	ds.tables[name] = &Table{
		Name: name,
		Rows: make([]Row, 0),
	}
	return nil
}

func (ds *DataStore) GetTable(name string) (*Table, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	table, exists := ds.tables[name]
	return table, exists
}

func (ds *DataStore) DropTable(name string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	delete(ds.tables, name)
}

func (ds *DataStore) ListTables() []string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	
	tables := make([]string, 0, len(ds.tables))
	for name := range ds.tables {
		tables = append(tables, name)
	}
	return tables
}
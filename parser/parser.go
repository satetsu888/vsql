package parser

import (
	"fmt"
	"regexp"
	"strings"
)

type QueryType int

const (
	SelectQuery QueryType = iota
	InsertQuery
	UpdateQuery
	DeleteQuery
	CreateTableQuery
	DropTableQuery
)

type ParsedQuery struct {
	Type      QueryType
	Table     string
	Columns   []string
	Values    [][]interface{}
	Where     *WhereClause
	OrderBy   []OrderByClause
	Limit     *int
}

type WhereClause struct {
	Column   string
	Operator string
	Value    interface{}
}

type OrderByClause struct {
	Column string
	Desc   bool
}

func Parse(query string) (*ParsedQuery, error) {
	query = strings.TrimSpace(query)
	queryUpper := strings.ToUpper(query)

	switch {
	case strings.HasPrefix(queryUpper, "SELECT"):
		return parseSelect(query)
	case strings.HasPrefix(queryUpper, "INSERT"):
		return parseInsert(query)
	case strings.HasPrefix(queryUpper, "CREATE TABLE"):
		return parseCreateTable(query)
	default:
		return nil, fmt.Errorf("unsupported query type")
	}
}

func parseSelect(query string) (*ParsedQuery, error) {
	selectRegex := regexp.MustCompile(`(?i)SELECT\s+(.+?)\s+FROM\s+(\w+)(?:\s+WHERE\s+(.+?))?(?:\s+ORDER\s+BY\s+(.+?))?(?:\s+LIMIT\s+(\d+))?$`)
	matches := selectRegex.FindStringSubmatch(query)
	
	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid SELECT query")
	}

	result := &ParsedQuery{
		Type:  SelectQuery,
		Table: matches[2],
	}

	columnPart := strings.TrimSpace(matches[1])
	if columnPart == "*" {
		result.Columns = []string{"*"}
	} else {
		columns := strings.Split(columnPart, ",")
		for i, col := range columns {
			columns[i] = strings.TrimSpace(col)
		}
		result.Columns = columns
	}

	if matches[3] != "" {
		where, err := parseWhere(matches[3])
		if err != nil {
			return nil, err
		}
		result.Where = where
	}

	return result, nil
}

func parseInsert(query string) (*ParsedQuery, error) {
	insertRegex := regexp.MustCompile(`(?i)INSERT\s+INTO\s+(\w+)\s*\(([^)]+)\)\s*VALUES\s*(.+)`)
	matches := insertRegex.FindStringSubmatch(query)
	
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid INSERT query")
	}

	result := &ParsedQuery{
		Type:  InsertQuery,
		Table: matches[1],
	}

	columns := strings.Split(matches[2], ",")
	for i, col := range columns {
		columns[i] = strings.TrimSpace(col)
	}
	result.Columns = columns

	valuesPart := matches[3]
	valueSetRegex := regexp.MustCompile(`\([^)]+\)`)
	valueSets := valueSetRegex.FindAllString(valuesPart, -1)
	
	for _, valueSet := range valueSets {
		valueSet = strings.Trim(valueSet, "()")
		values := parseValues(valueSet)
		result.Values = append(result.Values, values)
	}

	return result, nil
}

func parseCreateTable(query string) (*ParsedQuery, error) {
	createTableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(\w+)`)
	matches := createTableRegex.FindStringSubmatch(query)
	
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid CREATE TABLE query")
	}

	return &ParsedQuery{
		Type:  CreateTableQuery,
		Table: matches[1],
	}, nil
}

func parseWhere(whereClause string) (*WhereClause, error) {
	whereRegex := regexp.MustCompile(`(\w+)\s*(=|!=|<>|>|<|>=|<=)\s*(.+)`)
	matches := whereRegex.FindStringSubmatch(whereClause)
	
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid WHERE clause")
	}

	value := strings.TrimSpace(matches[3])
	value = strings.Trim(value, "'\"")

	return &WhereClause{
		Column:   matches[1],
		Operator: matches[2],
		Value:    value,
	}, nil
}

func parseValues(valueString string) []interface{} {
	values := []interface{}{}
	parts := strings.Split(valueString, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "'") && strings.HasSuffix(part, "'") {
			values = append(values, strings.Trim(part, "'"))
		} else if part == "NULL" || part == "null" {
			values = append(values, nil)
		} else {
			values = append(values, part)
		}
	}
	
	return values
}
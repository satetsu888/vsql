package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// toFloat64 converts various types to float64
func toFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}

// toNumber is a compatibility wrapper that returns bool instead of error
func toNumber(val interface{}) (float64, bool) {
	num, err := toFloat64(val)
	return num, err == nil
}

// extractAConstValue extracts the actual value from an A_Const node
func extractAConstValue(aConst *pg_query.A_Const) interface{} {
	if aConst.Isnull {
		return nil
	}

	switch val := aConst.Val.(type) {
	case *pg_query.A_Const_Sval:
		return val.Sval.Sval
	case *pg_query.A_Const_Ival:
		return int(val.Ival.Ival)
	case *pg_query.A_Const_Fval:
		return val.Fval.Fval
	case *pg_query.A_Const_Boolval:
		return val.Boolval.Boolval
	}
	return nil
}

// compareValuesPg compares two values using the given operator
// Implements SQL three-valued logic for NULL handling
func compareValuesPg(left interface{}, operator string, right interface{}) bool {
	// SQL three-valued logic: any comparison with NULL returns UNKNOWN (treated as false)
	// This includes NULL = NULL, which should return UNKNOWN, not true
	if left == nil || right == nil {
		return false
	}

	// Try to compare as booleans first
	leftBool, leftIsBool := left.(bool)
	rightBool, rightIsBool := right.(bool)
	
	if leftIsBool && rightIsBool {
		switch operator {
		case "=":
			return leftBool == rightBool
		case "!=", "<>":
			return leftBool != rightBool
		default:
			// Boolean values don't support <, >, <=, >= operators
			return false
		}
	}
	
	// Try to compare as numbers
	leftNum, leftIsNum := toNumber(left)
	rightNum, rightIsNum := toNumber(right)
	
	if leftIsNum && rightIsNum {
		switch operator {
		case "=":
			return leftNum == rightNum
		case "!=", "<>":
			return leftNum != rightNum
		case "<":
			return leftNum < rightNum
		case ">":
			return leftNum > rightNum
		case "<=":
			return leftNum <= rightNum
		case ">=":
			return leftNum >= rightNum
		}
	}
	
	// Fall back to string comparison
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	switch operator {
	case "=":
		return leftStr == rightStr
	case "!=", "<>":
		return leftStr != rightStr
	case "<":
		return leftStr < rightStr
	case ">":
		return leftStr > rightStr
	case "<=":
		return leftStr <= rightStr
	case ">=":
		return leftStr >= rightStr
	case "~~": // LIKE operator in PostgreSQL
		return matchPattern(leftStr, rightStr)
	case "!~~": // NOT LIKE operator
		return !matchPattern(leftStr, rightStr)
	case "~~*": // ILIKE operator (case-insensitive)
		return matchPattern(strings.ToLower(leftStr), strings.ToLower(rightStr))
	case "!~~*": // NOT ILIKE operator
		return !matchPattern(strings.ToLower(leftStr), strings.ToLower(rightStr))
	}

	return false
}

// matchPattern implements SQL LIKE pattern matching
// % matches any sequence of characters
// _ matches any single character
// \ is the escape character
func matchPattern(text, pattern string) bool {
	// Convert SQL pattern to regex pattern
	regexPattern := ""
	i := 0
	for i < len(pattern) {
		ch := pattern[i]
		if ch == '\\' && i+1 < len(pattern) {
			// Escape character - add the next character literally
			i++
			regexPattern += regexp.QuoteMeta(string(pattern[i]))
		} else if ch == '%' {
			regexPattern += ".*"
		} else if ch == '_' {
			regexPattern += "."
		} else {
			regexPattern += regexp.QuoteMeta(string(ch))
		}
		i++
	}
	
	// Anchor the pattern to match the entire string
	regexPattern = "^" + regexPattern + "$"
	
	matched, err := regexp.MatchString(regexPattern, text)
	if err != nil {
		return false
	}
	return matched
}

// extractColumnNameFromRef extracts column name from a ColumnRef node
// Returns the last field (column name) for both qualified and unqualified references
func extractColumnNameFromRef(colRef *pg_query.ColumnRef) string {
	if colRef == nil || len(colRef.Fields) == 0 {
		return ""
	}
	
	// Get the last field as the column name
	// For "table.column", fields[0] is table, fields[1] is column
	// For "column", fields[0] is column
	lastField := colRef.Fields[len(colRef.Fields)-1]
	if str, ok := lastField.Node.(*pg_query.Node_String_); ok {
		return str.String_.Sval
	}
	return ""
}

// extractTableAndColumnFromRef extracts both table and column names from a qualified reference
// Returns (tableName, columnName) - tableName is empty for unqualified references
func extractTableAndColumnFromRef(colRef *pg_query.ColumnRef) (string, string) {
	if colRef == nil || len(colRef.Fields) == 0 {
		return "", ""
	}
	
	if len(colRef.Fields) >= 2 {
		// Qualified reference (table.column)
		tableName := ""
		columnName := ""
		
		if str, ok := colRef.Fields[0].Node.(*pg_query.Node_String_); ok {
			tableName = str.String_.Sval
		}
		if str, ok := colRef.Fields[1].Node.(*pg_query.Node_String_); ok {
			columnName = str.String_.Sval
		}
		return tableName, columnName
	}
	
	// Unqualified reference
	if str, ok := colRef.Fields[0].Node.(*pg_query.Node_String_); ok {
		return "", str.String_.Sval
	}
	return "", ""
}

// isAggregateFunction checks if a function name is an aggregate function
func isAggregateFunction(funcName string) bool {
	switch strings.ToUpper(funcName) {
	case "COUNT", "SUM", "AVG", "MAX", "MIN":
		return true
	default:
		return false
	}
}

// getFunctionName extracts function name from FuncCall
func getFunctionName(funcCall *pg_query.FuncCall) string {
	if funcCall == nil || len(funcCall.Funcname) == 0 {
		return ""
	}
	if str, ok := funcCall.Funcname[0].Node.(*pg_query.Node_String_); ok {
		return strings.ToUpper(str.String_.Sval)
	}
	return ""
}
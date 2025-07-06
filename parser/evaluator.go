package parser

import (
	"fmt"
	"strconv"
	"vsql/storage"
)

func EvaluateWhere(row storage.Row, where *WhereClause) bool {
	if where == nil {
		return true
	}

	rowValue, exists := row[where.Column]
	if !exists {
		rowValue = nil
	}

	return compareValues(rowValue, where.Operator, where.Value)
}

func compareValues(rowValue interface{}, operator string, compareValue interface{}) bool {
	if rowValue == nil && compareValue == nil {
		return operator == "=" || operator == "<=" || operator == ">="
	}
	
	if rowValue == nil || compareValue == nil {
		return operator == "!=" || operator == "<>"
	}

	rowStr := fmt.Sprintf("%v", rowValue)
	compareStr := fmt.Sprintf("%v", compareValue)

	rowFloat, rowErr := strconv.ParseFloat(rowStr, 64)
	compareFloat, compareErr := strconv.ParseFloat(compareStr, 64)

	if rowErr == nil && compareErr == nil {
		switch operator {
		case "=":
			return rowFloat == compareFloat
		case "!=", "<>":
			return rowFloat != compareFloat
		case ">":
			return rowFloat > compareFloat
		case "<":
			return rowFloat < compareFloat
		case ">=":
			return rowFloat >= compareFloat
		case "<=":
			return rowFloat <= compareFloat
		}
	}

	switch operator {
	case "=":
		return rowStr == compareStr
	case "!=", "<>":
		return rowStr != compareStr
	case ">":
		return rowStr > compareStr
	case "<":
		return rowStr < compareStr
	case ">=":
		return rowStr >= compareStr
	case "<=":
		return rowStr <= compareStr
	}

	return false
}
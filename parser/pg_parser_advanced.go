package parser

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
	"vsql/storage"
)

type QueryContext struct {
	dataStore    *storage.DataStore
	metaStore    *storage.MetaStore
	tables       map[string]*TableContext
	subqueries   map[string][]storage.Row
	aggregations map[string]interface{}
}

type TableContext struct {
	name  string
	alias string
	rows  []storage.Row
}

func hasAggregateFunctions(targetList []*pg_query.Node) bool {
	for _, target := range targetList {
		if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok {
			if resTarget.ResTarget.Val != nil {
				if _, isFuncCall := resTarget.ResTarget.Val.Node.(*pg_query.Node_FuncCall); isFuncCall {
					return true
				}
			}
		}
	}
	return false
}

func executePgSelectAdvanced(stmt *pg_query.SelectStmt, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	ctx := &QueryContext{
		dataStore:    dataStore,
		metaStore:    metaStore,
		tables:       make(map[string]*TableContext),
		subqueries:   make(map[string][]storage.Row),
		aggregations: make(map[string]interface{}),
	}

	// Handle UNION/INTERSECT/EXCEPT queries
	if stmt.Op != pg_query.SetOperation_SETOP_NONE {
		return executeSetOperation(stmt, dataStore, metaStore)
	}

	// Handle FROM clause (including JOINs and subqueries)
	rows, err := processFromClause(ctx, stmt.FromClause)
	if err != nil {
		return nil, nil, "", err
	}

	// Apply WHERE clause
	if stmt.WhereClause != nil {
		rows = filterRows(rows, stmt.WhereClause, ctx)
	}

	// Check if we have aggregate functions
	hasAggregates := hasAggregateFunctions(stmt.TargetList)
	
	// Handle GROUP BY
	var groupedRows map[string][]storage.Row
	if len(stmt.GroupClause) > 0 {
		groupedRows = groupRows(rows, stmt.GroupClause)
	} else if hasAggregates {
		// If we have aggregates but no GROUP BY, treat all rows as one group
		groupedRows = map[string][]storage.Row{
			"__all__": rows,
		}
	}

	// Process SELECT columns (including aggregations)
	columns, resultRows, err := processSelectList(ctx, stmt.TargetList, rows, groupedRows, stmt.GroupClause)
	if err != nil {
		return nil, nil, "", err
	}

	// Apply DISTINCT
	if stmt.DistinctClause != nil && len(stmt.DistinctClause) > 0 {
		resultRows = applyDistinct(resultRows)
	}

	// Apply HAVING clause
	if stmt.HavingClause != nil && groupedRows != nil {
		resultRows = filterResultRows(resultRows, columns, stmt.HavingClause)
	}

	// Apply ORDER BY
	if len(stmt.SortClause) > 0 {
		resultRows = sortRows(resultRows, columns, stmt.SortClause)
	}

	// Apply LIMIT and OFFSET
	if stmt.LimitCount != nil || stmt.LimitOffset != nil {
		resultRows = applyLimitOffset(resultRows, stmt.LimitCount, stmt.LimitOffset)
	}

	return columns, resultRows, fmt.Sprintf("SELECT %d", len(resultRows)), nil
}

func processFromClause(ctx *QueryContext, fromClause []*pg_query.Node) ([]storage.Row, error) {
	if len(fromClause) == 0 {
		return nil, fmt.Errorf("no FROM clause specified")
	}

	var result []storage.Row
	for i, fromNode := range fromClause {
		switch from := fromNode.Node.(type) {
		case *pg_query.Node_JoinExpr:
			// Handle JOIN
			leftRows, err := processFromNode(ctx, from.JoinExpr.Larg)
			if err != nil {
				return nil, err
			}
			rightRows, err := processFromNode(ctx, from.JoinExpr.Rarg)
			if err != nil {
				return nil, err
			}
			result = performJoin(leftRows, rightRows, from.JoinExpr, ctx)
		case *pg_query.Node_RangeVar:
			// Handle simple table
			if i == 0 {
				rows, err := processFromNode(ctx, fromNode)
				if err != nil {
					return nil, err
				}
				result = rows
			}
		case *pg_query.Node_RangeSubselect:
			// Handle subquery in FROM
			subResult, err := executeSubquery(from.RangeSubselect.Subquery, ctx)
			if err != nil {
				return nil, err
			}
			result = subResult
		}
	}

	return result, nil
}

func processFromNode(ctx *QueryContext, node *pg_query.Node) ([]storage.Row, error) {
	switch n := node.Node.(type) {
	case *pg_query.Node_RangeVar:
		tableName := n.RangeVar.Relname
		alias := n.RangeVar.Alias
		if alias != nil {
			tableName = alias.Aliasname
		}

		table, exists := ctx.dataStore.GetTable(n.RangeVar.Relname)
		if !exists {
			return nil, fmt.Errorf("table '%s' does not exist", n.RangeVar.Relname)
		}

		rows := table.GetRows()
		ctx.tables[tableName] = &TableContext{
			name:  n.RangeVar.Relname,
			alias: tableName,
			rows:  rows,
		}
		return rows, nil
	case *pg_query.Node_JoinExpr:
		return processFromClause(ctx, []*pg_query.Node{node})
	case *pg_query.Node_RangeSubselect:
		return executeSubquery(n.RangeSubselect.Subquery, ctx)
	}
	return nil, fmt.Errorf("unsupported FROM node type")
}

func performJoin(leftRows, rightRows []storage.Row, joinExpr *pg_query.JoinExpr, ctx *QueryContext) []storage.Row {
	var result []storage.Row

	switch joinExpr.Jointype {
	case pg_query.JoinType_JOIN_INNER:
		// INNER JOIN
		for _, leftRow := range leftRows {
			for _, rightRow := range rightRows {
				if joinExpr.Quals == nil || evaluateJoinCondition(leftRow, rightRow, joinExpr.Quals, ctx) {
					mergedRow := mergeRows(leftRow, rightRow)
					result = append(result, mergedRow)
				}
			}
		}
	case pg_query.JoinType_JOIN_LEFT:
		// LEFT JOIN
		for _, leftRow := range leftRows {
			matched := false
			for _, rightRow := range rightRows {
				if joinExpr.Quals == nil || evaluateJoinCondition(leftRow, rightRow, joinExpr.Quals, ctx) {
					mergedRow := mergeRows(leftRow, rightRow)
					result = append(result, mergedRow)
					matched = true
				}
			}
			if !matched {
				// Add left row with NULLs for right columns
				mergedRow := make(storage.Row)
				for k, v := range leftRow {
					mergedRow[k] = v
				}
				result = append(result, mergedRow)
			}
		}
	case pg_query.JoinType_JOIN_RIGHT:
		// RIGHT JOIN
		for _, rightRow := range rightRows {
			matched := false
			for _, leftRow := range leftRows {
				if joinExpr.Quals == nil || evaluateJoinCondition(leftRow, rightRow, joinExpr.Quals, ctx) {
					mergedRow := mergeRows(leftRow, rightRow)
					result = append(result, mergedRow)
					matched = true
				}
			}
			if !matched {
				// Add right row with NULLs for left columns
				mergedRow := make(storage.Row)
				for k, v := range rightRow {
					mergedRow[k] = v
				}
				result = append(result, mergedRow)
			}
		}
	case pg_query.JoinType_JOIN_FULL:
		// FULL OUTER JOIN
		leftMatched := make(map[int]bool)
		rightMatched := make(map[int]bool)

		// First, do inner join part
		for i, leftRow := range leftRows {
			for j, rightRow := range rightRows {
				if joinExpr.Quals == nil || evaluateJoinCondition(leftRow, rightRow, joinExpr.Quals, ctx) {
					mergedRow := mergeRows(leftRow, rightRow)
					result = append(result, mergedRow)
					leftMatched[i] = true
					rightMatched[j] = true
				}
			}
		}

		// Add unmatched left rows
		for i, leftRow := range leftRows {
			if !leftMatched[i] {
				mergedRow := make(storage.Row)
				for k, v := range leftRow {
					mergedRow[k] = v
				}
				result = append(result, mergedRow)
			}
		}

		// Add unmatched right rows
		for j, rightRow := range rightRows {
			if !rightMatched[j] {
				mergedRow := make(storage.Row)
				for k, v := range rightRow {
					mergedRow[k] = v
				}
				result = append(result, mergedRow)
			}
		}
	}

	return result
}

func evaluateJoinCondition(leftRow, rightRow storage.Row, condition *pg_query.Node, ctx *QueryContext) bool {
	// Create a merged row for evaluation
	mergedRow := mergeRows(leftRow, rightRow)
	return evaluatePgWhere(mergedRow, condition)
}

func mergeRows(left, right storage.Row) storage.Row {
	merged := make(storage.Row)
	for k, v := range left {
		merged[k] = v
	}
	for k, v := range right {
		merged[k] = v
	}
	return merged
}

func executeSubquery(subquery *pg_query.Node, ctx *QueryContext) ([]storage.Row, error) {
	if selectStmt, ok := subquery.Node.(*pg_query.Node_SelectStmt); ok {
		columns, rows, _, err := executePgSelectAdvanced(selectStmt.SelectStmt, ctx.dataStore, ctx.metaStore)
		if err != nil {
			return nil, err
		}

		// Convert [][]interface{} to []storage.Row
		var result []storage.Row
		for _, row := range rows {
			storageRow := make(storage.Row)
			// Map values to their column names
			for i, val := range row {
				if i < len(columns) {
					storageRow[columns[i]] = val
				} else {
					storageRow[fmt.Sprintf("col%d", i)] = val
				}
			}
			result = append(result, storageRow)
		}
		return result, nil
	}
	return nil, fmt.Errorf("unsupported subquery type")
}

func filterRows(rows []storage.Row, whereClause *pg_query.Node, ctx *QueryContext) []storage.Row {
	var filtered []storage.Row
	for _, row := range rows {
		if evaluateWhereWithSubqueries(row, whereClause, ctx) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func evaluateWhereWithSubqueries(row storage.Row, expr *pg_query.Node, ctx *QueryContext) bool {
	switch e := expr.Node.(type) {
	case *pg_query.Node_SubLink:
		// Handle subquery in WHERE clause
		return evaluateSubqueryExpression(row, e.SubLink, ctx)
	default:
		// Use existing WHERE evaluation
		return evaluatePgWhere(row, expr)
	}
}

func evaluateSubqueryExpression(row storage.Row, sublink *pg_query.SubLink, ctx *QueryContext) bool {
	// Execute subquery
	subRows, err := executeSubquery(sublink.Subselect, ctx)
	if err != nil {
		return false
	}

	switch sublink.SubLinkType {
	case pg_query.SubLinkType_EXISTS_SUBLINK:
		return len(subRows) > 0
	case pg_query.SubLinkType_ALL_SUBLINK:
		// Handle ALL comparison (used for NOT IN)
		if len(subRows) == 0 {
			return true
		}
		
		// Get the test expression value
		var testValue interface{}
		if sublink.Testexpr != nil {
			testValue = extractValueFromNode(row, sublink.Testexpr)
		}
		
		// If test value is NULL, ALL comparison returns UNKNOWN (false)
		if testValue == nil {
			return false
		}
		
		// Check if any value in the list is NULL
		hasNull := false
		for _, subRow := range subRows {
			for _, val := range subRow {
				if val == nil {
					hasNull = true
				}
				break // Only check first column
			}
		}
		
		// If list contains NULL, NOT IN returns UNKNOWN (false) for all non-NULL values
		if hasNull {
			return false
		}
		
		// Check if testValue doesn't match all rows in subquery result
		for _, subRow := range subRows {
			// Get the first column value from subquery result
			for _, val := range subRow {
				if compareValues(fmt.Sprintf("%v", testValue), "=", val) {
					return false
				}
				break // Only check first column
			}
		}
		return true
	case pg_query.SubLinkType_ANY_SUBLINK:
		// Handle IN/ANY comparison
		if len(subRows) == 0 {
			return false
		}
		
		// Get the test expression value
		var testValue interface{}
		if sublink.Testexpr != nil {
			testValue = extractValueFromNode(row, sublink.Testexpr)
		}
		
		// If test value is NULL, IN always returns UNKNOWN (false)
		if testValue == nil {
			return false
		}
		
		// Check if testValue matches any row in subquery result
		for _, subRow := range subRows {
			// Get the first column value from subquery result
			for _, val := range subRow {
				// Skip NULL values in the list - they don't match anything
				if val == nil {
					continue
				}
				if compareValues(fmt.Sprintf("%v", testValue), "=", val) {
					return true
				}
				break // Only check first column
			}
		}
		return false
		
	case pg_query.SubLinkType_EXPR_SUBLINK:
		// Scalar subquery
		if len(subRows) != 1 {
			return false
		}
		// TODO: Implement proper scalar subquery comparison
		return true
	}
	return false
}

func groupRows(rows []storage.Row, groupClause []*pg_query.Node) map[string][]storage.Row {
	groups := make(map[string][]storage.Row)

	for _, row := range rows {
		groupKey := buildGroupKey(row, groupClause)
		groups[groupKey] = append(groups[groupKey], row)
	}

	return groups
}

func buildGroupKey(row storage.Row, groupClause []*pg_query.Node) string {
	var keyParts []string
	for _, groupNode := range groupClause {
		value := extractGroupValue(row, groupNode)
		keyParts = append(keyParts, fmt.Sprintf("%v", value))
	}
	return strings.Join(keyParts, "|")
}

func extractGroupValue(row storage.Row, node *pg_query.Node) interface{} {
	// Extract column reference from GROUP BY expression
	if colRef, ok := node.Node.(*pg_query.Node_ColumnRef); ok {
		if len(colRef.ColumnRef.Fields) > 0 {
			if str, ok := colRef.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
				return row[str.String_.Sval]
			}
		}
	}
	return nil
}

func processSelectList(ctx *QueryContext, targetList []*pg_query.Node, allRows []storage.Row, groupedRows map[string][]storage.Row, groupClause []*pg_query.Node) ([]string, [][]interface{}, error) {
	var columns []string
	var resultRows [][]interface{}

	// First pass: determine all columns (especially important for SELECT *)
	columns = determineAllColumns(ctx, targetList, allRows, groupedRows)

	if groupedRows != nil {
		// Process grouped results
		for _, groupRows := range groupedRows {
			// Handle empty groups (e.g., COUNT on empty table)
			var sampleRow storage.Row
			if len(groupRows) > 0 {
				sampleRow = groupRows[0]
			} else {
				sampleRow = make(storage.Row)
			}
			resultRow := processSelectTargetsWithColumns(ctx, targetList, sampleRow, groupRows, true, columns)
			resultRows = append(resultRows, resultRow)
		}
	} else {
		// Process non-grouped results
		for _, row := range allRows {
			resultRow := processSelectTargetsWithColumns(ctx, targetList, row, allRows, false, columns)
			resultRows = append(resultRows, resultRow)
		}
	}

	return columns, resultRows, nil
}

func determineAllColumns(ctx *QueryContext, targetList []*pg_query.Node, allRows []storage.Row, groupedRows map[string][]storage.Row) []string {
	var columns []string
	
	// Check if we have SELECT *
	hasStar := false
	for _, target := range targetList {
		if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok {
			if resTarget.ResTarget.Val != nil {
				if colRef, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_ColumnRef); ok {
					if len(colRef.ColumnRef.Fields) > 0 {
						if _, isStar := colRef.ColumnRef.Fields[0].Node.(*pg_query.Node_AStar); isStar {
							hasStar = true
							break
						}
					}
				}
			}
		}
	}
	
	if hasStar {
		// For SELECT *, we need all columns from all rows
		colMap := make(map[string]bool)
		
		// Get columns from table definitions first
		for tableName := range ctx.tables {
			tableCols := ctx.metaStore.GetTableColumns(tableName)
			for _, col := range tableCols {
				colMap[col] = true
			}
		}
		
		// Then add any additional columns from the actual rows
		rows := allRows
		if groupedRows != nil {
			rows = nil
			for _, groupRows := range groupedRows {
				rows = append(rows, groupRows...)
			}
		}
		
		for _, row := range rows {
			for col := range row {
				colMap[col] = true
			}
		}
		
		// Get ordered column list
		var orderedCols []string
		// First add columns from table definitions
		for tableName := range ctx.tables {
			tableCols := ctx.metaStore.GetTableColumns(tableName)
			for _, col := range tableCols {
				if colMap[col] {
					orderedCols = append(orderedCols, col)
					delete(colMap, col)
				}
			}
		}
		// Then add remaining columns in sorted order
		var remainingCols []string
		for col := range colMap {
			remainingCols = append(remainingCols, col)
		}
		sort.Strings(remainingCols)
		orderedCols = append(orderedCols, remainingCols...)
		
		columns = orderedCols
	} else {
		// For specific columns, just process the target list
		for _, target := range targetList {
			if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok {
				colName := resTarget.ResTarget.Name
				if colName == "" && resTarget.ResTarget.Val != nil {
					colName = extractColumnName(resTarget.ResTarget.Val)
				}
				columns = append(columns, colName)
			}
		}
	}
	
	return columns
}

func processSelectTargetsWithColumns(ctx *QueryContext, targetList []*pg_query.Node, currentRow storage.Row, groupRows []storage.Row, isGrouped bool, columns []string) []interface{} {
	values := make([]interface{}, len(columns))
	
	// Check if we have SELECT *
	hasStar := false
	for _, target := range targetList {
		if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok {
			if resTarget.ResTarget.Val != nil {
				if colRef, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_ColumnRef); ok {
					if len(colRef.ColumnRef.Fields) > 0 {
						if _, isStar := colRef.ColumnRef.Fields[0].Node.(*pg_query.Node_AStar); isStar {
							hasStar = true
							break
						}
					}
				}
			}
		}
	}
	
	if hasStar {
		// For SELECT *, fill in values for all columns
		for i, col := range columns {
			if val, exists := currentRow[col]; exists {
				values[i] = val
			} else {
				values[i] = nil
			}
		}
	} else {
		// For specific columns, evaluate each expression
		i := 0
		for _, target := range targetList {
			if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok {
				value, _ := evaluateSelectExpression(resTarget.ResTarget, currentRow, groupRows, isGrouped, ctx)
				if i < len(values) {
					values[i] = value
					i++
				}
			}
		}
	}
	
	return values
}


func evaluateSelectExpression(resTarget *pg_query.ResTarget, currentRow storage.Row, groupRows []storage.Row, isGrouped bool, ctx *QueryContext) (interface{}, string) {
	// Determine column name
	colName := resTarget.Name
	if colName == "" && resTarget.Val != nil {
		colName = extractColumnName(resTarget.Val)
	}

	// Evaluate expression
	if resTarget.Val != nil {
		switch val := resTarget.Val.Node.(type) {
		case *pg_query.Node_FuncCall:
			// Handle aggregate functions
			result := evaluateAggregateFunction(val.FuncCall, groupRows)
			if colName == "" {
				colName = strings.ToLower(val.FuncCall.Funcname[0].Node.(*pg_query.Node_String_).String_.Sval)
			}
			return result, colName
		case *pg_query.Node_ColumnRef:
			// Handle column reference with potential table alias
			var columnName string
			fields := val.ColumnRef.Fields
			
			if len(fields) == 2 {
				// table.column format - use the column name (second field)
				if str, ok := fields[1].Node.(*pg_query.Node_String_); ok {
					columnName = str.String_.Sval
				}
			} else if len(fields) == 1 {
				// Just column name
				if str, ok := fields[0].Node.(*pg_query.Node_String_); ok {
					columnName = str.String_.Sval
				}
			}
			
			if columnName != "" {
				return currentRow[columnName], columnName
			}
		case *pg_query.Node_AConst:
			// Handle constant
			return extractAConstValue(val.AConst), colName
		case *pg_query.Node_SubLink:
			// Handle subquery in SELECT
			subRows, _ := executeSubquery(val.SubLink.Subselect, ctx)
			if len(subRows) > 0 && len(subRows[0]) > 0 {
				for _, v := range subRows[0] {
					return v, colName
				}
			}
			return nil, colName
		}
	}

	return nil, colName
}

func evaluateAggregateFunction(funcCall *pg_query.FuncCall, rows []storage.Row) interface{} {
	if len(funcCall.Funcname) == 0 {
		return nil
	}

	funcName := strings.ToUpper(funcCall.Funcname[0].Node.(*pg_query.Node_String_).String_.Sval)
	
	// Extract column name for aggregate
	var colName string
	if len(funcCall.Args) > 0 {
		if colRef, ok := funcCall.Args[0].Node.(*pg_query.Node_ColumnRef); ok {
			if len(colRef.ColumnRef.Fields) > 0 {
				if str, ok := colRef.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
					colName = str.String_.Sval
				}
			}
		}
	}

	switch funcName {
	case "COUNT":
		if colName == "" || (len(funcCall.Args) > 0 && isStarExpr(funcCall.Args[0])) {
			return len(rows)
		}
		count := 0
		for _, row := range rows {
			if row[colName] != nil {
				count++
			}
		}
		return count

	case "SUM":
		var sum float64
		for _, row := range rows {
			if val := row[colName]; val != nil {
				if num, err := toFloat64(val); err == nil {
					sum += num
				}
			}
		}
		return sum

	case "AVG":
		var sum float64
		count := 0
		for _, row := range rows {
			if val := row[colName]; val != nil {
				if num, err := toFloat64(val); err == nil {
					sum += num
					count++
				}
			}
		}
		if count == 0 {
			return nil
		}
		return sum / float64(count)

	case "MAX":
		var max interface{}
		for _, row := range rows {
			if val := row[colName]; val != nil {
				if max == nil || compareValues(fmt.Sprintf("%v", val), ">", max) {
					max = val
				}
			}
		}
		return max

	case "MIN":
		var min interface{}
		for _, row := range rows {
			if val := row[colName]; val != nil {
				if min == nil || compareValues(fmt.Sprintf("%v", val), "<", min) {
					min = val
				}
			}
		}
		return min
	}

	return nil
}

func isStarExpr(node *pg_query.Node) bool {
	if colRef, ok := node.Node.(*pg_query.Node_ColumnRef); ok {
		if len(colRef.ColumnRef.Fields) > 0 {
			_, isStar := colRef.ColumnRef.Fields[0].Node.(*pg_query.Node_AStar)
			return isStar
		}
	}
	return false
}

func toFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert to float64")
	}
}

func extractColumnName(node *pg_query.Node) string {
	switch n := node.Node.(type) {
	case *pg_query.Node_ColumnRef:
		var parts []string
		for _, field := range n.ColumnRef.Fields {
			if str, ok := field.Node.(*pg_query.Node_String_); ok {
				parts = append(parts, str.String_.Sval)
			}
		}
		if len(parts) > 0 {
			return parts[len(parts)-1] // Return last part (column name)
		}
	case *pg_query.Node_FuncCall:
		if len(n.FuncCall.Funcname) > 0 {
			return strings.ToLower(n.FuncCall.Funcname[0].Node.(*pg_query.Node_String_).String_.Sval)
		}
	}
	return "?column?"
}

func filterResultRows(rows [][]interface{}, columns []string, havingClause *pg_query.Node) [][]interface{} {
	var filtered [][]interface{}
	for _, row := range rows {
		// Convert row to map for evaluation
		rowMap := make(storage.Row)
		for i, col := range columns {
			if i < len(row) {
				rowMap[col] = row[i]
			}
		}
		if evaluatePgWhere(rowMap, havingClause) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func sortRows(rows [][]interface{}, columns []string, sortClause []*pg_query.Node) [][]interface{} {
	// Simplified sorting implementation
	// In a real implementation, this would properly handle multiple sort keys and directions
	return rows
}

func applyLimitOffset(rows [][]interface{}, limitCount, limitOffset *pg_query.Node) [][]interface{} {
	offset := 0
	limit := len(rows)

	if limitOffset != nil {
		if constNode, ok := limitOffset.Node.(*pg_query.Node_AConst); ok {
			if val, ok := constNode.AConst.Val.(*pg_query.A_Const_Ival); ok {
				offset = int(val.Ival.Ival)
			}
		}
	}

	if limitCount != nil {
		if constNode, ok := limitCount.Node.(*pg_query.Node_AConst); ok {
			if val, ok := constNode.AConst.Val.(*pg_query.A_Const_Ival); ok {
				limit = int(val.Ival.Ival)
			}
		}
	}

	if offset >= len(rows) {
		return [][]interface{}{}
	}

	end := offset + limit
	if end > len(rows) {
		end = len(rows)
	}

	return rows[offset:end]
}

func extractValueFromNode(row storage.Row, node *pg_query.Node) interface{} {
	switch n := node.Node.(type) {
	case *pg_query.Node_ColumnRef:
		if len(n.ColumnRef.Fields) > 0 {
			if str, ok := n.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
				return row[str.String_.Sval]
			}
		}
	case *pg_query.Node_AConst:
		return extractAConstValue(n.AConst)
	}
	return nil
}

func compareValues(left, op string, right interface{}) bool {
	// Reuse existing comparison logic
	return compareValuesPg(left, op, right)
}

func applyDistinct(rows [][]interface{}) [][]interface{} {
	if len(rows) == 0 {
		return rows
	}
	
	var distinctRows [][]interface{}
	seen := make(map[string]bool)
	
	for _, row := range rows {
		// Create a key for the row, treating NULLs consistently
		key := rowToKey(row)
		if !seen[key] {
			seen[key] = true
			distinctRows = append(distinctRows, row)
		}
	}
	
	return distinctRows
}

func rowToKey(row []interface{}) string {
	var parts []string
	for _, val := range row {
		if val == nil {
			// Use a special marker for NULL that won't conflict with actual values
			parts = append(parts, "\x00NULL\x00")
		} else {
			parts = append(parts, fmt.Sprintf("%v", val))
		}
	}
	return strings.Join(parts, "\x01")
}

func executeSetOperation(stmt *pg_query.SelectStmt, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	// Execute left side query
	var leftColumns []string
	var leftRows [][]interface{}
	var err error
	
	if stmt.Larg != nil {
		leftColumns, leftRows, _, err = executePgSelectAdvanced(stmt.Larg, dataStore, metaStore)
		if err != nil {
			return nil, nil, "", err
		}
	}
	
	// Execute right side query
	var rightColumns []string
	var rightRows [][]interface{}
	
	if stmt.Rarg != nil {
		rightColumns, rightRows, _, err = executePgSelectAdvanced(stmt.Rarg, dataStore, metaStore)
		if err != nil {
			return nil, nil, "", err
		}
	}
	
	// Validate column counts match
	if len(leftColumns) != len(rightColumns) {
		return nil, nil, "", fmt.Errorf("each UNION query must have the same number of columns")
	}
	
	// Use column names from the first query
	columns := leftColumns
	
	// Perform set operation based on type
	var resultRows [][]interface{}
	
	switch stmt.Op {
	case pg_query.SetOperation_SETOP_UNION:
		// UNION removes duplicates (UNION ALL would keep them)
		if stmt.All {
			// UNION ALL - keep all rows
			resultRows = append(resultRows, leftRows...)
			resultRows = append(resultRows, rightRows...)
		} else {
			// UNION - remove duplicates
			seen := make(map[string]bool)
			
			// Add left rows
			for _, row := range leftRows {
				key := rowToKey(row)
				if !seen[key] {
					seen[key] = true
					resultRows = append(resultRows, row)
				}
			}
			
			// Add right rows
			for _, row := range rightRows {
				key := rowToKey(row)
				if !seen[key] {
					seen[key] = true
					resultRows = append(resultRows, row)
				}
			}
		}
		
	case pg_query.SetOperation_SETOP_INTERSECT:
		// INTERSECT - only rows in both
		leftSet := make(map[string]bool)
		for _, row := range leftRows {
			leftSet[rowToKey(row)] = true
		}
		
		seen := make(map[string]bool)
		for _, row := range rightRows {
			key := rowToKey(row)
			if leftSet[key] && !seen[key] {
				seen[key] = true
				resultRows = append(resultRows, row)
			}
		}
		
	case pg_query.SetOperation_SETOP_EXCEPT:
		// EXCEPT - rows in left but not in right
		rightSet := make(map[string]bool)
		for _, row := range rightRows {
			rightSet[rowToKey(row)] = true
		}
		
		seen := make(map[string]bool)
		for _, row := range leftRows {
			key := rowToKey(row)
			if !rightSet[key] && !seen[key] {
				seen[key] = true
				resultRows = append(resultRows, row)
			}
		}
		
	default:
		return nil, nil, "", fmt.Errorf("unsupported set operation")
	}
	
	// Apply ORDER BY if present
	if len(stmt.SortClause) > 0 {
		resultRows = sortRows(resultRows, columns, stmt.SortClause)
	}
	
	// Apply LIMIT and OFFSET if present
	if stmt.LimitCount != nil || stmt.LimitOffset != nil {
		resultRows = applyLimitOffset(resultRows, stmt.LimitCount, stmt.LimitOffset)
	}
	
	return columns, resultRows, fmt.Sprintf("SELECT %d", len(resultRows)), nil
}
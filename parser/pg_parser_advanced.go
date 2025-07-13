package parser

import (
	"fmt"
	"log"
	"sort"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
	"github.com/satetsu888/vsql/storage"
)

type QueryContext struct {
	dataStore    *storage.DataStore
	metaStore    *storage.MetaStore
	tables       map[string]*TableContext
	outerTables  map[string]*TableContext  // Tables from outer queries
	subqueries   map[string][]storage.Row
	aggregations map[string]interface{}
	currentJoinContext *JoinContext  // Track current join context
	currentRow   storage.Row          // Current row for correlated subqueries
	outerRows    []storage.Row        // Stack of rows from outer queries
}

type TableContext struct {
	name  string
	alias string
	rows  []storage.Row
}

type JoinContext struct {
	leftAlias  string
	rightAlias string
	leftRow    storage.Row
	rightRow   storage.Row
}

func hasAggregateFunctions(targetList []*pg_query.Node) bool {
	for _, target := range targetList {
		if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok {
			if resTarget.ResTarget.Val != nil {
				if funcCall, isFuncCall := resTarget.ResTarget.Val.Node.(*pg_query.Node_FuncCall); isFuncCall {
					funcName := getFunctionName(funcCall.FuncCall)
					if isAggregateFunction(funcName) {
						return true
					}
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
		outerTables:  make(map[string]*TableContext),
		subqueries:   make(map[string][]storage.Row),
		aggregations: make(map[string]interface{}),
		outerRows:    []storage.Row{},
	}

	return executePgSelectWithContext(stmt, ctx)
}

func executePgSelectWithContext(stmt *pg_query.SelectStmt, ctx *QueryContext) ([]string, [][]interface{}, string, error) {
	// Handle UNION/INTERSECT/EXCEPT queries
	if stmt.Op != pg_query.SetOperation_SETOP_NONE {
		return executeSetOperation(stmt, ctx.dataStore, ctx.metaStore)
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
		// Use ordered groups if available
		var orderedGroups [][]storage.Row
		if groupOrder, ok := ctx.aggregations["__groupOrder__"].([][]storage.Row); ok {
			orderedGroups = groupOrder
		} else {
			// Fallback to unordered
			for _, group := range groupedRows {
				orderedGroups = append(orderedGroups, group)
			}
		}
		resultRows = filterResultRowsWithGroups(resultRows, columns, stmt.HavingClause, orderedGroups)
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
		// PostgreSQL-compatible: SELECT without FROM returns a single empty row
		return []storage.Row{make(storage.Row)}, nil
	}

	// If there's only one item in FROM clause, process it directly
	if len(fromClause) == 1 {
		return processFromNode(ctx, fromClause[0])
	}

	// Multiple items in FROM clause - handle as CROSS JOIN
	var result []storage.Row
	for i, fromNode := range fromClause {
		rows, err := processFromNode(ctx, fromNode)
		if err != nil {
			return nil, err
		}
		
		if i == 0 {
			result = rows
		} else {
			// Perform CROSS JOIN with previous results
			var newResult []storage.Row
			for _, r1 := range result {
				for _, r2 := range rows {
					merged := make(storage.Row)
					for k, v := range r1 {
						merged[k] = v
					}
					for k, v := range r2 {
						merged[k] = v
					}
					newResult = append(newResult, merged)
				}
			}
			result = newResult
		}
	}

	return result, nil
}

func processFromNode(ctx *QueryContext, node *pg_query.Node) ([]storage.Row, error) {
	switch n := node.Node.(type) {
	case *pg_query.Node_RangeVar:
		realTableName := n.RangeVar.Relname
		aliasName := realTableName
		if n.RangeVar.Alias != nil && n.RangeVar.Alias.Aliasname != "" {
			aliasName = n.RangeVar.Alias.Aliasname
		}

		var rows []storage.Row
		table, exists := ctx.dataStore.GetTable(realTableName)
		if exists {
			rows = table.GetRows()
		} else {
			// Table doesn't exist - return empty row set
			rows = []storage.Row{}
		}
		
		// Store table context with both real name and alias
		ctx.tables[aliasName] = &TableContext{
			name:  realTableName,
			alias: aliasName,
			rows:  rows,
		}
		if aliasName != realTableName {
			ctx.tables[realTableName] = ctx.tables[aliasName]
		}
		
		// If we have an alias, enrich rows with qualified column names
		if aliasName != "" {
			enrichedRows := make([]storage.Row, len(rows))
			for i, row := range rows {
				enrichedRow := make(storage.Row)
				for k, v := range row {
					// Add unqualified name
					enrichedRow[k] = v
					// Add qualified name
					enrichedRow[aliasName+"."+k] = v
				}
				enrichedRows[i] = enrichedRow
			}
			return enrichedRows, nil
		}
		
		return rows, nil
	case *pg_query.Node_JoinExpr:
		// Handle JOIN
		// fmt.Printf("DEBUG processFromNode: Processing JoinExpr\n")
		// Get left table info
		leftAlias := extractTableAlias(n.JoinExpr.Larg)
		// fmt.Printf("DEBUG processFromNode: Left alias extracted: '%s'\n", leftAlias)
		leftRows, err := processFromNode(ctx, n.JoinExpr.Larg)
		if err != nil {
			return nil, err
		}
		// fmt.Printf("DEBUG processFromNode: Left rows count: %d\n", len(leftRows))
		
		// Get right table info
		rightAlias := extractTableAlias(n.JoinExpr.Rarg)
		// fmt.Printf("DEBUG processFromNode: Right alias extracted: '%s'\n", rightAlias)
		rightRows, err := processFromNode(ctx, n.JoinExpr.Rarg)
		if err != nil {
			return nil, err
		}
		// fmt.Printf("DEBUG processFromNode: Right rows count: %d\n", len(rightRows))
		
		return performJoinWithContext(leftRows, rightRows, n.JoinExpr, leftAlias, rightAlias, ctx), nil
	case *pg_query.Node_RangeSubselect:
		// Execute the subquery
		rows, err := executeSubquery(n.RangeSubselect.Subquery, ctx)
		if err != nil {
			return nil, err
		}
		
		// Handle alias for the subquery result
		if n.RangeSubselect.Alias != nil && n.RangeSubselect.Alias.Aliasname != "" {
			aliasName := n.RangeSubselect.Alias.Aliasname
			
			// Create a table context for the subquery result
			ctx.tables[aliasName] = &TableContext{
				name:  aliasName,
				alias: aliasName, 
				rows:  rows,
			}
		}
		
		return rows, nil
	}
	log.Printf("WARNING: Unsupported FROM node type in query. Returning empty result.\n")
	return []storage.Row{}, nil
}

func extractTableAlias(node *pg_query.Node) string {
	if node == nil {
		return ""
	}
	
	switch n := node.Node.(type) {
	case *pg_query.Node_RangeVar:
		if n.RangeVar.Alias != nil && n.RangeVar.Alias.Aliasname != "" {
			// fmt.Printf("DEBUG extractTableAlias: RangeVar with alias '%s' (table name '%s')\n", 
			//	n.RangeVar.Alias.Aliasname, n.RangeVar.Relname)
			return n.RangeVar.Alias.Aliasname
		}
		// fmt.Printf("DEBUG extractTableAlias: RangeVar without alias, using table name '%s'\n", n.RangeVar.Relname)
		return n.RangeVar.Relname
	case *pg_query.Node_JoinExpr:
		// For nested joins, this gets more complex
		// For now, just return empty
		// fmt.Printf("DEBUG extractTableAlias: JoinExpr node (nested join) - returning empty\n")
		return ""
	}
	// fmt.Printf("DEBUG extractTableAlias: Unknown node type - returning empty\n")
	return ""
}

func performJoinWithContext(leftRows, rightRows []storage.Row, joinExpr *pg_query.JoinExpr, leftAlias, rightAlias string, ctx *QueryContext) []storage.Row {
	// Store the aliases in context for use during evaluation
	joinCtx := &JoinContext{
		leftAlias:  leftAlias,
		rightAlias: rightAlias,
	}
	ctx.currentJoinContext = joinCtx
	
	result := performJoinWithAliases(leftRows, rightRows, joinExpr, leftAlias, rightAlias, ctx)
	
	// Clear join context after use
	ctx.currentJoinContext = nil
	
	return result
}

func performJoin(leftRows, rightRows []storage.Row, joinExpr *pg_query.JoinExpr, ctx *QueryContext) []storage.Row {
	// Delegate to performJoinWithAliases with empty aliases
	return performJoinWithAliases(leftRows, rightRows, joinExpr, "", "", ctx)
}

func performJoinWithAliases(leftRows, rightRows []storage.Row, joinExpr *pg_query.JoinExpr, leftAlias, rightAlias string, ctx *QueryContext) []storage.Row {
	var result []storage.Row

	// DEBUG: Print join info
	// // fmt.Printf("DEBUG performJoinWithAliases: leftAlias='%s', rightAlias='%s', leftRows=%d, rightRows=%d\n", 
	// 	leftAlias, rightAlias, len(leftRows), len(rightRows))
	// if joinExpr.Quals != nil {
	// 	// fmt.Printf("DEBUG performJoinWithAliases: Has join condition\n")
	// } else {
	// 	// fmt.Printf("DEBUG performJoinWithAliases: No join condition (CROSS JOIN)\n")
	// }

	switch joinExpr.Jointype {
	case pg_query.JoinType_JOIN_INNER:
		// INNER JOIN
		for _, leftRow := range leftRows {
			for _, rightRow := range rightRows {
				if joinExpr.Quals == nil || evaluateJoinCondition(leftRow, rightRow, joinExpr.Quals, ctx) {
					mergedRow := mergeRowsWithAliases(leftRow, rightRow, leftAlias, rightAlias)
					// DEBUG: Print merged row
					// // fmt.Printf("DEBUG performJoinWithAliases: Merged row: %v\n", mergedRow)
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
					mergedRow := mergeRowsWithAliases(leftRow, rightRow, leftAlias, rightAlias)
					result = append(result, mergedRow)
					matched = true
				}
			}
			if !matched {
				// Add left row with NULLs for right columns
				// We need to create a dummy right row with all NULL values
				nullRightRow := make(storage.Row)
				
				// Get column names from the right table
				if rightAlias != "" && ctx.tables[rightAlias] != nil {
					// Get columns from metastore
					rightTableName := ctx.tables[rightAlias].name
					rightColumns := ctx.metaStore.GetTableColumns(rightTableName)
					for _, col := range rightColumns {
						nullRightRow[col] = nil
					}
				} else if len(rightRows) > 0 {
					// If no alias or metadata, infer columns from first row
					for k := range rightRows[0] {
						nullRightRow[k] = nil
					}
				}
				
				mergedRow := mergeRowsWithAliases(leftRow, nullRightRow, leftAlias, rightAlias)
				result = append(result, mergedRow)
			}
		}
	case pg_query.JoinType_JOIN_RIGHT:
		// RIGHT JOIN
		for _, rightRow := range rightRows {
			matched := false
			for _, leftRow := range leftRows {
				if joinExpr.Quals == nil || evaluateJoinCondition(leftRow, rightRow, joinExpr.Quals, ctx) {
					mergedRow := mergeRowsWithAliases(leftRow, rightRow, leftAlias, rightAlias)
					result = append(result, mergedRow)
					matched = true
				}
			}
			if !matched {
				// Add right row with NULLs for left columns
				// We need to create a dummy left row with all NULL values
				nullLeftRow := make(storage.Row)
				
				// Get column names from the left table
				if leftAlias != "" && ctx.tables[leftAlias] != nil {
					// Get columns from metastore
					leftTableName := ctx.tables[leftAlias].name
					leftColumns := ctx.metaStore.GetTableColumns(leftTableName)
					for _, col := range leftColumns {
						nullLeftRow[col] = nil
					}
				} else if len(leftRows) > 0 {
					// If no alias or metadata, infer columns from first row
					for k := range leftRows[0] {
						nullLeftRow[k] = nil
					}
				}
				
				mergedRow := mergeRowsWithAliases(nullLeftRow, rightRow, leftAlias, rightAlias)
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
					mergedRow := mergeRowsWithAliases(leftRow, rightRow, leftAlias, rightAlias)
					result = append(result, mergedRow)
					leftMatched[i] = true
					rightMatched[j] = true
				}
			}
		}

		// Add unmatched left rows
		for i, leftRow := range leftRows {
			if !leftMatched[i] {
				// Create NULL right row
				nullRightRow := make(storage.Row)
				if rightAlias != "" && ctx.tables[rightAlias] != nil {
					rightTableName := ctx.tables[rightAlias].name
					rightColumns := ctx.metaStore.GetTableColumns(rightTableName)
					for _, col := range rightColumns {
						nullRightRow[col] = nil
					}
				} else if len(rightRows) > 0 {
					for k := range rightRows[0] {
						nullRightRow[k] = nil
					}
				}
				mergedRow := mergeRowsWithAliases(leftRow, nullRightRow, leftAlias, rightAlias)
				result = append(result, mergedRow)
			}
		}

		// Add unmatched right rows
		for j, rightRow := range rightRows {
			if !rightMatched[j] {
				// Create NULL left row
				nullLeftRow := make(storage.Row)
				if leftAlias != "" && ctx.tables[leftAlias] != nil {
					leftTableName := ctx.tables[leftAlias].name
					leftColumns := ctx.metaStore.GetTableColumns(leftTableName)
					for _, col := range leftColumns {
						nullLeftRow[col] = nil
					}
				} else if len(leftRows) > 0 {
					for k := range leftRows[0] {
						nullLeftRow[k] = nil
					}
				}
				mergedRow := mergeRowsWithAliases(nullLeftRow, rightRow, leftAlias, rightAlias)
				result = append(result, mergedRow)
			}
		}
	default:
		// Handle other join types including CROSS JOIN
		// For CROSS JOIN, we join every row from left with every row from right
		// without any join condition
		for _, leftRow := range leftRows {
			for _, rightRow := range rightRows {
				mergedRow := mergeRowsWithAliases(leftRow, rightRow, leftAlias, rightAlias)
				result = append(result, mergedRow)
			}
		}
	}

	return result
}

func evaluateJoinCondition(leftRow, rightRow storage.Row, condition *pg_query.Node, ctx *QueryContext) bool {
	// For JOIN conditions, we need to handle qualified column references specially
	return evaluateQualifiedExpr(leftRow, rightRow, condition, ctx)
}

// evaluateQualifiedExpr handles expressions with table-qualified column references
func evaluateQualifiedExpr(leftRow, rightRow storage.Row, expr *pg_query.Node, ctx *QueryContext) bool {
	switch e := expr.Node.(type) {
	case *pg_query.Node_AExpr:
		return evaluateQualifiedAExpr(leftRow, rightRow, e.AExpr, ctx)
	case *pg_query.Node_BoolExpr:
		return evaluateQualifiedBoolExpr(leftRow, rightRow, e.BoolExpr, ctx)
	default:
		// Fall back to regular evaluation on merged row
		mergedRow := mergeRows(leftRow, rightRow)
		return evaluatePgWhere(mergedRow, expr)
	}
}

func evaluateQualifiedAExpr(leftRow, rightRow storage.Row, expr *pg_query.A_Expr, ctx *QueryContext) bool {
	// Extract values from qualified column references
	leftVal := extractQualifiedValue(leftRow, rightRow, expr.Lexpr, ctx)
	rightVal := extractQualifiedValue(leftRow, rightRow, expr.Rexpr, ctx)
	
	// Get operator
	opName := ""
	if len(expr.Name) > 0 {
		if str, ok := expr.Name[0].Node.(*pg_query.Node_String_); ok {
			opName = str.String_.Sval
		}
	}
	
	// DEBUG: Print comparison
	result := compareValuesPg(leftVal, opName, rightVal)
	// fmt.Printf("DEBUG evaluateQualifiedAExpr: %v %s %v = %v\n", leftVal, opName, rightVal, result)
	
	return result
}

func evaluateQualifiedBoolExpr(leftRow, rightRow storage.Row, expr *pg_query.BoolExpr, ctx *QueryContext) bool {
	switch expr.Boolop {
	case pg_query.BoolExprType_AND_EXPR:
		for _, arg := range expr.Args {
			if !evaluateQualifiedExpr(leftRow, rightRow, arg, ctx) {
				return false
			}
		}
		return true
	case pg_query.BoolExprType_OR_EXPR:
		for _, arg := range expr.Args {
			if evaluateQualifiedExpr(leftRow, rightRow, arg, ctx) {
				return true
			}
		}
		return false
	case pg_query.BoolExprType_NOT_EXPR:
		if len(expr.Args) > 0 {
			return !evaluateQualifiedExpr(leftRow, rightRow, expr.Args[0], ctx)
		}
		return true
	}
	return true
}

func extractQualifiedValue(leftRow, rightRow storage.Row, node *pg_query.Node, ctx *QueryContext) interface{} {
	if node == nil {
		return nil
	}
	
	// fmt.Printf("DEBUG extractQualifiedValue: node type = %T\n", node.Node)
	
	switch n := node.Node.(type) {
	case *pg_query.Node_ColumnRef:
		// fmt.Printf("DEBUG extractQualifiedValue: ColumnRef with %d fields\n", len(n.ColumnRef.Fields))
		// Handle qualified column references (table.column)
		if len(n.ColumnRef.Fields) >= 2 {
			// table.column format
			tableName := ""
			columnName := ""
			
			if str, ok := n.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
				tableName = str.String_.Sval
			}
			if str, ok := n.ColumnRef.Fields[1].Node.(*pg_query.Node_String_); ok {
				columnName = str.String_.Sval
			}
			
			// DEBUG: Print qualified column lookup
			// fmt.Printf("DEBUG extractQualifiedValue: Looking for %s.%s\n", tableName, columnName)
			// fmt.Printf("DEBUG extractQualifiedValue: leftRow=%v\n", leftRow)
			// fmt.Printf("DEBUG extractQualifiedValue: rightRow=%v\n", rightRow)
			
			// Check which row to use based on table name/alias
			if ctx.currentJoinContext != nil {
				// fmt.Printf("DEBUG extractQualifiedValue: currentJoinContext leftAlias='%s', rightAlias='%s'\n",
				//	ctx.currentJoinContext.leftAlias, ctx.currentJoinContext.rightAlias)
				// fmt.Printf("DEBUG extractQualifiedValue: Comparing tableName='%s' with leftAlias='%s' and rightAlias='%s'\n",
				//	tableName, ctx.currentJoinContext.leftAlias, ctx.currentJoinContext.rightAlias)
				if tableName == ctx.currentJoinContext.leftAlias {
					if val, exists := leftRow[columnName]; exists {
						// fmt.Printf("DEBUG extractQualifiedValue: Found in leftRow: %v\n", val)
						return val
					}
					// fmt.Printf("DEBUG extractQualifiedValue: Not found in leftRow\n")
				} else if tableName == ctx.currentJoinContext.rightAlias {
					if val, exists := rightRow[columnName]; exists {
						// fmt.Printf("DEBUG extractQualifiedValue: Found in rightRow: %v\n", val)
						return val
					}
					// fmt.Printf("DEBUG extractQualifiedValue: Not found in rightRow\n")
				} else {
					// The table might be from a previous join, check if it exists in the qualified columns
					// fmt.Printf("DEBUG extractQualifiedValue: Table '%s' not in current join context, checking qualified columns\n", tableName)
					qualifiedName := tableName + "." + columnName
					if val, exists := leftRow[qualifiedName]; exists {
						// fmt.Printf("DEBUG extractQualifiedValue: Found as qualified column '%s' in leftRow: %v\n", qualifiedName, val)
						return val
					}
					// Also check for the unqualified column in case it's from the nested join
					if val, exists := leftRow[columnName]; exists {
						// fmt.Printf("DEBUG extractQualifiedValue: Found unqualified column '%s' in leftRow: %v\n", columnName, val)
						return val
					}
				}
			} else {
				// Fallback to checking both rows
				if val, exists := leftRow[columnName]; exists {
					// fmt.Printf("DEBUG extractQualifiedValue: Found in leftRow (fallback): %v\n", val)
					return val
				}
				if val, exists := rightRow[columnName]; exists {
					// fmt.Printf("DEBUG extractQualifiedValue: Found in rightRow (fallback): %v\n", val)
					return val
				}
			}
		} else if len(n.ColumnRef.Fields) == 1 {
			// Unqualified column - check both rows
			if str, ok := n.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
				columnName := str.String_.Sval
				// fmt.Printf("DEBUG extractQualifiedValue: Looking for unqualified column %s\n", columnName)
				// First check left row, then right row
				if val, exists := leftRow[columnName]; exists {
					// fmt.Printf("DEBUG extractQualifiedValue: Found in leftRow: %v\n", val)
					return val
				}
				if val, exists := rightRow[columnName]; exists {
					// fmt.Printf("DEBUG extractQualifiedValue: Found in rightRow: %v\n", val)
					return val
				}
				// fmt.Printf("DEBUG extractQualifiedValue: Not found in either row\n")
			}
		}
	case *pg_query.Node_AConst:
		return extractAConstValue(n.AConst)
	}
	return nil
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

func mergeRowsWithAliases(left, right storage.Row, leftAlias, rightAlias string) storage.Row {
	merged := make(storage.Row)
	
	// DEBUG: Print merge input
	// fmt.Printf("DEBUG mergeRowsWithAliases: leftAlias='%s', rightAlias='%s'\n", leftAlias, rightAlias)
	// fmt.Printf("DEBUG mergeRowsWithAliases: leftRow=%v\n", left)
	// fmt.Printf("DEBUG mergeRowsWithAliases: rightRow=%v\n", right)
	
	// Check for column conflicts
	conflicts := make(map[string]bool)
	for k := range left {
		if _, exists := right[k]; exists {
			conflicts[k] = true
		}
	}
	
	// Add left row columns
	for k, v := range left {
		// Skip already qualified names when processing
		if strings.Contains(k, ".") {
			merged[k] = v
			continue
		}
		
		// Always add the original key
		merged[k] = v
		
		// Always add the qualified version when we have an alias
		if leftAlias != "" {
			merged[leftAlias+"."+k] = v
		}
	}
	
	// Add right row columns
	for k, v := range right {
		// Skip already qualified names when processing
		if strings.Contains(k, ".") {
			merged[k] = v
			continue
		}
		
		// Only add unqualified if not already present (left takes precedence for unqualified names)
		if _, exists := merged[k]; !exists {
			merged[k] = v
		}
		
		// Always add the qualified version when we have an alias
		if rightAlias != "" {
			merged[rightAlias+"."+k] = v
		}
	}
	
	// DEBUG: Print merged result
	// fmt.Printf("DEBUG mergeRowsWithAliases: merged=%v\n", merged)
	
	return merged
}

func executeSubquery(subquery *pg_query.Node, ctx *QueryContext) ([]storage.Row, error) {
	if selectStmt, ok := subquery.Node.(*pg_query.Node_SelectStmt); ok {
		// Create a new context for the subquery that inherits currentRow and tables from outer context
		subCtx := &QueryContext{
			dataStore:    ctx.dataStore,
			metaStore:    ctx.metaStore,
			tables:       make(map[string]*TableContext),
			outerTables:  make(map[string]*TableContext),
			subqueries:   make(map[string][]storage.Row),
			aggregations: make(map[string]interface{}),
			currentRow:   ctx.currentRow, // Pass the outer query's row
			outerRows:    make([]storage.Row, len(ctx.outerRows)),
		}
		
		// Copy outer rows stack
		copy(subCtx.outerRows, ctx.outerRows)
		
		// Merge outer tables: current query's tables become outer tables for subquery
		for k, v := range ctx.tables {
			subCtx.outerTables[k] = v
		}
		// Also preserve any outer tables from parent contexts
		for k, v := range ctx.outerTables {
			subCtx.outerTables[k] = v
		}
		
		// Execute the subquery with context
		columns, rows, _, err := executePgSelectWithContext(selectStmt.SelectStmt, subCtx)
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
					colName := columns[i]
					
					// For derived tables, also store values with unqualified column names
					// This allows the outer query to reference columns without table prefixes
					if strings.Contains(colName, ".") {
						parts := strings.Split(colName, ".")
						simpleColName := parts[len(parts)-1]
						storageRow[simpleColName] = val
					}
					
					// Always store with the original column name too
					storageRow[colName] = val
				} else {
					storageRow[fmt.Sprintf("col%d", i)] = val
				}
			}
			result = append(result, storageRow)
		}
		return result, nil
	}
	log.Printf("WARNING: Unsupported subquery type. Returning empty result.\n")
	return []storage.Row{}, nil
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
		// Store current row for correlated subqueries
		oldRow := ctx.currentRow
		oldOuterRows := ctx.outerRows
		
		// Push current row onto outer rows stack
		ctx.outerRows = append([]storage.Row{row}, ctx.outerRows...)
		ctx.currentRow = row
		
		// Handle subquery in WHERE clause
		result := evaluateSubqueryExpression(row, e.SubLink, ctx)
		
		// Restore previous state
		ctx.currentRow = oldRow
		ctx.outerRows = oldOuterRows
		return result
	case *pg_query.Node_AExpr:
		return evaluateAExprWithContext(row, e.AExpr, ctx)
	case *pg_query.Node_BoolExpr:
		return evaluateBoolExprWithContext(row, e.BoolExpr, ctx)
	case *pg_query.Node_NullTest:
		return evaluateNullTestWithContext(row, e.NullTest, ctx)
	default:
		// Fall back to basic evaluation
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
		// Handle ALL comparison
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
		
		// Extract operator from operName if present
		operator := "="
		if sublink.OperName != nil && len(sublink.OperName) > 0 {
			if opNode, ok := sublink.OperName[0].Node.(*pg_query.Node_String_); ok {
				operator = opNode.String_.Sval
			}
		}
		
		// Check if any value in the list is NULL (only for equality operators)
		if operator == "=" || operator == "<>" || operator == "!=" {
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
			if hasNull && operator == "=" {
				return false
			}
		}
		
		// Check if testValue satisfies the operator with ALL rows in subquery result
		for _, subRow := range subRows {
			// Get the first column value from subquery result
			for _, val := range subRow {
				if val == nil {
					continue // Skip NULL values
				}
				// For ALL, the condition must be true for all values
				if !compareValuesPg(fmt.Sprintf("%v", testValue), operator, val) {
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
		
		// Extract operator from operName if present
		operator := "="
		if sublink.OperName != nil && len(sublink.OperName) > 0 {
			if opNode, ok := sublink.OperName[0].Node.(*pg_query.Node_String_); ok {
				operator = opNode.String_.Sval
			}
		}
		
		// Check if testValue matches any row in subquery result based on operator
		for _, subRow := range subRows {
			// Get the first column value from subquery result
			for _, val := range subRow {
				// Skip NULL values in the list - they don't match anything
				if val == nil {
					continue
				}
				if compareValuesPg(fmt.Sprintf("%v", testValue), operator, val) {
					return true
				}
				break // Only check first column
			}
		}
		return false
		
	case pg_query.SubLinkType_EXPR_SUBLINK:
		// Scalar subquery - should not be handled here
		// Scalar subqueries are handled directly in extractValueFromNodeWithContext
		return false
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
		fields := colRef.ColumnRef.Fields
		if len(fields) == 2 {
			// Qualified column reference (table.column)
			if str, ok := fields[1].Node.(*pg_query.Node_String_); ok {
				return row[str.String_.Sval]
			}
		} else if len(fields) == 1 {
			// Unqualified column reference
			if str, ok := fields[0].Node.(*pg_query.Node_String_); ok {
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
		// Keep track of group order for HAVING clause evaluation
		ctx.aggregations["__groupOrder__"] = [][]storage.Row{}
		
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
			
			// Store group rows in order
			if groupOrder, ok := ctx.aggregations["__groupOrder__"].([][]storage.Row); ok {
				ctx.aggregations["__groupOrder__"] = append(groupOrder, groupRows)
			}
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
			// Check if it's an aggregate function
			funcName := getFunctionName(val.FuncCall)
			if isAggregateFunction(funcName) {
				// Handle aggregate functions
				result := evaluateAggregateFunction(val.FuncCall, groupRows)
				if colName == "" {
					colName = strings.ToLower(funcName)
				}
				return result, colName
			} else {
				// Handle scalar functions
				result := evaluateScalarFunction(val.FuncCall, currentRow, ctx)
				if colName == "" {
					colName = strings.ToLower(funcName)
				}
				return result, colName
			}
		case *pg_query.Node_CoalesceExpr:
			// Handle COALESCE expression
			for _, arg := range val.CoalesceExpr.Args {
				result := evaluateExpression(arg, currentRow, ctx)
				if result != nil {
					return result, colName
				}
			}
			return nil, colName
		case *pg_query.Node_ColumnRef:
			// Handle column reference with potential table alias or schema qualification
			fields := val.ColumnRef.Fields
			
			if len(fields) > 0 {
				// Extract all parts
				var parts []string
				for _, field := range fields {
					if str, ok := field.Node.(*pg_query.Node_String_); ok {
						parts = append(parts, str.String_.Sval)
					}
				}
				
				// The last part is always the column name
				columnName := parts[len(parts)-1]
				
				// Try different qualified forms from most specific to least specific
				// e.g., schema.table.column -> table.column -> column
				for i := 0; i < len(parts); i++ {
					qualifiedName := strings.Join(parts[i:], ".")
					if val, exists := currentRow[qualifiedName]; exists {
						if colName == "" {
							// Preserve the full qualified name in output
							colName = strings.Join(parts, ".")
						}
						return val, colName
					}
				}
				
				// Final fallback: check if column exists in row
				if val, exists := currentRow[columnName]; exists {
					if colName == "" {
						colName = strings.Join(parts, ".")
					}
					return val, colName
				}
			}
		case *pg_query.Node_AConst:
			// Handle constant
			return extractAConstValue(val.AConst), colName
		case *pg_query.Node_AExpr:
			// Handle arithmetic/string expressions
			result := evaluateAExprValue(currentRow, val.AExpr, ctx)
			if colName == "" {
				// For expressions, use a descriptive name
				colName = "expr"
			}
			return result, colName
		case *pg_query.Node_SubLink:
			// Handle subquery in SELECT
			subRows, _ := executeSubquery(val.SubLink.Subselect, ctx)
			if len(subRows) > 0 && len(subRows[0]) > 0 {
				for _, v := range subRows[0] {
					return v, colName
				}
			}
			return nil, colName
		case *pg_query.Node_NullTest:
			// Handle IS NULL / IS NOT NULL
			result := evaluateNullTestWithContext(currentRow, val.NullTest, ctx)
			if colName == "" {
				colName = "?column?"
			}
			return result, colName
		default:
			// Try to extract value using general method
			result := extractValueFromNodeWithContext(currentRow, resTarget.Val, ctx)
			return result, colName
		}
	}

	return nil, colName
}

// isAggregateFunction is now in pg_parser_utils.go

func evaluateScalarFunction(funcCall *pg_query.FuncCall, row storage.Row, ctx *QueryContext) interface{} {
	if len(funcCall.Funcname) == 0 {
		return nil
	}
	
	funcName := getFunctionName(funcCall)
	
	switch funcName {
	case "COALESCE":
		// COALESCE returns the first non-NULL argument
		for _, arg := range funcCall.Args {
			val := evaluateExpression(arg, row, ctx)
			if val != nil {
				return val
			}
		}
		return nil
	case "UPPER":
		// UPPER converts string to uppercase
		if len(funcCall.Args) > 0 {
			val := evaluateExpression(funcCall.Args[0], row, ctx)
			if val != nil {
				if str, ok := val.(string); ok {
					return strings.ToUpper(str)
				}
			}
		}
		return nil
	case "LOWER":
		// LOWER converts string to lowercase
		if len(funcCall.Args) > 0 {
			val := evaluateExpression(funcCall.Args[0], row, ctx)
			if val != nil {
				if str, ok := val.(string); ok {
					return strings.ToLower(str)
				}
			}
		}
		return nil
	default:
		// Unknown function, return nil
		return nil
	}
}

func evaluateExpression(node *pg_query.Node, row storage.Row, ctx *QueryContext) interface{} {
	if node == nil {
		return nil
	}
	
	switch n := node.Node.(type) {
	case *pg_query.Node_ColumnRef:
		if len(n.ColumnRef.Fields) > 0 {
			// Handle both qualified and unqualified column references
			var colName string
			
			if len(n.ColumnRef.Fields) == 1 {
				// Unqualified column reference
				if str, ok := n.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
					colName = str.String_.Sval
				}
			} else if len(n.ColumnRef.Fields) == 2 {
				// Qualified column reference (table.column)
				if str, ok := n.ColumnRef.Fields[1].Node.(*pg_query.Node_String_); ok {
					colName = str.String_.Sval
				}
			}
			
			// First try the simple column name
			if val, exists := row[colName]; exists {
				return val
			}
			
			// If not found and it's a qualified reference, try the qualified name
			if len(n.ColumnRef.Fields) > 1 {
				var parts []string
				for _, field := range n.ColumnRef.Fields {
					if str, ok := field.Node.(*pg_query.Node_String_); ok {
						parts = append(parts, str.String_.Sval)
					}
				}
				qualifiedName := strings.Join(parts, ".")
				if val, exists := row[qualifiedName]; exists {
					return val
				}
			}
			
			return nil
		}
	case *pg_query.Node_AConst:
		return extractAConstValue(n.AConst)
	case *pg_query.Node_AExpr:
		// Handle arithmetic/string expressions
		return evaluateAExprValue(row, n.AExpr, ctx)
	case *pg_query.Node_FuncCall:
		funcName := getFunctionName(n.FuncCall)
		if isAggregateFunction(funcName) {
			// For aggregates in scalar context, we need the group rows
			// This shouldn't normally happen in a well-formed query
			return nil
		}
		return evaluateScalarFunction(n.FuncCall, row, ctx)
	case *pg_query.Node_NullTest:
		// Handle IS NULL / IS NOT NULL
		return evaluateNullTestWithContext(row, n.NullTest, ctx)
	}
	return nil
}

func evaluateAggregateFunction(funcCall *pg_query.FuncCall, rows []storage.Row) interface{} {
	if len(funcCall.Funcname) == 0 {
		return nil
	}

	funcName := getFunctionName(funcCall)
	
	// Extract column name or prepare to evaluate expression
	var colName string
	var isExpression bool
	var argExpr *pg_query.Node
	
	if len(funcCall.Args) > 0 {
		argExpr = funcCall.Args[0]
		
		if colRef, ok := funcCall.Args[0].Node.(*pg_query.Node_ColumnRef); ok {
			fields := colRef.ColumnRef.Fields
			if len(fields) == 2 {
				// Qualified column reference (table.column)
				if str, ok := fields[1].Node.(*pg_query.Node_String_); ok {
					colName = str.String_.Sval
				}
			} else if len(fields) == 1 {
				// Unqualified column reference
				if str, ok := fields[0].Node.(*pg_query.Node_String_); ok {
					colName = str.String_.Sval
				}
			}
		} else {
			// It's an expression, not just a column reference
			isExpression = true
		}
	}

	switch funcName {
	case "COUNT":
		if colName == "" || (len(funcCall.Args) > 0 && isStarExpr(funcCall.Args[0])) {
			return len(rows)
		}
		
		if funcCall.AggDistinct {
			// COUNT(DISTINCT column)
			seen := make(map[string]bool)
			count := 0
			for _, row := range rows {
				val := row[colName]
				if val != nil {
					key := fmt.Sprintf("%v", val)
					if !seen[key] {
						seen[key] = true
						count++
					}
				}
			}
			return count
		} else {
			// Regular COUNT(column or expression)
			count := 0
			
			if isExpression && argExpr != nil {
				// Create a temporary context for expression evaluation
				tmpCtx := &QueryContext{
					dataStore: nil,
					metaStore: nil,
					tables:    make(map[string]*TableContext),
				}
				
				for _, row := range rows {
					tmpCtx.currentRow = row
					val := evaluateExpression(argExpr, row, tmpCtx)
					if val != nil {
						count++
					}
				}
			} else {
				// Simple column reference
				for _, row := range rows {
					if row[colName] != nil {
						count++
					}
				}
			}
			return count
		}

	case "SUM":
		var sum float64
		hasNonNullValue := false
		
		if isExpression && argExpr != nil {
			// Create a temporary context for expression evaluation
			tmpCtx := &QueryContext{
				dataStore: nil,
				metaStore: nil,
				tables:    make(map[string]*TableContext),
			}
			
			// Evaluate the expression for each row
			for _, row := range rows {
				tmpCtx.currentRow = row
				val := evaluateExpression(argExpr, row, tmpCtx)
				if val != nil {
					if num, err := toFloat64(val); err == nil {
						sum += num
						hasNonNullValue = true
					}
				}
			}
		} else {
			// Simple column reference
			for _, row := range rows {
				if val := row[colName]; val != nil {
					if num, err := toFloat64(val); err == nil {
						sum += num
						hasNonNullValue = true
					}
				}
			}
		}
		
		// SQL standard: SUM returns NULL if no non-NULL values
		if !hasNonNullValue {
			return nil
		}
		return sum

	case "AVG":
		var sum float64
		count := 0
		
		if isExpression && argExpr != nil {
			// Create a temporary context for expression evaluation
			tmpCtx := &QueryContext{
				dataStore: nil,
				metaStore: nil,
				tables:    make(map[string]*TableContext),
			}
			
			for _, row := range rows {
				tmpCtx.currentRow = row
				val := evaluateExpression(argExpr, row, tmpCtx)
				if val != nil {
					if num, err := toFloat64(val); err == nil {
						sum += num
						count++
					}
				}
			}
		} else {
			// Simple column reference
			for _, row := range rows {
				if val := row[colName]; val != nil {
					if num, err := toFloat64(val); err == nil {
						sum += num
						count++
					}
				}
			}
		}
		
		if count == 0 {
			return nil
		}
		return sum / float64(count)

	case "MAX":
		var max interface{}
		
		if isExpression && argExpr != nil {
			// Create a temporary context for expression evaluation
			tmpCtx := &QueryContext{
				dataStore: nil,
				metaStore: nil,
				tables:    make(map[string]*TableContext),
			}
			
			for _, row := range rows {
				tmpCtx.currentRow = row
				val := evaluateExpression(argExpr, row, tmpCtx)
				if val != nil {
					if max == nil || compareValuesPg(fmt.Sprintf("%v", val), ">", max) {
						max = val
					}
				}
			}
		} else {
			// Simple column reference
			for _, row := range rows {
				if val := row[colName]; val != nil {
					if max == nil || compareValuesPg(fmt.Sprintf("%v", val), ">", max) {
						max = val
					}
				}
			}
		}
		return max

	case "MIN":
		var min interface{}
		
		if isExpression && argExpr != nil {
			// Create a temporary context for expression evaluation
			tmpCtx := &QueryContext{
				dataStore: nil,
				metaStore: nil,
				tables:    make(map[string]*TableContext),
			}
			
			for _, row := range rows {
				tmpCtx.currentRow = row
				val := evaluateExpression(argExpr, row, tmpCtx)
				if val != nil {
					if min == nil || compareValuesPg(fmt.Sprintf("%v", val), "<", min) {
						min = val
					}
				}
			}
		} else {
			// Simple column reference
			for _, row := range rows {
				if val := row[colName]; val != nil {
					if min == nil || compareValuesPg(fmt.Sprintf("%v", val), "<", min) {
						min = val
					}
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

// toFloat64 is now in pg_parser_utils.go

func extractColumnName(node *pg_query.Node) string {
	switch n := node.Node.(type) {
	case *pg_query.Node_ColumnRef:
		var parts []string
		for _, field := range n.ColumnRef.Fields {
			if str, ok := field.Node.(*pg_query.Node_String_); ok {
				parts = append(parts, str.String_.Sval)
			}
		}
		if len(parts) > 1 {
			// For qualified names (table.column), return the full qualified name
			return strings.Join(parts, ".")
		} else if len(parts) == 1 {
			// For unqualified names, return just the column name
			return parts[0]
		}
	case *pg_query.Node_FuncCall:
		funcName := getFunctionName(n.FuncCall)
		if funcName != "" {
			return strings.ToLower(funcName)
		}
	}
	return "?column?"
}

func filterResultRowsWithGroups(rows [][]interface{}, columns []string, havingClause *pg_query.Node, orderedGroups [][]storage.Row) [][]interface{} {
	var filtered [][]interface{}
	
	// If we don't have matching groups, fall back to simple evaluation
	if len(orderedGroups) == 0 {
		for _, row := range rows {
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
	
	for idx, row := range rows {
		// Convert row to map for evaluation
		rowMap := make(storage.Row)
		for i, col := range columns {
			if i < len(row) {
				rowMap[col] = row[i]
			}
		}
		
		// Get the corresponding group rows for this result row
		var groupRows []storage.Row
		if idx < len(orderedGroups) {
			groupRows = orderedGroups[idx]
		}
		
		// Also add entries for common aggregate function names
		// This helps when HAVING uses SUM(x) but the column is aliased
		for i, col := range columns {
			if i < len(row) && row[i] != nil {
				// If this looks like an aliased aggregate, also map the base function name
				// e.g., if column is "total_spent" and value is numeric, also map "sum" -> value
				// This is a heuristic but helps with HAVING SUM(x) > n when SELECT has SUM(x) AS total_spent
				switch v := row[i].(type) {
				case int, int64, float64:
					// For numeric columns, also map common aggregate function names
					if col != "sum" && col != "count" && col != "avg" && col != "max" && col != "min" {
						// This might be an aliased aggregate
						// For now, we'll map "sum" to any numeric alias
						// This is imperfect but handles the common case
						if !strings.Contains(col, ".") { // Not a qualified column
							rowMap["sum"] = v
						}
					}
				}
			}
		}
		
		// Evaluate HAVING clause - may need to compute aggregates on demand
		if evaluateHavingClause(rowMap, havingClause, groupRows) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func evaluateHavingClause(rowMap storage.Row, havingClause *pg_query.Node, groupRows []storage.Row) bool {
	// First try regular evaluation
	regularResult := evaluatePgWhere(rowMap, havingClause)
	
	if regularResult {
		return true
	}
	
	// If groupRows is empty, we can't compute aggregates
	if len(groupRows) == 0 {
		return false
	}
	
	// If that fails, check if we need to compute aggregates on demand
	switch expr := havingClause.Node.(type) {
	case *pg_query.Node_AExpr:
		// Check if expression contains aggregate functions
		if needsAggregateComputation(expr.AExpr) {
			return evaluateHavingWithAggregates(rowMap, expr.AExpr, groupRows)
		}
	case *pg_query.Node_BoolExpr:
		// Handle AND/OR expressions
		return evaluateHavingBoolExpr(rowMap, expr.BoolExpr, groupRows)
	}
	
	return false
}

func needsAggregateComputation(expr *pg_query.A_Expr) bool {
	// Check if left or right side contains function calls
	if expr.Lexpr != nil {
		if funcCall, ok := expr.Lexpr.Node.(*pg_query.Node_FuncCall); ok {
			funcName := getFunctionName(funcCall.FuncCall)
			if isAggregateFunction(funcName) {
				return true
			}
		}
	}
	if expr.Rexpr != nil {
		if funcCall, ok := expr.Rexpr.Node.(*pg_query.Node_FuncCall); ok {
			funcName := getFunctionName(funcCall.FuncCall)
			if isAggregateFunction(funcName) {
				return true
			}
		}
	}
	return false
}

func evaluateHavingWithAggregates(rowMap storage.Row, expr *pg_query.A_Expr, groupRows []storage.Row) bool {
	// Extract values, computing aggregates on demand if needed
	leftVal := extractHavingValue(rowMap, expr.Lexpr, groupRows)
	rightVal := extractHavingValue(rowMap, expr.Rexpr, groupRows)
	
	// Get operator
	op := ""
	if len(expr.Name) > 0 {
		if str, ok := expr.Name[0].Node.(*pg_query.Node_String_); ok {
			op = str.String_.Sval
		}
	}
	
	return compareValuesPg(fmt.Sprintf("%v", leftVal), op, rightVal)
}

func extractHavingValue(rowMap storage.Row, node *pg_query.Node, groupRows []storage.Row) interface{} {
	if node == nil {
		return nil
	}
	
	switch n := node.Node.(type) {
	case *pg_query.Node_FuncCall:
		funcName := getFunctionName(n.FuncCall)
		if isAggregateFunction(funcName) {
			// Compute aggregate on demand
			result := evaluateAggregateFunction(n.FuncCall, groupRows)
			return result
		}
		// Try to find in rowMap (aggregate columns are stored with lowercase names)
		lowerFuncName := strings.ToLower(funcName)
		if val, exists := rowMap[lowerFuncName]; exists {
			return val
		}
		// Also try the original case
		if val, exists := rowMap[funcName]; exists {
			return val
		}
	case *pg_query.Node_AConst:
		return extractAConstValue(n.AConst)
	case *pg_query.Node_ColumnRef:
		// Extract column value from rowMap
		if len(n.ColumnRef.Fields) > 0 {
			if str, ok := n.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
				return rowMap[str.String_.Sval]
			}
		}
	}
	return nil
}

func evaluateHavingBoolExpr(rowMap storage.Row, expr *pg_query.BoolExpr, groupRows []storage.Row) bool {
	switch expr.Boolop {
	case pg_query.BoolExprType_AND_EXPR:
		for _, arg := range expr.Args {
			if !evaluateHavingClause(rowMap, arg, groupRows) {
				return false
			}
		}
		return true
	case pg_query.BoolExprType_OR_EXPR:
		for _, arg := range expr.Args {
			if evaluateHavingClause(rowMap, arg, groupRows) {
				return true
			}
		}
		return false
	case pg_query.BoolExprType_NOT_EXPR:
		if len(expr.Args) > 0 {
			return !evaluateHavingClause(rowMap, expr.Args[0], groupRows)
		}
		return true
	}
	return true
}

func sortRows(rows [][]interface{}, columns []string, sortClause []*pg_query.Node) [][]interface{} {
	if len(sortClause) == 0 || len(rows) == 0 {
		return rows
	}
	
	// Create a copy of rows to avoid modifying the original
	result := make([][]interface{}, len(rows))
	copy(result, rows)
	
	// Sort using all sort clauses
	sort.Slice(result, func(i, j int) bool {
		// Compare using each sort clause in order
		for _, sortNode := range sortClause {
			if sortBy, ok := sortNode.Node.(*pg_query.Node_SortBy); ok {
				// Extract the column to sort by
				var colIndex int = -1
				var colName string
				
				if sortBy.SortBy.Node != nil {
					switch n := sortBy.SortBy.Node.Node.(type) {
					case *pg_query.Node_ColumnRef:
						// Get column name from the reference (handle qualified names)
						var parts []string
						for _, field := range n.ColumnRef.Fields {
							if str, ok := field.Node.(*pg_query.Node_String_); ok {
								parts = append(parts, str.String_.Sval)
							}
						}
						if len(parts) > 1 {
							// For qualified names (table.column), use the full qualified name
							colName = strings.Join(parts, ".")
						} else if len(parts) == 1 {
							// For unqualified names, use just the column name
							colName = parts[0]
						}
					case *pg_query.Node_AConst:
						// Handle ORDER BY position (e.g., ORDER BY 1)
						if val, ok := n.AConst.Val.(*pg_query.A_Const_Ival); ok {
							colIndex = int(val.Ival.Ival) - 1 // PostgreSQL uses 1-based indexing
						}
					}
				}
				
				// Find column index if we have a column name
				if colName != "" {
					for idx, col := range columns {
						if col == colName {
							colIndex = idx
							break
						}
					}
				}
				
				// Skip if column not found
				if colIndex < 0 || colIndex >= len(columns) {
					continue
				}
				
				// Get values to compare
				var val1, val2 interface{}
				if colIndex < len(result[i]) {
					val1 = result[i][colIndex]
				}
				if colIndex < len(result[j]) {
					val2 = result[j][colIndex]
				}
				
				// Handle NULLs according to NULLS FIRST/LAST
				nullsFirst := sortBy.SortBy.SortbyNulls == pg_query.SortByNulls_SORTBY_NULLS_FIRST
				if sortBy.SortBy.SortbyNulls == pg_query.SortByNulls_SORTBY_NULLS_DEFAULT {
					// Default: NULLS LAST for ASC, NULLS FIRST for DESC
					nullsFirst = sortBy.SortBy.SortbyDir == pg_query.SortByDir_SORTBY_DESC
				}
				
				if val1 == nil && val2 == nil {
					continue // Both NULL, check next sort clause
				}
				if val1 == nil {
					return nullsFirst // NULL vs non-NULL
				}
				if val2 == nil {
					return !nullsFirst // non-NULL vs NULL
				}
				
				// Compare non-NULL values
				cmp := compareForSort(val1, val2)
				if cmp == 0 {
					continue // Equal, check next sort clause
				}
				
				// Apply sort direction
				if sortBy.SortBy.SortbyDir == pg_query.SortByDir_SORTBY_DESC {
					return cmp > 0
				}
				return cmp < 0
			}
		}
		return false // All sort clauses resulted in equality
	})
	
	return result
}

// compareForSort compares two values for sorting purposes
// Returns -1 if val1 < val2, 0 if equal, 1 if val1 > val2
func compareForSort(val1, val2 interface{}) int {
	// Try to compare as numbers first
	num1, err1 := toFloat64(val1)
	num2, err2 := toFloat64(val2)
	
	if err1 == nil && err2 == nil {
		// Both are numbers
		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
		return 0
	}
	
	// Compare as strings
	str1 := fmt.Sprintf("%v", val1)
	str2 := fmt.Sprintf("%v", val2)
	
	if str1 < str2 {
		return -1
	} else if str1 > str2 {
		return 1
	}
	return 0
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

func evaluateAExprWithContext(row storage.Row, expr *pg_query.A_Expr, ctx *QueryContext) bool {
	var leftVal, rightVal interface{}

	// Extract values using context-aware function
	if expr.Lexpr != nil {
		leftVal = extractValueFromNodeWithContext(row, expr.Lexpr, ctx)
	}
	if expr.Rexpr != nil {
		rightVal = extractValueFromNodeWithContext(row, expr.Rexpr, ctx)
	}

	// Get operator
	op := ""
	if len(expr.Name) > 0 {
		if str, ok := expr.Name[0].Node.(*pg_query.Node_String_); ok {
			op = str.String_.Sval
		}
	}

	// Handle different expression kinds
	switch expr.Kind {
	case pg_query.A_Expr_Kind_AEXPR_OP:
		return compareValuesPg(fmt.Sprintf("%v", leftVal), op, rightVal)
	case pg_query.A_Expr_Kind_AEXPR_IN:
		// IN expression is handled by evaluatePgWhere
		return evaluatePgWhere(row, &pg_query.Node{Node: &pg_query.Node_AExpr{AExpr: expr}})
	case pg_query.A_Expr_Kind_AEXPR_LIKE:
		// LIKE expression
		return compareValuesPg(fmt.Sprintf("%v", leftVal), "~~", rightVal)
	case pg_query.A_Expr_Kind_AEXPR_ILIKE:
		// ILIKE expression (case-insensitive)
		return compareValuesPg(fmt.Sprintf("%v", leftVal), "~~*", rightVal)
	default:
		// For other expression kinds, try to handle them
		return evaluatePgWhere(row, &pg_query.Node{Node: &pg_query.Node_AExpr{AExpr: expr}})
	}
}

func evaluateBoolExprWithContext(row storage.Row, expr *pg_query.BoolExpr, ctx *QueryContext) bool {
	switch expr.Boolop {
	case pg_query.BoolExprType_AND_EXPR:
		for _, arg := range expr.Args {
			if !evaluateWhereWithSubqueries(row, arg, ctx) {
				return false
			}
		}
		return true
	case pg_query.BoolExprType_OR_EXPR:
		for _, arg := range expr.Args {
			if evaluateWhereWithSubqueries(row, arg, ctx) {
				return true
			}
		}
		return false
	case pg_query.BoolExprType_NOT_EXPR:
		if len(expr.Args) > 0 {
			return !evaluateWhereWithSubqueries(row, expr.Args[0], ctx)
		}
		return true
	}
	return true
}

func evaluateNullTestWithContext(row storage.Row, expr *pg_query.NullTest, ctx *QueryContext) bool {
	val := extractValueFromNodeWithContext(row, expr.Arg, ctx)
	
	switch expr.Nulltesttype {
	case pg_query.NullTestType_IS_NULL:
		return val == nil
	case pg_query.NullTestType_IS_NOT_NULL:
		return val != nil
	default:
		return false
	}
}

func evaluateAExprValue(row storage.Row, expr *pg_query.A_Expr, ctx *QueryContext) interface{} {
	// Extract left and right values
	var leftVal, rightVal interface{}
	
	if expr.Lexpr != nil {
		leftVal = extractValueFromNodeWithContext(row, expr.Lexpr, ctx)
	}
	if expr.Rexpr != nil {
		rightVal = extractValueFromNodeWithContext(row, expr.Rexpr, ctx)
	}
	
	// Get operator
	op := ""
	if len(expr.Name) > 0 {
		if str, ok := expr.Name[0].Node.(*pg_query.Node_String_); ok {
			op = str.String_.Sval
		}
	}
	
	
	// Perform arithmetic operation
	switch op {
	case "+":
		left, _ := toFloat64(leftVal)
		right, _ := toFloat64(rightVal)
		return left + right
	case "-":
		left, _ := toFloat64(leftVal)
		right, _ := toFloat64(rightVal)
		return left - right
	case "*":
		left, _ := toFloat64(leftVal)
		right, _ := toFloat64(rightVal)
		return left * right
	case "/":
		left, _ := toFloat64(leftVal)
		right, _ := toFloat64(rightVal)
		if right != 0 {
			return left / right
		}
		return nil
	case "||":
		// String concatenation - if either operand is NULL, result is NULL
		if leftVal == nil || rightVal == nil {
			return nil
		}
		return fmt.Sprintf("%v%v", leftVal, rightVal)
	default:
		// For comparison operators, return boolean result
		return compareValuesPg(leftVal, op, rightVal)
	}
}

func extractValueFromNodeWithContext(row storage.Row, node *pg_query.Node, ctx *QueryContext) interface{} {
	switch n := node.Node.(type) {
	case *pg_query.Node_ColumnRef:
		// Handle qualified column references (table.column)
		if len(n.ColumnRef.Fields) >= 2 {
			// table.column format
			tableName := ""
			columnName := ""
			
			if str, ok := n.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
				tableName = str.String_.Sval
			}
			if str, ok := n.ColumnRef.Fields[1].Node.(*pg_query.Node_String_); ok {
				columnName = str.String_.Sval
			}
			
			// First try the qualified column name (table.column)
			qualifiedName := tableName + "." + columnName
			if val, exists := row[qualifiedName]; exists {
				return val
			}
			
			// If not found in current row, check if this is an outer table
			if _, isOuterTable := ctx.outerTables[tableName]; isOuterTable {
				// This table is from an outer query
				// First check the most recent outer row in the stack
				if len(ctx.outerRows) > 0 {
					for _, outerRow := range ctx.outerRows {
						if val, exists := outerRow[columnName]; exists {
							return val
						}
						if val, exists := outerRow[qualifiedName]; exists {
							return val
						}
					}
				}
				// Also check currentRow if available
				if ctx.currentRow != nil {
					if val, exists := ctx.currentRow[columnName]; exists {
						return val
					}
					if val, exists := ctx.currentRow[qualifiedName]; exists {
						return val
					}
				}
			}
			
			// If not found in current row, check outer rows stack
			// Start from the most recent outer row (index 0)
			for _, outerRow := range ctx.outerRows {
				// Try qualified name first
				if val, exists := outerRow[qualifiedName]; exists {
					return val
				}
				// Then try unqualified name
				if val, exists := outerRow[columnName]; exists {
					// TODO: Should verify this belongs to the correct table
					return val
				}
			}
			
			// Finally check currentRow if available
			if ctx.currentRow != nil {
				if val, exists := ctx.currentRow[qualifiedName]; exists {
					return val
				}
				if val, exists := ctx.currentRow[columnName]; exists {
					return val
				}
			}
		} else if len(n.ColumnRef.Fields) == 1 {
			// Unqualified column - first check current row, then outer row
			if str, ok := n.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
				columnName := str.String_.Sval
				// First check current subquery row
				if val, exists := row[columnName]; exists {
					return val
				}
				// Then check outer query row if available
				if ctx.currentRow != nil {
					if val, exists := ctx.currentRow[columnName]; exists {
						return val
					}
				}
			}
		}
	case *pg_query.Node_AConst:
		return extractAConstValue(n.AConst)
	case *pg_query.Node_AExpr:
		// Handle arithmetic expressions
		return evaluateAExprValue(row, n.AExpr, ctx)
	case *pg_query.Node_SubLink:
		// Handle scalar subquery
		// Store current row for correlated subqueries
		oldRow := ctx.currentRow
		oldOuterRows := ctx.outerRows
		
		// Push current row onto outer rows stack
		if row != nil {
			ctx.outerRows = append([]storage.Row{row}, ctx.outerRows...)
			ctx.currentRow = row
		}
		
		subRows, err := executeSubquery(n.SubLink.Subselect, ctx)
		
		// Restore previous state
		ctx.currentRow = oldRow
		ctx.outerRows = oldOuterRows
		
		if err != nil || len(subRows) != 1 || len(subRows[0]) == 0 {
			return nil
		}
		// Return the first column of the first row
		for _, val := range subRows[0] {
			return val
		}
	case *pg_query.Node_NullTest:
		// Handle IS NULL / IS NOT NULL
		return evaluateNullTestWithContext(row, n.NullTest, ctx)
	}
	return nil
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
		log.Printf("WARNING: Unsupported set operation type. Returning empty result.\n")
		return []string{}, [][]interface{}{}, "SELECT 0", nil
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
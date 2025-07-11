package parser

import (
	"fmt"
	"sort"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
	"vsql/storage"
)

func ParsePostgreSQL(query string) (*pg_query.ParseResult, error) {
	return pg_query.Parse(query)
}

func ExecutePgQuery(query string, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	result, err := ParsePostgreSQL(query)
	if err != nil {
		return nil, nil, "", err
	}

	if len(result.Stmts) == 0 {
		return nil, nil, "", fmt.Errorf("no statements found")
	}

	stmt := result.Stmts[0].Stmt
	switch node := stmt.Node.(type) {
	case *pg_query.Node_SelectStmt:
		return executePgSelect(node.SelectStmt, dataStore, metaStore)
	case *pg_query.Node_InsertStmt:
		return executePgInsert(node.InsertStmt, dataStore, metaStore)
	case *pg_query.Node_UpdateStmt:
		return executePgUpdate(node.UpdateStmt, dataStore, metaStore)
	case *pg_query.Node_DeleteStmt:
		return executePgDelete(node.DeleteStmt, dataStore)
	case *pg_query.Node_CreateStmt:
		return executePgCreateTable(node.CreateStmt, dataStore, metaStore)
	case *pg_query.Node_DropStmt:
		return executePgDropTable(node.DropStmt, dataStore, metaStore)
	default:
		return nil, nil, "", fmt.Errorf("unsupported statement type: %T", node)
	}
}

func executePgSelect(stmt *pg_query.SelectStmt, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	// Check if this is a complex query that needs advanced processing
	if needsAdvancedProcessing(stmt) {
		return executePgSelectAdvanced(stmt, dataStore, metaStore)
	}

	// Simple single-table query - use optimized path
	if len(stmt.FromClause) != 1 {
		return nil, nil, "", fmt.Errorf("only single table SELECT is supported in simple mode")
	}

	tableName := extractTableName(stmt.FromClause[0])
	if tableName == "" {
		return nil, nil, "", fmt.Errorf("could not extract table name")
	}

	table, exists := dataStore.GetTable(tableName)
	if !exists {
		return nil, nil, "", fmt.Errorf("table '%s' does not exist", tableName)
	}

	rows := table.GetRows()
	columns := extractSelectColumns(stmt, tableName, metaStore, rows)

	var resultRows [][]interface{}
	for _, row := range rows {
		if stmt.WhereClause != nil && !evaluatePgWhere(row, stmt.WhereClause) {
			continue
		}

		resultRow := make([]interface{}, len(columns))
		for i, col := range columns {
			resultRow[i] = row[col]
		}
		resultRows = append(resultRows, resultRow)
	}

	return columns, resultRows, fmt.Sprintf("SELECT %d", len(resultRows)), nil
}

func needsAdvancedProcessing(stmt *pg_query.SelectStmt) bool {
	// Check if query has features requiring advanced processing
	if len(stmt.FromClause) > 1 {
		return true
	}
	
	// Check for UNION/INTERSECT/EXCEPT (SetOp)
	if stmt.Op != pg_query.SetOperation_SETOP_NONE {
		return true
	}
	
	// Check for JOINs
	if len(stmt.FromClause) > 0 {
		if _, ok := stmt.FromClause[0].Node.(*pg_query.Node_JoinExpr); ok {
			return true
		}
		// Check for subquery in FROM
		if _, ok := stmt.FromClause[0].Node.(*pg_query.Node_RangeSubselect); ok {
			return true
		}
	}
	
	// Check for GROUP BY
	if len(stmt.GroupClause) > 0 {
		return true
	}
	
	// Check for HAVING
	if stmt.HavingClause != nil {
		return true
	}
	
	// Check for aggregate functions in target list
	for _, target := range stmt.TargetList {
		if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok {
			if funcCall, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_FuncCall); ok && funcCall.FuncCall != nil {
				return true
			}
			// Check for subquery in SELECT
			if _, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_SubLink); ok {
				return true
			}
		}
	}
	
	// Check for subquery in WHERE
	if stmt.WhereClause != nil {
		if hasSubquery(stmt.WhereClause) {
			return true
		}
	}
	
	// Check for ORDER BY
	if len(stmt.SortClause) > 0 {
		return true
	}
	
	// Check for LIMIT/OFFSET
	if stmt.LimitCount != nil || stmt.LimitOffset != nil {
		return true
	}
	
	// Check for DISTINCT
	if stmt.DistinctClause != nil && len(stmt.DistinctClause) > 0 {
		return true
	}
	
	return false
}

func hasSubquery(node *pg_query.Node) bool {
	switch n := node.Node.(type) {
	case *pg_query.Node_SubLink:
		return true
	case *pg_query.Node_BoolExpr:
		for _, arg := range n.BoolExpr.Args {
			if hasSubquery(arg) {
				return true
			}
		}
	case *pg_query.Node_AExpr:
		if hasSubquery(n.AExpr.Lexpr) || hasSubquery(n.AExpr.Rexpr) {
			return true
		}
	}
	return false
}

func executePgInsert(stmt *pg_query.InsertStmt, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	tableName := extractTableNameFromRangeVar(stmt.Relation)
	if tableName == "" {
		return nil, nil, "", fmt.Errorf("could not extract table name")
	}

	if err := dataStore.CreateTable(tableName); err != nil {
		return nil, nil, "", err
	}

	table, _ := dataStore.GetTable(tableName)

	var columns []string
	if stmt.Cols != nil {
		for _, col := range stmt.Cols {
			if target, ok := col.Node.(*pg_query.Node_ResTarget); ok {
				columns = append(columns, target.ResTarget.Name)
			}
		}
	} else {
		// No columns specified in INSERT, try to get them from metastore
		columns = metaStore.GetTableColumns(tableName)
	}

	rowsInserted := 0
	if selectStmt, ok := stmt.SelectStmt.Node.(*pg_query.Node_SelectStmt); ok {
		if len(selectStmt.SelectStmt.ValuesLists) > 0 {
			for _, valuesList := range selectStmt.SelectStmt.ValuesLists {
				row := make(storage.Row)
				if list, ok := valuesList.Node.(*pg_query.Node_List); ok {
					values := list.List.Items
					for i, val := range values {
						if i < len(columns) {
							row[columns[i]] = extractPgValue(val)
						}
					}
					table.Insert(row)
					metaStore.UpdateFromRow(tableName, row)
					rowsInserted++
				}
			}
		}
	}

	return nil, nil, fmt.Sprintf("INSERT 0 %d", rowsInserted), nil
}

func executePgUpdate(stmt *pg_query.UpdateStmt, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	tableName := extractTableNameFromRangeVar(stmt.Relation)
	if tableName == "" {
		return nil, nil, "", fmt.Errorf("could not extract table name")
	}

	table, exists := dataStore.GetTable(tableName)
	if !exists {
		return nil, nil, "", fmt.Errorf("table '%s' does not exist", tableName)
	}

	rows := table.GetRows()
	updatedCount := 0

	for _, row := range rows {
		if stmt.WhereClause != nil && !evaluatePgWhere(row, stmt.WhereClause) {
			continue
		}

		for _, target := range stmt.TargetList {
			if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok {
				colName := resTarget.ResTarget.Name
				value := extractPgValue(resTarget.ResTarget.Val)
				row[colName] = value
				metaStore.AddColumn(tableName, colName)
			}
		}
		updatedCount++
	}

	return nil, nil, fmt.Sprintf("UPDATE %d", updatedCount), nil
}

func executePgDelete(stmt *pg_query.DeleteStmt, dataStore *storage.DataStore) ([]string, [][]interface{}, string, error) {
	tableName := extractTableNameFromRangeVar(stmt.Relation)
	if tableName == "" {
		return nil, nil, "", fmt.Errorf("could not extract table name")
	}

	table, exists := dataStore.GetTable(tableName)
	if !exists {
		return nil, nil, "", fmt.Errorf("table '%s' does not exist", tableName)
	}

	rows := table.GetRows()
	var newRows []storage.Row
	deletedCount := 0

	for _, row := range rows {
		if stmt.WhereClause != nil && evaluatePgWhere(row, stmt.WhereClause) {
			deletedCount++
		} else {
			newRows = append(newRows, row)
		}
	}

	table.Rows = newRows

	return nil, nil, fmt.Sprintf("DELETE %d", deletedCount), nil
}

func executePgCreateTable(stmt *pg_query.CreateStmt, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	tableName := extractTableNameFromRangeVar(stmt.Relation)
	if tableName == "" {
		return nil, nil, "", fmt.Errorf("could not extract table name")
	}

	if err := dataStore.CreateTable(tableName); err != nil {
		return nil, nil, "", err
	}

	// Extract column names from table elements
	var columns []string
	for _, elem := range stmt.TableElts {
		if colDef, ok := elem.Node.(*pg_query.Node_ColumnDef); ok {
			columns = append(columns, colDef.ColumnDef.Colname)
		}
	}

	// Store column names in metastore if any were defined
	if len(columns) > 0 {
		metaStore.AddColumns(tableName, columns)
	}

	return nil, nil, "CREATE TABLE", nil
}

func executePgDropTable(stmt *pg_query.DropStmt, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	for _, obj := range stmt.Objects {
		if list, ok := obj.Node.(*pg_query.Node_List); ok && len(list.List.Items) > 0 {
			if str, ok := list.List.Items[0].Node.(*pg_query.Node_String_); ok {
				tableName := str.String_.Sval
				dataStore.DropTable(tableName)
				metaStore.DropTable(tableName)
			}
		}
	}

	return nil, nil, "DROP TABLE", nil
}

func extractTableName(node *pg_query.Node) string {
	switch n := node.Node.(type) {
	case *pg_query.Node_RangeVar:
		return n.RangeVar.Relname
	}
	return ""
}

func extractTableNameFromRangeVar(rv *pg_query.RangeVar) string {
	if rv != nil {
		return rv.Relname
	}
	return ""
}

func extractSelectColumns(stmt *pg_query.SelectStmt, tableName string, metaStore *storage.MetaStore, rows []storage.Row) []string {
	var columns []string

	for _, target := range stmt.TargetList {
		if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok {
			if resTarget.ResTarget.Name != "" {
				columns = append(columns, resTarget.ResTarget.Name)
			} else if colRef, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_ColumnRef); ok {
				if len(colRef.ColumnRef.Fields) > 0 {
					if str, ok := colRef.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
						columns = append(columns, str.String_.Sval)
					} else if _, ok := colRef.ColumnRef.Fields[0].Node.(*pg_query.Node_AStar); ok {
						columns = metaStore.GetTableColumns(tableName)
						if len(columns) == 0 && len(rows) > 0 {
							// Collect column names and sort them for consistent ordering
							var allCols []string
							for key := range rows[0] {
								allCols = append(allCols, key)
							}
							sort.Strings(allCols)
							columns = allCols
						}
						break
					}
				}
			}
		}
	}

	return columns
}

func evaluatePgWhere(row storage.Row, whereClause *pg_query.Node) bool {
	switch expr := whereClause.Node.(type) {
	case *pg_query.Node_AExpr:
		return evaluateAExpr(row, expr.AExpr)
	case *pg_query.Node_BoolExpr:
		return evaluateBoolExpr(row, expr.BoolExpr)
	case *pg_query.Node_NullTest:
		return evaluateNullTest(row, expr.NullTest)
	case *pg_query.Node_SubLink:
		// Handle subqueries - simplified version
		return true // Will be handled by advanced query processor
	default:
		return true
	}
}

func evaluateNullTest(row storage.Row, expr *pg_query.NullTest) bool {
	var val interface{}

	// Get the column value
	if colRef, ok := expr.Arg.Node.(*pg_query.Node_ColumnRef); ok {
		if len(colRef.ColumnRef.Fields) > 0 {
			if str, ok := colRef.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
				val = row[str.String_.Sval]
			}
		}
	}

	// Check null test type
	switch expr.Nulltesttype {
	case pg_query.NullTestType_IS_NULL:
		return val == nil
	case pg_query.NullTestType_IS_NOT_NULL:
		return val != nil
	default:
		return false
	}
}

func evaluateAExpr(row storage.Row, expr *pg_query.A_Expr) bool {
	var leftVal, rightVal interface{}

	// Check if this is an IN expression with a value list
	if expr.Kind == pg_query.A_Expr_Kind_AEXPR_IN {
		// Extract left value
		if expr.Lexpr != nil {
			leftVal = extractValueFromExpr(row, expr.Lexpr)
		}
		
		// Check if right side is a list
		if expr.Rexpr != nil {
			if listNode, ok := expr.Rexpr.Node.(*pg_query.Node_List); ok {
				// Determine if this is IN or NOT IN by checking the operator
				isNotIn := false
				if len(expr.Name) > 0 {
					if str, ok := expr.Name[0].Node.(*pg_query.Node_String_); ok {
						isNotIn = (str.String_.Sval == "<>")
					}
				}
				
				if isNotIn {
					// NOT IN logic
					// If left value is NULL, NOT IN always returns false
					if leftVal == nil {
						return false
					}
					
					// Check if any value in the list is NULL
					hasNull := false
					for _, item := range listNode.List.Items {
						itemVal := extractValueFromExpr(row, item)
						if itemVal == nil {
							hasNull = true
							break
						}
					}
					
					// If list contains NULL, NOT IN returns false for all non-NULL values
					if hasNull {
						return false
					}
					
					// Check if left value matches any value in the list
					for _, item := range listNode.List.Items {
						itemVal := extractValueFromExpr(row, item)
						if compareValuesPg(leftVal, "=", itemVal) {
							// Found a match, so NOT IN returns false
							return false
						}
					}
					// No match found, NOT IN returns true
					return true
				} else {
					// Regular IN logic
					// If left value is NULL, IN always returns false
					if leftVal == nil {
						return false
					}
					
					// Check if left value matches any value in the list
					for _, item := range listNode.List.Items {
						itemVal := extractValueFromExpr(row, item)
						// Skip NULL values in the list
						if itemVal == nil {
							continue
						}
						// Use the same comparison logic as regular equals
						if compareValuesPg(leftVal, "=", itemVal) {
							return true
						}
					}
					// No match found
					return false
				}
			}
		}
	}

	// Check if this is an IN or EXISTS expression
	opName := ""
	if len(expr.Name) > 0 {
		if str, ok := expr.Name[0].Node.(*pg_query.Node_String_); ok {
			opName = str.String_.Sval
		}
	}

	// Handle BETWEEN expression
	if expr.Kind == pg_query.A_Expr_Kind_AEXPR_BETWEEN || 
	   expr.Kind == pg_query.A_Expr_Kind_AEXPR_NOT_BETWEEN {
		// Extract the value to test
		if expr.Lexpr != nil {
			leftVal = extractValueFromExpr(row, expr.Lexpr)
		}
		
		// BETWEEN requires a list with exactly 2 elements (lower and upper bounds)
		if expr.Rexpr != nil {
			if listNode, ok := expr.Rexpr.Node.(*pg_query.Node_List); ok && len(listNode.List.Items) == 2 {
				lowerBound := extractValueFromExpr(row, listNode.List.Items[0])
				upperBound := extractValueFromExpr(row, listNode.List.Items[1])
				
				// BETWEEN is inclusive: value >= lower AND value <= upper
				result := compareValuesPg(leftVal, ">=", lowerBound) && compareValuesPg(leftVal, "<=", upperBound)
				
				// If NOT BETWEEN, negate the result
				if expr.Kind == pg_query.A_Expr_Kind_AEXPR_NOT_BETWEEN {
					result = !result
				}
				
				return result
			}
		}
		// If we can't extract proper bounds, return false
		return false
	}

	// Handle IN expression with subquery
	if opName == "=" && expr.Rexpr != nil {
		if _, ok := expr.Rexpr.Node.(*pg_query.Node_SubLink); ok {
			// This is an IN subquery - should be handled by advanced processor
			return true
		}
	}

	// Extract left value
	if expr.Lexpr != nil {
		leftVal = extractValueFromExpr(row, expr.Lexpr)
	}

	// Extract right value
	if expr.Rexpr != nil {
		rightVal = extractValueFromExpr(row, expr.Rexpr)
	}

	return compareValuesPg(leftVal, opName, rightVal)
}

func evaluateBoolExpr(row storage.Row, expr *pg_query.BoolExpr) bool {
	switch expr.Boolop {
	case pg_query.BoolExprType_AND_EXPR:
		for _, arg := range expr.Args {
			if !evaluatePgWhere(row, arg) {
				return false
			}
		}
		return true
	case pg_query.BoolExprType_OR_EXPR:
		for _, arg := range expr.Args {
			if evaluatePgWhere(row, arg) {
				return true
			}
		}
		return false
	case pg_query.BoolExprType_NOT_EXPR:
		if len(expr.Args) > 0 {
			return !evaluatePgWhere(row, expr.Args[0])
		}
		return true
	}
	return true
}

func extractValueFromExpr(row storage.Row, node *pg_query.Node) interface{} {
	if node == nil {
		return nil
	}
	
	switch n := node.Node.(type) {
	case *pg_query.Node_ColumnRef:
		// Handle qualified (table.column) and unqualified (column) references
		if len(n.ColumnRef.Fields) >= 2 {
			// Qualified column reference (table.column)
			tableName := ""
			columnName := ""
			
			if str, ok := n.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
				tableName = str.String_.Sval
			}
			if str, ok := n.ColumnRef.Fields[1].Node.(*pg_query.Node_String_); ok {
				columnName = str.String_.Sval
			}
			
			// First try the qualified column name
			qualifiedName := tableName + "." + columnName
			if val, exists := row[qualifiedName]; exists {
				return val
			}
			
			// Fall back to unqualified column name
			val := row[columnName]
			return val
		} else if len(n.ColumnRef.Fields) > 0 {
			// Unqualified column reference
			if str, ok := n.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
				colName := str.String_.Sval
				return row[colName]
			}
		}
	case *pg_query.Node_AConst:
		return extractAConstValue(n.AConst)
	case *pg_query.Node_FuncCall:
		// Handle function calls in HAVING context
		// When we're evaluating HAVING, aggregate results are already computed and stored in the row
		funcName := getFunctionName(n.FuncCall)
		
		if funcName != "" {
			// Try lowercase version (how aggregates are stored in result rows)
			lowerFuncName := strings.ToLower(funcName)
			if val, exists := row[lowerFuncName]; exists {
				return val
			}
			// Also try uppercase version
			if val, exists := row[funcName]; exists {
				return val
			}
			
			// Special handling for COUNT(*) - it's stored as "count" in result rows
			if funcName == "COUNT" && len(n.FuncCall.Args) > 0 {
				// Check if it's COUNT(*)
				if colRef, ok := n.FuncCall.Args[0].Node.(*pg_query.Node_ColumnRef); ok {
					if len(colRef.ColumnRef.Fields) > 0 {
						if _, isStar := colRef.ColumnRef.Fields[0].Node.(*pg_query.Node_AStar); isStar {
							// This is COUNT(*), look for "count" column
							if val, exists := row["count"]; exists {
								return val
							}
						}
					}
				}
			}
		}
	case *pg_query.Node_AExpr:
		// Handle arithmetic expressions
		return evaluateArithmeticExpr(row, n.AExpr)
	}
	
	return nil
}

func evaluateArithmeticExpr(row storage.Row, expr *pg_query.A_Expr) interface{} {
	// Only handle arithmetic operations
	if expr.Kind != pg_query.A_Expr_Kind_AEXPR_OP {
		return nil
	}
	
	// Extract operator
	var op string
	if len(expr.Name) > 0 {
		if str, ok := expr.Name[0].Node.(*pg_query.Node_String_); ok {
			op = str.String_.Sval
		}
	}
	
	// Extract left and right values
	leftVal := extractValueFromExpr(row, expr.Lexpr)
	rightVal := extractValueFromExpr(row, expr.Rexpr)
	
	// Convert to float64 for arithmetic
	leftNum, err1 := toFloat64(leftVal)
	rightNum, err2 := toFloat64(rightVal)
	
	if err1 != nil || err2 != nil {
		return nil
	}
	
	// Perform arithmetic operation
	switch op {
	case "+":
		return leftNum + rightNum
	case "-":
		return leftNum - rightNum
	case "*":
		return leftNum * rightNum
	case "/":
		if rightNum == 0 {
			return nil // Division by zero
		}
		return leftNum / rightNum
	default:
		return nil
	}
}

func extractPgValue(node *pg_query.Node) interface{} {
	if node == nil {
		return nil
	}

	switch val := node.Node.(type) {
	case *pg_query.Node_AConst:
		return extractAConstValue(val.AConst)
	case *pg_query.Node_String_:
		return val.String_.Sval
	case *pg_query.Node_Integer:
		return val.Integer.Ival
	}
	return nil
}

// extractAConstValue is now in pg_parser_utils.go

// compareValuesPg is now in pg_parser_utils.go

// toNumber is now in pg_parser_utils.go

// matchPattern is now in pg_parser_utils.go
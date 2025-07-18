package parser

import (
	"fmt"
	"log"
	"sort"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
	"github.com/satetsu888/vsql/storage"
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
	case *pg_query.Node_PrepareStmt:
		return executePgPrepare(node.PrepareStmt, dataStore, metaStore)
	case *pg_query.Node_ExecuteStmt:
		return executePgExecute(node.ExecuteStmt, dataStore, metaStore)
	case *pg_query.Node_DeallocateStmt:
		return executePgDeallocate(node.DeallocateStmt)
	default:
		// Log warning for unsupported statement types but return empty result
		log.Printf("WARNING: Unsupported SQL statement type: %T. Query will be ignored.\n", node)
		return []string{}, [][]interface{}{}, "SELECT 0", nil
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

	var rows []storage.Row
	table, exists := dataStore.GetTable(tableName)
	if exists {
		rows = table.GetRows()
	} else {
		// Table doesn't exist - return empty result set
		rows = []storage.Row{}
	}
	
	columns := extractSelectColumns(stmt, tableName, metaStore, rows)

	var resultRows [][]interface{}
	for _, row := range rows {
		if stmt.WhereClause != nil && !evaluatePgWhere(row, stmt.WhereClause) {
			continue
		}

		resultRow := make([]interface{}, len(columns))
		for i, col := range columns {
			// Try the column name as-is first
			if val, exists := row[col]; exists {
				resultRow[i] = val
			} else if strings.Contains(col, ".") {
				// If it's a qualified name and not found, try the unqualified part
				parts := strings.Split(col, ".")
				unqualified := parts[len(parts)-1]
				resultRow[i] = row[unqualified]
			} else {
				resultRow[i] = row[col]
			}
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
		if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok && resTarget.ResTarget.Val != nil {
			if funcCall, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_FuncCall); ok && funcCall.FuncCall != nil {
				return true
			}
			// Check for subquery in SELECT
			if _, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_SubLink); ok {
				return true
			}
			// Check for constants (string literals, numbers)
			if _, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_AConst); ok {
				return true
			}
			// Check for expressions (arithmetic, string concatenation)
			if _, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_AExpr); ok {
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
				// Remove quotes if present for consistency
				colName := strings.Trim(target.ResTarget.Name, `"`)
				columns = append(columns, colName)
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
							extractedValue := extractPgValue(val)
							// Validate type before inserting
							if err := metaStore.ValidateValueType(tableName, columns[i], extractedValue); err != nil {
								return nil, nil, "", err
							}
							row[columns[i]] = extractedValue
							// Set type information
							if err := metaStore.SetColumnType(tableName, columns[i], extractedValue); err != nil {
								return nil, nil, "", err
							}
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
		// Table doesn't exist - return 0 updated rows
		return nil, nil, "UPDATE 0", nil
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
				// Validate type before updating
				if err := metaStore.ValidateValueType(tableName, colName, value); err != nil {
					return nil, nil, "", err
				}
				row[colName] = value
				// Set type information
				if err := metaStore.SetColumnType(tableName, colName, value); err != nil {
					return nil, nil, "", err
				}
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
		// Table doesn't exist - return 0 deleted rows
		return nil, nil, "DELETE 0", nil
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

	// Extract column names and types from table elements
	var columns []string
	var columnTypes []storage.ColumnType
	
	// Collect column names and types
	for _, elem := range stmt.TableElts {
		if colDef, ok := elem.Node.(*pg_query.Node_ColumnDef); ok {
			// Remove quotes if present for consistency
			colName := strings.Trim(colDef.ColumnDef.Colname, `"`)
			columns = append(columns, colName)
			
			// Extract column type if specified
			if colDef.ColumnDef.TypeName != nil {
				colType := getColumnTypeFromTypeName(colDef.ColumnDef.TypeName)
				columnTypes = append(columnTypes, colType)
			} else {
				// Default to unknown if no type specified
				columnTypes = append(columnTypes, storage.TypeUnknown)
			}
		}
	}

	// Store column names in metastore
	if len(columns) > 0 {
		metaStore.AddColumns(tableName, columns)
		
		// Store declared column types
		for i, colName := range columns {
			if i < len(columnTypes) && columnTypes[i] != storage.TypeUnknown {
				metaStore.SetColumnTypeFromSchema(tableName, colName, columnTypes[i])
			}
		}
	}

	return nil, nil, "CREATE TABLE", nil
}

func executePgDropTable(stmt *pg_query.DropStmt, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	for _, obj := range stmt.Objects {
		if list, ok := obj.Node.(*pg_query.Node_List); ok && len(list.List.Items) > 0 {
			if str, ok := list.List.Items[0].Node.(*pg_query.Node_String_); ok {
				// Remove quotes if present for consistency
				tableName := strings.Trim(str.String_.Sval, `"`)
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
		// Remove quotes if present for consistency
		return strings.Trim(n.RangeVar.Relname, `"`)
	}
	return ""
}

func extractTableNameFromRangeVar(rv *pg_query.RangeVar) string {
	if rv != nil {
		// Remove quotes if present for consistency
		return strings.Trim(rv.Relname, `"`)
	}
	return ""
}

// getColumnTypeFromTypeName extracts the column type from a TypeName node
func getColumnTypeFromTypeName(typeName *pg_query.TypeName) storage.ColumnType {
	if typeName == nil || len(typeName.Names) == 0 {
		return storage.TypeString
	}
	
	// Get the type name - it might be schema-qualified (e.g., pg_catalog.integer)
	var typeStr string
	for i, name := range typeName.Names {
		if str, ok := name.Node.(*pg_query.Node_String_); ok {
			// Take the last part (the actual type name)
			if i == len(typeName.Names)-1 {
				typeStr = strings.ToLower(str.String_.Sval)
			}
		}
	}
	
	// Map PostgreSQL types to VSQL types
	switch typeStr {
	case "bool", "boolean":
		return storage.TypeBoolean
	case "int", "int2", "int4", "int8", "integer", "smallint", "bigint":
		return storage.TypeInteger
	case "float", "float4", "float8", "real", "double", "numeric", "decimal":
		return storage.TypeFloat
	case "timestamp", "timestamptz", "date", "time", "timetz":
		return storage.TypeTimestamp
	case "text", "varchar", "char", "bpchar":
		return storage.TypeString
	default:
		return storage.TypeString
	}
}

func extractSelectColumns(stmt *pg_query.SelectStmt, tableName string, metaStore *storage.MetaStore, rows []storage.Row) []string {
	var columns []string

	for _, target := range stmt.TargetList {
		if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok {
			if resTarget.ResTarget.Name != "" {
				columns = append(columns, resTarget.ResTarget.Name)
			} else if colRef, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_ColumnRef); ok {
				if len(colRef.ColumnRef.Fields) > 0 {
					// Check if it's a star (*)
					if _, ok := colRef.ColumnRef.Fields[0].Node.(*pg_query.Node_AStar); ok {
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
					} else {
						// Extract column name (preserve qualified names like table.column)
						var parts []string
						for _, field := range colRef.ColumnRef.Fields {
							if str, ok := field.Node.(*pg_query.Node_String_); ok {
								parts = append(parts, str.String_.Sval)
							}
						}
						if len(parts) > 1 {
							// For qualified names (table.column), use the full qualified name
							columns = append(columns, strings.Join(parts, "."))
						} else if len(parts) == 1 {
							// For unqualified names, use just the column name
							columns = append(columns, parts[0])
						}
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
	case *pg_query.Node_ColumnRef:
		// Handle boolean column directly in WHERE clause
		val := extractValueFromExpr(row, whereClause)
		if val == nil {
			return false
		}
		// Try to evaluate as boolean
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
		// For other types, non-nil/non-zero values are considered true
		return true
	case *pg_query.Node_AConst:
		// Handle boolean literals (true/false)
		val := extractValueFromExpr(row, whereClause)
		if val == nil {
			return false
		}
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
		return true
	default:
		return true
	}
}

func evaluateNullTest(row storage.Row, expr *pg_query.NullTest) bool {
	var val interface{}

	// Get the column value
	if colRef, ok := expr.Arg.Node.(*pg_query.Node_ColumnRef); ok {
		if len(colRef.ColumnRef.Fields) > 0 {
			// Extract the last field as the column name (handles schema.table.column)
			lastFieldIdx := len(colRef.ColumnRef.Fields) - 1
			if str, ok := colRef.ColumnRef.Fields[lastFieldIdx].Node.(*pg_query.Node_String_); ok {
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
			// Special handling for NOT with NULL values
			arg := expr.Args[0]
			
			// Check if the argument is a column reference
			if _, ok := arg.Node.(*pg_query.Node_ColumnRef); ok {
				val := extractValueFromExpr(row, arg)
				if val == nil {
					// NOT NULL = NULL (unknown), which is treated as false in WHERE
					return false
				}
				// For boolean values, apply NOT
				if boolVal, ok := val.(bool); ok {
					return !boolVal
				}
			}
			
			// Default behavior for other cases
			return !evaluatePgWhere(row, arg)
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
		// Handle qualified (table.column or schema.table.column) and unqualified (column) references
		if len(n.ColumnRef.Fields) > 0 {
			// Extract the last field as the column name
			lastFieldIdx := len(n.ColumnRef.Fields) - 1
			if str, ok := n.ColumnRef.Fields[lastFieldIdx].Node.(*pg_query.Node_String_); ok {
				columnName := str.String_.Sval
				
				// For qualified references, try multiple variations
				if len(n.ColumnRef.Fields) >= 2 {
					// Try the full qualified name first
					var parts []string
					for i := 0; i < len(n.ColumnRef.Fields); i++ {
						if s, ok := n.ColumnRef.Fields[i].Node.(*pg_query.Node_String_); ok {
							parts = append(parts, s.String_.Sval)
						}
					}
					
					// Try progressively shorter qualified names
					for i := 0; i < len(parts)-1; i++ {
						qualifiedName := strings.Join(parts[i:], ".")
						if val, exists := row[qualifiedName]; exists {
							return val
						}
					}
				}
				
				// Fall back to unqualified column name
				return row[columnName]
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

// executePgPrepare handles PREPARE statements
func executePgPrepare(stmt *pg_query.PrepareStmt, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	// For now, we'll return an error saying PREPARE is not fully implemented
	// A complete implementation would need to store the query and handle parameter types
	return nil, nil, "", fmt.Errorf("PREPARE statement is not yet implemented")
}

// executePgExecute handles EXECUTE statements
func executePgExecute(stmt *pg_query.ExecuteStmt, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	// For now, we'll return an error saying EXECUTE is not fully implemented
	return nil, nil, "", fmt.Errorf("EXECUTE statement is not yet implemented")
}

// executePgDeallocate handles DEALLOCATE statements
func executePgDeallocate(stmt *pg_query.DeallocateStmt) ([]string, [][]interface{}, string, error) {
	// For now, we'll return an error saying DEALLOCATE is not fully implemented
	return nil, nil, "", fmt.Errorf("DEALLOCATE statement is not yet implemented")
}
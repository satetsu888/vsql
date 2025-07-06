package parser

import (
	"fmt"

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
		return executePgCreateTable(node.CreateStmt, dataStore)
	case *pg_query.Node_DropStmt:
		return executePgDropTable(node.DropStmt, dataStore, metaStore)
	default:
		return nil, nil, "", fmt.Errorf("unsupported statement type: %T", node)
	}
}

func executePgSelect(stmt *pg_query.SelectStmt, dataStore *storage.DataStore, metaStore *storage.MetaStore) ([]string, [][]interface{}, string, error) {
	if len(stmt.FromClause) != 1 {
		return nil, nil, "", fmt.Errorf("only single table SELECT is supported")
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

func executePgCreateTable(stmt *pg_query.CreateStmt, dataStore *storage.DataStore) ([]string, [][]interface{}, string, error) {
	tableName := extractTableNameFromRangeVar(stmt.Relation)
	if tableName == "" {
		return nil, nil, "", fmt.Errorf("could not extract table name")
	}

	if err := dataStore.CreateTable(tableName); err != nil {
		return nil, nil, "", err
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
							for key := range rows[0] {
								columns = append(columns, key)
							}
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
	default:
		return true
	}
}

func evaluateAExpr(row storage.Row, expr *pg_query.A_Expr) bool {
	var leftVal, rightVal interface{}

	if colRef, ok := expr.Lexpr.Node.(*pg_query.Node_ColumnRef); ok {
		if len(colRef.ColumnRef.Fields) > 0 {
			if str, ok := colRef.ColumnRef.Fields[0].Node.(*pg_query.Node_String_); ok {
				leftVal = row[str.String_.Sval]
			}
		}
	}

	rightVal = extractPgValue(expr.Rexpr)

	opName := ""
	if len(expr.Name) > 0 {
		if str, ok := expr.Name[0].Node.(*pg_query.Node_String_); ok {
			opName = str.String_.Sval
		}
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

func extractAConstValue(aConst *pg_query.A_Const) interface{} {
	if aConst.Isnull {
		return nil
	}

	switch val := aConst.Val.(type) {
	case *pg_query.A_Const_Sval:
		return val.Sval.Sval
	case *pg_query.A_Const_Ival:
		return fmt.Sprintf("%d", val.Ival.Ival)
	case *pg_query.A_Const_Fval:
		return val.Fval.Fval
	}
	return nil
}

func compareValuesPg(left interface{}, operator string, right interface{}) bool {
	if left == nil && right == nil {
		return operator == "=" || operator == "<=" || operator == ">="
	}

	if left == nil || right == nil {
		return operator == "!=" || operator == "<>"
	}

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
	}

	return false
}
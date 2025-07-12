package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"vsql/parser"
	"vsql/server"
	"vsql/storage"
)

// fileList is a custom flag type for accepting multiple file paths
type fileList []string

func (f *fileList) String() string {
	return strings.Join(*f, ",")
}

func (f *fileList) Set(value string) error {
	*f = append(*f, value)
	return nil
}

// commandList is a custom flag type for accepting multiple commands
type commandList []string

func (c *commandList) String() string {
	return strings.Join(*c, "; ")
}

func (c *commandList) Set(value string) error {
	*c = append(*c, value)
	return nil
}

func main() {
	var port int
	var commands commandList
	var filePaths fileList
	var quit bool
	var help bool
	
	flag.IntVar(&port, "port", 5432, "Port to listen on")
	flag.Var(&commands, "c", "Execute command (can be specified multiple times)")
	flag.Var(&filePaths, "f", "Execute SQL from file (can be specified multiple times)")
	flag.BoolVar(&quit, "q", false, "Quit after executing commands (don't start server)")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.Parse()
	
	if help {
		fmt.Fprintf(os.Stderr, "VSQL - A PostgreSQL-compatible, schema-less, in-memory database\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  vsql [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -port PORT    Port to listen on (default: 5432)\n")
		fmt.Fprintf(os.Stderr, "  -c COMMAND    Execute command (can be specified multiple times)\n")
		fmt.Fprintf(os.Stderr, "  -f FILE       Execute SQL from file (can be specified multiple times)\n")
		fmt.Fprintf(os.Stderr, "  -q            Quit after executing commands (don't start server)\n")
		fmt.Fprintf(os.Stderr, "  -h, -help     Show this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Start server on default port\n")
		fmt.Fprintf(os.Stderr, "  vsql\n\n")
		fmt.Fprintf(os.Stderr, "  # Start server on custom port\n")
		fmt.Fprintf(os.Stderr, "  vsql -port 5433\n\n")
		fmt.Fprintf(os.Stderr, "  # Execute query and exit\n")
		fmt.Fprintf(os.Stderr, "  vsql -c \"SELECT * FROM users;\" -q\n\n")
		fmt.Fprintf(os.Stderr, "  # Execute queries then start server (seed data)\n")
		fmt.Fprintf(os.Stderr, "  vsql -c \"CREATE TABLE users (id int, name text); INSERT INTO users VALUES (1, 'Alice');\"\n\n")
		fmt.Fprintf(os.Stderr, "  # Execute SQL from file and exit\n")
		fmt.Fprintf(os.Stderr, "  vsql -f queries.sql -q\n\n")
		fmt.Fprintf(os.Stderr, "  # Execute SQL from file as seed data then start server\n")
		fmt.Fprintf(os.Stderr, "  vsql -f seed.sql\n\n")
		fmt.Fprintf(os.Stderr, "  # Execute multiple SQL files in order\n")
		fmt.Fprintf(os.Stderr, "  vsql -f schema.sql -f data.sql -f indexes.sql -q\n\n")
		fmt.Fprintf(os.Stderr, "  # Execute multiple commands\n")
		fmt.Fprintf(os.Stderr, "  vsql -c \"CREATE TABLE t1 (id int)\" -c \"CREATE TABLE t2 (id int)\" -q\n")
		os.Exit(0)
	}

	store := storage.NewDataStore()
	metaStore := storage.NewMetaStore()

	// Execute files if provided (first)
	for _, filePath := range filePaths {
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
			os.Exit(1)
		}
		executeCommand(string(content), store, metaStore)
	}

	// Execute commands if provided (second)
	for _, command := range commands {
		executeCommand(command, store, metaStore)
	}

	// If quit flag is set, exit after executing commands
	if quit {
		return
	}

	// Otherwise, start the server
	srv := server.New(port, store, metaStore)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		srv.Stop()
		os.Exit(0)
	}()

	fmt.Printf("VSQL server starting on port %d\n", port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// splitSQLStatements splits SQL statements by semicolon while respecting comments and string literals
func splitSQLStatements(sql string) []string {
	var statements []string
	var currentStmt strings.Builder
	inSingleLineComment := false
	inMultiLineComment := false
	inSingleQuote := false
	inDoubleQuote := false
	
	runes := []rune(sql)
	for i := 0; i < len(runes); i++ {
		char := runes[i]
		
		// Check for single-line comment start
		if !inSingleQuote && !inDoubleQuote && !inMultiLineComment && i+1 < len(runes) && char == '-' && runes[i+1] == '-' {
			inSingleLineComment = true
		}
		
		// Check for multi-line comment start
		if !inSingleQuote && !inDoubleQuote && !inSingleLineComment && i+1 < len(runes) && char == '/' && runes[i+1] == '*' {
			inMultiLineComment = true
		}
		
		// Check for multi-line comment end
		if inMultiLineComment && i+1 < len(runes) && char == '*' && runes[i+1] == '/' {
			currentStmt.WriteRune(char)
			currentStmt.WriteRune(runes[i+1])
			i++
			inMultiLineComment = false
			continue
		}
		
		// Check for newline (ends single-line comment)
		if char == '\n' {
			inSingleLineComment = false
		}
		
		// Toggle string literals
		if !inSingleLineComment && !inMultiLineComment {
			if char == '\'' && !inDoubleQuote {
				// Check for escaped quote
				if i == 0 || runes[i-1] != '\\' {
					inSingleQuote = !inSingleQuote
				}
			} else if char == '"' && !inSingleQuote {
				// Check for escaped quote
				if i == 0 || runes[i-1] != '\\' {
					inDoubleQuote = !inDoubleQuote
				}
			}
		}
		
		// Check for statement separator
		if char == ';' && !inSingleLineComment && !inMultiLineComment && !inSingleQuote && !inDoubleQuote {
			stmt := currentStmt.String()
			if strings.TrimSpace(stmt) != "" {
				statements = append(statements, stmt)
			}
			currentStmt.Reset()
		} else {
			currentStmt.WriteRune(char)
		}
	}
	
	// Add the last statement if any
	stmt := currentStmt.String()
	if strings.TrimSpace(stmt) != "" {
		statements = append(statements, stmt)
	}
	
	return statements
}

func executeCommand(command string, store *storage.DataStore, metaStore *storage.MetaStore) {
	// Split multiple commands by semicolon, respecting comments
	commands := splitSQLStatements(command)
	
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}
		
		// Execute the query
		columns, rows, message, err := parser.ExecutePgQuery(cmd, store, metaStore)
		if err != nil {
			// Skip "no statements found" errors which happen with comment-only segments
			if err.Error() == "no statements found" {
				continue
			}
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		
		// Print results
		if len(columns) > 0 && len(rows) > 0 {
			// Print column headers
			fmt.Println(strings.Join(columns, "\t"))
			
			// Print separator
			separators := make([]string, len(columns))
			for i := range separators {
				separators[i] = "----"
			}
			fmt.Println(strings.Join(separators, "\t"))
			
			// Print rows
			for _, row := range rows {
				values := make([]string, len(row))
				for i, val := range row {
					if val == nil {
						values[i] = ""
					} else {
						values[i] = fmt.Sprintf("%v", val)
					}
				}
				fmt.Println(strings.Join(values, "\t"))
			}
			fmt.Printf("(%d rows)\n\n", len(rows))
		} else {
			// For non-SELECT queries, just print the message
			fmt.Println(message)
		}
	}
}
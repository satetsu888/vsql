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

func main() {
	var port int
	var command string
	var help bool
	
	flag.IntVar(&port, "port", 5432, "Port to listen on")
	flag.StringVar(&command, "c", "", "Execute command and exit")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.Parse()
	
	if help {
		fmt.Fprintf(os.Stderr, "VSQL - A PostgreSQL-compatible, schema-less, in-memory database\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  vsql [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -port PORT    Port to listen on (default: 5432)\n")
		fmt.Fprintf(os.Stderr, "  -c COMMAND    Execute command and exit\n")
		fmt.Fprintf(os.Stderr, "  -h, -help     Show this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Start server on default port\n")
		fmt.Fprintf(os.Stderr, "  vsql\n\n")
		fmt.Fprintf(os.Stderr, "  # Start server on custom port\n")
		fmt.Fprintf(os.Stderr, "  vsql -port 5433\n\n")
		fmt.Fprintf(os.Stderr, "  # Execute query and exit\n")
		fmt.Fprintf(os.Stderr, "  vsql -c \"SELECT * FROM users;\"\n\n")
		fmt.Fprintf(os.Stderr, "  # Execute multiple queries\n")
		fmt.Fprintf(os.Stderr, "  vsql -c \"CREATE TABLE users (id int, name text); INSERT INTO users VALUES (1, 'Alice');\"\n")
		os.Exit(0)
	}

	store := storage.NewDataStore()
	metaStore := storage.NewMetaStore()

	// If command is provided, execute it and exit
	if command != "" {
		executeCommand(command, store, metaStore)
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

func executeCommand(command string, store *storage.DataStore, metaStore *storage.MetaStore) {
	// Split multiple commands by semicolon
	commands := strings.Split(command, ";")
	
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}
		
		// Execute the query
		columns, rows, message, err := parser.ExecutePgQuery(cmd, store, metaStore)
		if err != nil {
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
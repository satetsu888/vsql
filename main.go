package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"vsql/server"
	"vsql/storage"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 5432, "Port to listen on")
	flag.Parse()

	store := storage.NewDataStore()
	metaStore := storage.NewMetaStore()

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
package server

import (
	"../config"
	"log"
	"net/http"
	"time"
)

var localLogger *log.Logger

func StartServer(logger *log.Logger) {
	cfg := config.GlobalConfig()
	localLogger = logger
	defer func() {
		if err := recover(); err != nil {
			logger.Print(err)
		}
	}()
	logger.Printf("Start HTTP Server: %s\n", cfg.Listen)
	server := createServer(cfg.Listen)
	err := server.ListenAndServe()
	if err != nil {
		logger.Print("Cannot Start HTTP Server")
		logger.Fatal(err)
	}
}

func createServer(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/cut", downloadCutM3u8)
	mux.HandleFunc("/download", downloadStream)
	mux.HandleFunc("/delete", deleteVideo)
	var server = http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return &server
}

package main

import (
	"./config"
	"./server"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
)

func writePIDFile(filename string) error {
	data := fmt.Sprintf("%d\n", os.Getpid())
	return ioutil.WriteFile(filename, []byte(data), 0644)
}

func processKill(enableProfile bool, logger *log.Logger) {
	killCh := make(chan os.Signal, 1)
	signal.Notify(killCh, os.Interrupt, os.Kill)
	sig := <-killCh

	if enableProfile {
		pprof.StopCPUProfile()
	}
	logger.Fatal("Stop Server ", sig)
}

func main() {
	coresPtr := flag.Int("C", 0, "Number Cores can be used, if ignored use all Cores")
	profilePtr := flag.Bool("P", false, "Enable profile, profile result file will stored to ./momo-downloader-go.prof")
	pidPtr := flag.String("p", "", "PID file path, if ignored it will not be created")
	configPtr := flag.String("c", "", "Config file path, if ignored will be load from ./config.json; /etc/momo-downloader/config.json")
	logPtr := flag.String("l", "", "Log file path, if ignored log will output to stdout")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of momo-downloader-file-go\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	var logger *log.Logger

	if *logPtr != "" {
		logWriter, err := server.NewRotateWriter(*logPtr)
		if err != nil {
			log.Fatal(err)
		}
		server.SetGlobalWriter(logWriter)
		go server.RotateScheduler(logWriter, 86400)
		logger = log.New(logWriter, "", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	}

	if *profilePtr {
		pfp, err := os.Create("momo-downloader-file-go.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(pfp)
	}

	if *coresPtr != 0 {
		logger.Printf("Use %d Cores\n", *coresPtr)
		runtime.GOMAXPROCS(*coresPtr)
	}

	if *pidPtr != "" {
		logger.Printf("Write PID to: %s\n", *pidPtr)
		err := writePIDFile(*pidPtr)
		if err != nil {
			log.Print(err)
		}
	}

	err := config.LoadConfigFile(*configPtr, logger)
	if err != nil {
		log.Fatal(err)
		return
	}

	go server.StartServer(logger)

	processKill(*profilePtr, logger)
}

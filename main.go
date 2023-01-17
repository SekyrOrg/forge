package main

import (
	"fmt"
	"github.com/projectdiscovery/goflags"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
)

func main() {
	options := ParseCLIArguments()
	for _, filepath := range options.FilePaths {
		binary, err := os.Open(filepath)
		if err != nil {
			log.Fatalf("error opening filepath: %s, %v", filepath, err)
		}
		urlPath := fmt.Sprintf(`/api/v1/upload?debug=false&static=true&upx=false&connection_string=%s&os=%s&arch=%s&transport=%s`, options.ConnectionString, runtime.GOOS, runtime.GOARCH, options.Transport)

		response, err := http.Post(options.BeaconCreatorUrl+urlPath, "application/octet-stream", binary)
		if err != nil {
			log.Fatalf("error uploading filepath: %s, %v", filepath, err)
		}
		log.Println(response.Status)
		if response.StatusCode != 200 {
			log.Fatalf("error uploading filepath: %s, %v", filepath, err)
		}
		file, err := os.OpenFile("test.binary", os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer file.Close()
		if _, err := io.Copy(file, response.Body); err != nil {
			log.Fatalf("error copying file: %v", err)
		}
	}

}

type Args struct {
	BeaconCreatorUrl string
	ConnectionString string
	Transport        string
	FilePaths        []string
	Verbose          bool
	ConfigPath       string
}

func ParseCLIArguments() *Args {
	var args Args
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`BeaconServer is a server that listens for beacon events and sends them to the backend server.`)
	flagSet.BoolVarP(&args.Verbose, "verbose", "v", false, "Enable verbose output")
	flagSet.StringVarP(&args.BeaconCreatorUrl, "beaconCreatorUrl", "u", "http://127.0.0.1:8080", "Address of the beacon creator server")
	flagSet.StringVarP(&args.ConnectionString, "connectionString", "c", "127.0.0.1:5353", "Connection string for the beacon server")
	flagSet.StringVarP(&args.Transport, "transport", "t", "http", "Transport to use for the beacon server (http, dns, tcp)")
	flagSet.StringVarP(&args.ConfigPath, "config", "C", "", "Path to the configuration file")
	flagSet.StringSliceVarP((*goflags.StringSlice)(&args.FilePaths), "files", "f", []string{}, "FilePaths to be used as beacon", goflags.StringSliceOptions)
	if err := flagSet.Parse(); err != nil {
		log.Fatal("error parsing arguments: ", err)
		return nil
	}
	if args.ConfigPath != "" {
		// check if file exists
		if _, err := os.Stat(args.ConfigPath); os.IsNotExist(err) {
			log.Fatalln("error opening config file, does not exits: ", err)
		}
		if err := flagSet.MergeConfigFile(args.ConfigPath); err != nil {
			log.Fatalln("error merging config file: ", err)
		}
	}
	if apiAddr := os.Getenv("BEACON_CREATOR_ADDR"); apiAddr != "" {
		args.BeaconCreatorUrl = apiAddr
	}
	if connectionString := os.Getenv("CONNECTION_STRING"); connectionString != "" {
		args.ConnectionString = connectionString
	}
	return &args
}

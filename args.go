package beaconforge

import (
	"github.com/projectdiscovery/goflags"
	"log"
	"os"
)

type Args struct {
	Addr           string
	FilePaths      []string
	Verbose        bool
	ConfigPath     string
	BeacponOptions BeaconOptions
}
type BeaconOptions struct {
	ConnectionString string
	Transport        string
	StaticBinary     bool
	Upx              bool
	UpxLevel         int
	Debug            bool
}

func ParseCLIArguments() *Args {
	var args Args
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`BeaconForge is a tool for generating beacons from filepaths.`)
	flagSet.StringVarP(&args.Addr, "addr", "a", "http://127.0.0.1:8080", "Address of the beaconCreator server")
	flagSet.StringSliceVarP((*goflags.StringSlice)(&args.FilePaths), "files", "f", []string{}, "File path for binaries to convert into beacon", goflags.StringSliceOptions)
	flagSet.BoolVarP(&args.Verbose, "verbose", "v", false, "Enable verbose output")
	flagSet.StringVarP(&args.ConfigPath, "config", "C", "", "Path to the configuration file")
	flagSet.CreateGroup("Beacon Options", "Options for the beacons",
		flagSet.StringVarP(&args.BeacponOptions.ConnectionString, "connectionString", "c", "127.0.0.1:5353", "Connection string for the beacon server"),
		flagSet.StringVarP(&args.BeacponOptions.Transport, "transport", "t", "http", "Transport protocol to use for the beacon server (http, dns, tcp)"),
		flagSet.BoolVarP(&args.BeacponOptions.StaticBinary, "static", "s", false, "Build a static binary"),
		flagSet.BoolVarP(&args.BeacponOptions.Upx, "upx", "u", false, "Compress the binary with upx"),
		flagSet.IntVarP(&args.BeacponOptions.UpxLevel, "upxLevel", "ul", 1, "Compression level for upx"),
		flagSet.BoolVarP(&args.BeacponOptions.Debug, "debug", "d", false, "Enable beacon debug output"),
	)
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
		args.Addr = apiAddr
	}
	if connectionString := os.Getenv("CONNECTION_STRING"); connectionString != "" {
		args.BeacponOptions.ConnectionString = connectionString
	}
	return &args
}

package forge

import (
	"github.com/SekyrOrg/creator/builder"
	"github.com/projectdiscovery/goflags"
	"log"
	"os"
	"runtime"
)

type Args struct {
	BeaconCreatorUrl string
	FilePaths        []string
	Verbose          bool
	ConfigPath       string
	OutputFolder     string
	BeaconOpts       builder.Options
}

func ParseCLIArguments() *Args {
	var args Args
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`BeaconForge is a tool for generating beacons from file paths.`)
	flagSet.StringVarP(&args.BeaconCreatorUrl, "addr", "a", "http://127.0.0.1:8080", "Address of the beaconCreator server")
	flagSet.StringSliceVarP((*goflags.StringSlice)(&args.FilePaths), "files", "f", []string{}, "File path for binaries to convert into beacon", goflags.StringSliceOptions)
	flagSet.BoolVarP(&args.Verbose, "verbose", "v", false, "Enable verbose output")
	flagSet.StringVarP(&args.OutputFolder, "output", "o", "", "Output folder for the beacons")
	flagSet.StringVarP(&args.ConfigPath, "config", "C", "", "Path to the configuration file")
	flagSet.CreateGroup("Beacon Options", "Options for the beacons",
		flagSet.StringVarP(&args.BeaconOpts.GroupId, "group-id", "id", "", "Group ID for the beacon"),
		flagSet.StringVarP(&args.BeaconOpts.ReportAddress, "connection-string", "c", "", "Connection string for the beacon"),
		flagSet.StringVar(&args.BeaconOpts.GOARCH, "arch", runtime.GOARCH, "GOARCH for the beacon"),
		flagSet.StringVar(&args.BeaconOpts.GOOS, "os", runtime.GOOS, "GOOS for the beacon"),
		flagSet.StringVar(&args.BeaconOpts.Lldflags, "lldflags", "-s -w", "Lldflags for the beacon"),
		flagSet.BoolVar(&args.BeaconOpts.StaticBinary, "static", false, "Static binary for the beacon"),
		flagSet.BoolVar(&args.BeaconOpts.Gzip, "gzip", true, "Gzip the beacon"),
		flagSet.BoolVar(&args.BeaconOpts.Upx, "upx", false, "Upx the beacon"),
		flagSet.IntVar(&args.BeaconOpts.UpxLevel, "upx-level", 1, "Upx level for the beacon"),
		flagSet.StringVar(&args.BeaconOpts.TransportTag, "transport", "dns", "Transport tag for the beacon"),
		flagSet.BoolVarP(&args.BeaconOpts.Debug, "debug", "D", false, "Enable debug output for the beacon"),
	)
	if err := flagSet.Parse(); err != nil {
		log.Fatal("error parsing arguments: ", err)
	}

	mergeConfig(args, flagSet)
	mergeEnvironment(args)
	if len(args.FilePaths) == 0 {
		log.Fatal("no files provided, use -f to provide a file paths, use , to separate multiple files")
	}

	return &args
}

func mergeEnvironment(args Args) {
	if apiAddr := os.Getenv("BEACON_CREATOR_ADDR"); apiAddr != "" {
		args.BeaconCreatorUrl = apiAddr
	}
	if connectionString := os.Getenv("CONNECTION_STRING"); connectionString != "" {
		args.BeaconOpts.ReportAddress = connectionString
	}
}

func mergeConfig(args Args, flagSet *goflags.FlagSet) {
	// merge config file
	if args.ConfigPath == "" {
		return
	}
	// check if file exists
	if _, err := os.Stat(args.ConfigPath); os.IsNotExist(err) {
		log.Fatalln("error opening config file, does not exits: ", err)
	}
	if err := flagSet.MergeConfigFile(args.ConfigPath); err != nil {
		log.Fatalln("error merging config file: ", err)
	}
}

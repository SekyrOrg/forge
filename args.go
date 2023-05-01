package forge

import (
	"github.com/SekyrOrg/forge/openapi"
	"github.com/google/uuid"
	"github.com/projectdiscovery/goflags"
	"log"
	"os"
	"runtime"
)

type Args struct {
	CreatorUrl   string
	FilePaths    []string
	Verbose      bool
	ConfigPath   string
	OutputFolder string
	Overwrite    bool
	BeaconOpts   beaconOptions
}

func ParseCLIArguments() *Args {
	var args Args
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`Forge is a tool for generating beacons from file paths.`)
	flagSet.StringVarP(&args.CreatorUrl, "addr", "a", "https://gateway.sekyr.com", "Address of the gateway server")
	flagSet.StringSliceVarP((*goflags.StringSlice)(&args.FilePaths), "files", "f", []string{}, "File path for binaries to convert into beacon", goflags.StringSliceOptions)
	flagSet.BoolVarP(&args.Verbose, "verbose", "v", false, "Enable verbose output")
	flagSet.BoolVarP(&args.Overwrite, "overwrite", "w", false, "Overwrite the beacons, this overrides the output flag")
	flagSet.StringVarP(&args.OutputFolder, "output", "o", "out", "Output folder for the beacons, if not provided beacons are overwritten")
	flagSet.StringVarP(&args.ConfigPath, "config", "C", "", "Path to the configuration file")
	flagSet.CreateGroup("Beacon Options", "Options for the beacons",
		flagSet.StringVarP(&args.BeaconOpts.GroupId, "group-id", "id", "", "Group ID for the beacon"),
		flagSet.StringVarP(&args.BeaconOpts.ReportAddr, "connection-string", "c", "sekyr.com:5353", "Connection string for the beacon"),
		flagSet.StringVar(&args.BeaconOpts.Arch, "arch", runtime.GOARCH, "GOARCH for the beacon"),
		flagSet.StringVar(&args.BeaconOpts.Os, "os", runtime.GOOS, "GOOS for the beacon"),
		flagSet.BoolVar(&args.BeaconOpts.Upx, "upx", false, "Upx the beacon"),
		flagSet.IntVar(&args.BeaconOpts.UpxLevel, "upx-level", 1, "Upx level for the beacon"),
		flagSet.StringVar(&args.BeaconOpts.Transport, "transport", "dns", "Transport tag for the beacon"),
		flagSet.BoolVarP(&args.BeaconOpts.Debug, "debug", "D", false, "Enable debug output for the beacon"),
	)
	if err := flagSet.Parse(); err != nil {
		log.Fatal("error parsing arguments: ", err)
	}
	args.BeaconOpts.Lldflags = "-s -w"
	args.BeaconOpts.Static = true
	args.BeaconOpts.Gzip = true

	mergeConfig(args, flagSet)
	mergeEnvironment(args)
	if len(args.FilePaths) == 0 {
		log.Fatal("no files provided, use -f to provide a file paths, use , to separate multiple files")
	}

	return &args
}

func mergeEnvironment(args Args) {
	if apiAddr := os.Getenv("CREATOR_ADDR"); apiAddr != "" {
		args.CreatorUrl = apiAddr
	}
	if connectionString := os.Getenv("REPORT_ADDR"); connectionString != "" {
		args.BeaconOpts.ReportAddr = connectionString
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

type beaconOptions struct {
	ReportAddr string
	Os         string
	Arch       string
	GroupId    string
	Static     bool
	Upx        bool
	UpxLevel   int
	Gzip       bool
	Debug      bool
	Lldflags   string
	Transport  string
}

func (b *beaconOptions) toPostCreatorParams() *openapi.PostCreatorParams {
	params := openapi.PostCreatorParams{
		ReportAddr: b.ReportAddr,
		Os:         b.Os,
		Arch:       b.Arch,
	}
	if b.GroupId != "" {
		groupId, err := uuid.Parse(b.GroupId)
		if err != nil {
			log.Fatal(err)
		}
		params.GroupUuid = &groupId
	}
	if b.Static {
		params.Static = &b.Static
	}
	if b.Upx {
		params.Upx = &b.Upx
	}
	if b.UpxLevel != 0 {
		params.UpxLevel = &b.UpxLevel
	}
	if b.Gzip {
		params.Gzip = &b.Gzip
	}
	if b.Debug {
		params.Debug = &b.Debug
	}
	if b.Lldflags != "" {
		params.Lldflags = &b.Lldflags
	}
	if b.Transport != "" {
		params.Transport = &b.Transport
	}
	return &params
}

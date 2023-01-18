package main

import (
	"fmt"
	"github.com/SekyrOrg/beaconforge"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"os"
	"runtime"
)

func main() {
	zapLogger := CreateZapLogger()
	defer zapLogger.Sync()

	arguments := beaconforge.ParseCLIArguments()
	zapLogger.With(zap.Strings("files", arguments.FilePaths)).Info("beaconForge Starting")
	zapLogger.With(zap.Any("arguments", arguments)).Debug("Arguments:")
	if err := Run(zapLogger, arguments); err != nil {
		zapLogger.Fatal(fmt.Sprintf("beaconForge encountered an error: %s", err))
	}
	os.Exit(0)
}

func Run(logger *zap.Logger, options *beaconforge.Args) error {
	for _, filepath := range options.FilePaths {
		if err := CreateBeacon(logger, options, filepath); err != nil {
			if options.ContinueOnFailure {
				logger.Warn(fmt.Sprintf("Error creating beacon for %s: %s", filepath, err))
				continue
			}
			return fmt.Errorf("error creating beacon: %s", err)
		}
	}
	return nil
}

func CreateBeacon(logger *zap.Logger, args *beaconforge.Args, filepath string) error {
	logger.Info(fmt.Sprintf(`Creating beacon for "%s"`, filepath))
	binary, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening file: %s", err)
	}

	urlPath := fmt.Sprintf(`/api/v1/upload?debug=%t&static=%t&upx=%t&upx_level=%d&connection_string=%s&os=%s&arch=%s&transport=%s`, args.BeaconOptions.Debug, args.BeaconOptions.StaticBinary, args.BeaconOptions.Upx, args.BeaconOptions.UpxLevel, args.BeaconOptions.BeaconServerUrl, runtime.GOOS, runtime.GOARCH, args.BeaconOptions.Transport)
	url := fmt.Sprintf("%s%s", args.BeaconCreatorUrl, urlPath)
	logger.With(zap.String("url", url)).Debug("URL:")
	response, err := http.Post(url, "application/octet-stream", binary)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	logger.With(zap.Int("status", response.StatusCode)).Debug("Response:")
	if response.StatusCode != 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %s", err)
		}
		return fmt.Errorf("error uploading beacon, status: %s, body: %s", response.Status, body)
	}
	originalBinary, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("error opening file: %s", err)
	}

	if _, err := io.Copy(originalBinary, response.Body); err != nil {
		return fmt.Errorf("error copying response body to file: %s", err)
	}
	logger.Info(fmt.Sprintf(`Successfully created beacon "%s"`, filepath))
	return nil
}

func CreateZapLogger() *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:       "message",
		ConsoleSeparator: " ",
		LevelKey:         "level",
		EncodeLevel:      zapcore.CapitalColorLevelEncoder,
		TimeKey:          "time",
		EncodeTime:       zapcore.TimeEncoderOfLayout("15:04:05"),
	}

	zapLogger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	))

	return zapLogger
}

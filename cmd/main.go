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
			return fmt.Errorf("error creating beacon: %s", err)
		}
	}
	return nil
}

func CreateBeacon(logger *zap.Logger, options *beaconforge.Args, filepath string) error {
	logger.Info(fmt.Sprintf("Creating beacon for %s", filepath))
	binary, err := os.OpenFile(filepath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("error opening file: %s", err)
	}

	//defer binary.Close()

	urlPath := fmt.Sprintf(`/api/v1/upload?debug=%t&static=%t&upx=%t&upx_level=%d&connection_string=%s&os=%s&arch=%s&transport=%s`, options.BeacponOptions.Debug, options.BeacponOptions.StaticBinary, options.BeacponOptions.Upx, options.BeacponOptions.UpxLevel, options.BeacponOptions.ConnectionString, runtime.GOOS, runtime.GOARCH, options.BeacponOptions.Transport)
	url := fmt.Sprintf("%s%s", options.Addr, urlPath)
	logger.Debug(fmt.Sprintf("Uploading beacon to %s", url))
	response, err := http.Post(url, "application/octet-stream", binary)
	if err != nil {
		return err
	}
	//defer response.Body.Close()
	logger.Debug(fmt.Sprintf("Response: %s", response.Status))
	if response.StatusCode != 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %s", err)
		}
		return fmt.Errorf("error uploading beacon, status: %s, body: %s", response.Status, body)
	}
	logger.Debug(fmt.Sprintf("Successfully created beacon for %s", filepath))
	if _, err := binary.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("error seeking to beginning of file: %s", err)
	}
	if _, err := io.Copy(binary, response.Body); err != nil {
		return fmt.Errorf("error copying response body to file: %s", err)
	}
	logger.Info(fmt.Sprintf("Successfully created beacon %s", filepath))
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

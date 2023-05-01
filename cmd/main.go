package main

import (
	"github.com/SekyrOrg/forge"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"runtime"
)

func main() {
	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs * 2)
	logger := CreateZapLogger()
	defer logger.Sync()

	arguments := forge.ParseCLIArguments()
	logger.
		With(zap.Strings("files", arguments.FilePaths)).
		Info("beaconForge Starting")

	runner, err := forge.NewRunner(logger, arguments)
	if err != nil {
		logger.Fatal("error creating runner", zap.Error(err))
	}

	if err := runner.Run(); err != nil {
		logger.Fatal("beaconForge encountered an error", zap.Error(err))
	}
	logger.Info("beaconForge finished successfully")

}

func CreateZapLogger() *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:       "msg",
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

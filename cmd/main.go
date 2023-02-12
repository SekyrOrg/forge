package main

import (
	"fmt"
	"github.com/SekyrOrg/forge"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func main() {
	zapLogger := CreateZapLogger()
	defer zapLogger.Sync()

	arguments := forge.ParseCLIArguments()
	zapLogger.
		With(zap.Strings("files", arguments.FilePaths)).
		Info("beaconForge Starting")

	runner := forge.NewRunner(zapLogger, arguments)

	if err := runner.Run(); err != nil {
		zapLogger.Fatal(fmt.Sprintf("beaconForge encountered an error: %s", err))
	}
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

package main

import (
	"github.com/SekyrOrg/forge"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"os/exec"
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

	if runtime.GOOS == "windows" {
		dirPath := "C:\\Program Files\\Sekyr"
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			logger.Fatal("error creating directory", zap.Error(err))
			return
		}

		// Add permission for any user to execute a binary located inside the directory
		cmd := exec.Command("icacls", dirPath, "/grant", "*S-1-1-0:(OI)(CI)RX")
		if output, err := cmd.CombinedOutput(); err != nil {
			logger.Fatal("error setting ACL", zap.Error(err), zap.String("output", string(output)))
			return
		}
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

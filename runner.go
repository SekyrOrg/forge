package forge

import (
	"context"
	"fmt"
	"github.com/SekyrOrg/forge/openapi"
	"github.com/sourcegraph/conc/iter"
	"go.uber.org/zap"
	"io"
	"net/http"
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
	client *openapi.Client
}

func NewRunner(logger *zap.Logger, args *Args) (*Runner, error) {
	client, err := openapi.NewClient(args.CreatorUrl)
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}
	return &Runner{
		logger: logger,
		args:   args,
		client: client,
	}, nil
}

func (r *Runner) Run() error {
	r.logger.With(zap.Any("arguments", r.args)).Debug("Starting Runner")
	binaryFiles, err := iter.MapErr(r.args.FilePaths, r.CreateBinary)
	if err != nil {
		return fmt.Errorf("error creating temp binary, file: %s, error: %s", r.args.FilePaths, err)
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
	responseBody, err := r.sendBinary(*filePath)
	if err != nil {
		return TempBinary{}, fmt.Errorf("error sending binary: %w", err)
	}
	defer responseBody.Close()

	return r.createTempBinaryFile(filePath, responseBody)
}

// createTempBinaryFile creates a temporary file from the given response body
func (r *Runner) createTempBinaryFile(filePath *string, responseBody io.Reader) (TempBinary, error) {
	r.logger.With(zap.String("file", *filePath)).Debug("Creating temp file")
	tempFile, err := os.CreateTemp(os.TempDir(), path.Base(*filePath))
	if err != nil {
		return TempBinary{}, fmt.Errorf("error creating temp file: %w", err)
	}
	defer tempFile.Close()

	if _, err = io.Copy(tempFile, responseBody); err != nil {
		return TempBinary{}, fmt.Errorf("error copying binary to temp file: %w", err)
	}

	return TempBinary{
		filePath:     *filePath,
		tempFilePath: tempFile.Name(),
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

	destinationFilePath := r.getDestinationFilePath(file)
	destinationFile, err := os.OpenFile(destinationFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		r.logger.Fatal(fmt.Sprintf("error opening original file: %r", err))
	}
	defer destinationFile.Close()

	if _, err = io.Copy(destinationFile, tempFile); err != nil {
		r.logger.Fatal(fmt.Sprintf("error copying temp file to original file: %r", err))
	}
}

// getDestinationFilePath returns the path for the destination file, based on user-specified output folder
func (r *Runner) getDestinationFilePath(file *TempBinary) string {
	if r.args.OutputFolder != "" {
		if err := os.MkdirAll(r.args.OutputFolder, 0755); err != nil {
			r.logger.Fatal(fmt.Sprintf("error creating output folder: %r", err))
		}
		return fmt.Sprintf("%s/%s", r.args.OutputFolder, path.Base(file.filePath))
	}
	return file.filePath
}

// sendBinary sends the binary to the beaconCreator and returns the response body
func (r *Runner) sendBinary(filepath string) (io.ReadCloser, error) {
	binary, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer binary.Close()

	response, err := r.client.PostCreatorWithBody(context.Background(), r.args.BeaconOpts.toPostCreatorParams(), "application/octet-stream", binary)
	if err != nil {
		return nil, fmt.Errorf("error sending binary: %w", err)
	}

	return r.checkResponseStatus(response)
}

// checkResponseStatus checks the response status and returns the response body if successful
func (r *Runner) checkResponseStatus(response *http.Response) (io.ReadCloser, error) {
	if response.StatusCode != 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body, err: %w", err)
		}
		return nil, fmt.Errorf("error uploading beacon, status: %s, body: %s", response.Status, string(body))
	}
	r.logger.Debug("Successfully created binary")
	return response.Body, nil
}

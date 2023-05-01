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
	originalFilePath string
	tempFilePath     *os.File
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
	// delete all temp files once done
	defer iter.ForEach(binaryFiles, func(filePointer **TempBinary) {
		os.Remove((*filePointer).tempFilePath.Name())
	})
	iter.ForEach(binaryFiles, r.OverwriteBinary)

	return nil
}

// CreateBinary sends the binary to the beaconCreator and stores the beacon in a temporary file
// Returns the path to the temporary file and the path to the original file
func (r *Runner) CreateBinary(filePathPointer *string) (*TempBinary, error) {
	filePath := *filePathPointer
	responseBody, err := r.sendBinary(filePath)
	if err != nil {
		return nil, fmt.Errorf("error sending binary: %w", err)
	}
	defer responseBody.Close()

	return r.createTempBinaryFile(filePath, responseBody)
}

// createTempBinaryFile creates a temporary file from the given response body
func (r *Runner) createTempBinaryFile(filePath string, responseBody io.Reader) (*TempBinary, error) {
	r.logger.With(zap.String("file", filePath)).Debug("Creating temp file")
	tempFile, err := os.CreateTemp(os.TempDir(), path.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %w", err)
	}

	if _, err = io.Copy(tempFile, responseBody); err != nil {
		return nil, fmt.Errorf("error copying binary to temp file: %w", err)
	}

	return &TempBinary{
		originalFilePath: filePath,
		tempFilePath:     tempFile,
	}, nil
}

// OverwriteBinary overwrites the original binary with the beacon stored in the temporary file
func (r *Runner) OverwriteBinary(filePointer **TempBinary) {
	file := *filePointer
	r.logger.
		With(
			zap.String("tempFilePath", file.tempFilePath.Name()),
			zap.String("destinationFilePath", file.originalFilePath)).
		Info("Overwriting binary")
	if err := r.CopyFilePermissions(file.originalFilePath, file.tempFilePath); err != nil {
		r.logger.Fatal(fmt.Sprintf("error copying file permissions: %s", err))
		return
	}
	if err := os.Rename(file.tempFilePath.Name(), file.originalFilePath); err != nil {
		r.logger.Fatal(fmt.Sprintf("error renaming temp file to original file: %s", err))
	}
}

func (r *Runner) checkResponseStatus(response *http.Response) (io.ReadCloser, error) {
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %s", response.Status)
	}
	return response.Body, nil
}

// getDestinationFilePath returns the path for the destination file, based on user-specified output folder
func (r *Runner) getDestinationFilePath(file *TempBinary) string {
	if r.args.Overwrite {
		return file.originalFilePath
	}
	if err := os.MkdirAll(r.args.OutputFolder, 0755); err != nil {
		r.logger.Fatal(fmt.Sprintf("error creating output folder: %s", err))
	}
	return path.Join(r.args.OutputFolder, path.Base(file.originalFilePath))
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

// CopyFilePermissions copies the file permissions from the original file to the temporary file
func (r *Runner) CopyFilePermissions(originalFile string, tempfile *os.File) error {
	originalFileInfo, err := os.Stat(originalFile)
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}
	if err := tempfile.Chmod(originalFileInfo.Mode()); err != nil {
		return fmt.Errorf("error changing file permissions: %w", err)
	}
	return nil
}

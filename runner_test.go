package beaconforge

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRunner_sendBinary(t *testing.T) {
	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("testServer got request %s\n", r.URL)
		w.Write([]byte("test"))
	}))
	testServer.Start()
	defer testServer.Close()
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	t.Run("Runner_sendBinary works", func(t *testing.T) {
		runner := Runner{logger: logger, args: &Args{BeaconCreatorUrl: testServer.URL}}
		testFile := createAndWriteTempFile(t, "test")
		defer os.Remove(testFile)

		r, err := runner.sendBinary(testFile)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		content, err := io.ReadAll(r)
		assert.NoError(t, err)
		assert.Equal(t, content, []byte("test"), "content of returned reader should be content returned by testServer")
	})
}

func TestRunner_CreateBinary(t *testing.T) {
	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("testServer got request %s\n", r.URL)
		w.Write([]byte("test"))
	}))
	testServer.Start()
	defer testServer.Close()
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	t.Run("Runner_CreateBinary works", func(t *testing.T) {
		runner := Runner{logger: logger, args: &Args{BeaconCreatorUrl: testServer.URL}}
		testFile := createAndWriteTempFile(t, "test")
		defer os.Remove(testFile)

		binary, err := runner.CreateBinary(&testFile)
		assert.NoError(t, err, "error should be nil")
		assert.NotNil(t, binary, "binary should not be nil")
		assert.NotEmpty(t, binary.filePath, "filePath should be set")
		assert.Equal(t, binary.filePath, testFile, "filePath should be set to testFile")
		assert.NotEmpty(t, binary.tempFilePath, "tempFilePath should be set")
		assert.FileExistsf(t, binary.tempFilePath, "tempFilePath should be a file")

	})
	t.Run("Runner_CreateBinary tempfile contains the content sent from testServer", func(t *testing.T) {
		runner := Runner{logger: logger, args: &Args{BeaconCreatorUrl: testServer.URL}}
		testFile := createAndWriteTempFile(t, "test")
		defer os.Remove(testFile)

		binary, err := runner.CreateBinary(&testFile)
		assert.NoError(t, err, "error should be nil")
		assert.NotNil(t, binary, "binary should not be nil")
		assert.FileExists(t, binary.tempFilePath, "tempFilePath should be a file")
		tempFileContend, err := os.ReadFile(binary.tempFilePath)
		assert.NoError(t, err, "error should be nil")
		assert.Equal(t, tempFileContend, []byte("test"), "content of temp file should be content returned by testServer")
	})
}

func TestRunner_OverwriteBinary(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	t.Run("Runner_OverwriteBinary overwrites destination with tempFile ", func(t *testing.T) {

		runner := Runner{logger: logger, args: &Args{}}
		// create temp file
		tempFile := createAndWriteTempFile(t, "temp")
		defer os.Remove(tempFile)
		destinationFile := createAndWriteTempFile(t, "destination")
		defer os.Remove(destinationFile)

		tempBinary := &TempBinary{
			filePath:     destinationFile,
			tempFilePath: tempFile,
		}
		runner.OverwriteBinary(tempBinary)
		tempFileContend, err := os.ReadFile(tempFile)
		assert.NoError(t, err, "error should be nil")
		destinationFileContend, err := os.ReadFile(destinationFile)
		assert.NoError(t, err, "error should be nil")
		assert.Equal(t, tempFileContend, destinationFileContend, "tempFile and destinationFile should be equal")
	})

	t.Run("Runner_OverwriteBinary temp written to outFolder if specified ", func(t *testing.T) {

		outdir, err := os.MkdirTemp(os.TempDir(), "outDir")
		assert.NoError(t, err, "error should be nil")
		runner := Runner{logger: logger, args: &Args{OutputFolder: outdir}}
		// create temp file

		tempFile := createAndWriteTempFile(t, "temp")
		defer os.Remove(tempFile)
		destinationFile := createAndWriteTempFile(t, "destination")
		defer os.Remove(destinationFile)

		tempBinary := &TempBinary{
			filePath:     destinationFile,
			tempFilePath: tempFile,
		}
		runner.OverwriteBinary(tempBinary)

		tempFileContend, err := os.ReadFile(tempFile)
		assert.NoError(t, err, "error should be nil")
		outDirDestinationPath := filepath.Join(outdir, filepath.Base(destinationFile))
		outDirDestinationFileContend, err := os.ReadFile(outDirDestinationPath)
		assert.NoError(t, err, "error should be nil")
		assert.Equal(t, tempFileContend, outDirDestinationFileContend, "tempFile and outDirDestinationFile should be equal")
	})
	t.Run("Runner_OverwriteBinary destination file not overwritten when utFolder if specified ", func(t *testing.T) {

		outdir, err := os.MkdirTemp(os.TempDir(), "outDir")
		assert.NoError(t, err, "error should be nil")
		runner := Runner{logger: logger, args: &Args{OutputFolder: outdir}}
		// create temp file
		tempFile := createAndWriteTempFile(t, "temp")
		defer os.Remove(tempFile)
		destinationFile := createAndWriteTempFile(t, "destination")
		defer os.Remove(destinationFile)

		tempBinary := &TempBinary{
			filePath:     destinationFile,
			tempFilePath: tempFile,
		}
		runner.OverwriteBinary(tempBinary)

		tempFileContend, err := os.ReadFile(tempFile)
		assert.NoError(t, err, "error should be nil")
		destinationFileContend, err := os.ReadFile(destinationFile)
		assert.NoError(t, err, "error should be nil")
		assert.NotEqual(t, tempFileContend, destinationFileContend, "tempFile and destinationFile should be equal")
	})
}

func createAndWriteTempFile(t *testing.T, nameAndContent string) string {
	t.Helper()
	tempFile, err := os.CreateTemp(os.TempDir(), nameAndContent)
	assert.NoError(t, err, "error should be nil")
	defer tempFile.Close()
	_, err = tempFile.Write([]byte(nameAndContent))
	assert.NoError(t, err, "error should be nil")
	return tempFile.Name()
}

func TestRunner_Run(t *testing.T) {
	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("testServer got request %s\n", r.URL)
		w.Write([]byte("test"))
	}))
	testServer.Start()
	defer testServer.Close()
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	t.Run("Runner_Run creates binary from path and overwrites it", func(t *testing.T) {
		tempFile1 := createAndWriteTempFile(t, "temp1")
		defer os.Remove(tempFile1)
		tempFile2 := createAndWriteTempFile(t, "temp2")
		defer os.Remove(tempFile2)
		runner := Runner{logger: logger, args: &Args{BeaconCreatorUrl: testServer.URL, FilePaths: []string{tempFile1, tempFile2}}}
		err := runner.Run()
		assert.NoError(t, err, "error should be nil")
		assert.FileExists(t, tempFile1, "tempFile1 should be a file")
		assert.FileExists(t, tempFile2, "tempFile2 should be a file")
		tempFile1Content, err := os.ReadFile(tempFile1)
		assert.NoError(t, err, "error should be nil")
		assert.Equal(t, tempFile1Content, []byte("test"), "content of tempFile1 should be content returned by testServer")

		tempFile2Content, err := os.ReadFile(tempFile2)
		assert.NoError(t, err, "error should be nil")
		assert.Equal(t, tempFile2Content, []byte("test"), "content of tempFile2 should be content returned by testServer")
	})
}

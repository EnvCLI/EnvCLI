package aliases

import (
	"io"
	"net/http"
	"os"

	sentry "github.com/EnvCLI/EnvCLI/pkg/sentry"
)

/**
 * DownloadFile will download a url to a local file.
 */
func DownloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		sentry.HandleError(err)
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		sentry.HandleError(err)
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		sentry.HandleError(err)
		return err
	}

	return nil
}

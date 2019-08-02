package build

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gosimple/slug"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"
)

func getBinary(projectRoot, u string, writer io.Writer) (binaryLocation string, _ error) {

	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	binaryLocation = fmt.Sprintf("%s/.velocityci/plugins/%s", projectRoot, slug.Make(parsedURL.Path))

	if _, err := os.Stat(binaryLocation); os.IsNotExist(err) {
		logging.GetLogger().Debug("downloading binary", zap.String("from", u), zap.String("to", binaryLocation))
		writer.Write([]byte(fmt.Sprintf("Downloading binary: %s", parsedURL.String())))
		outFile, err := os.Create(binaryLocation)
		if err != nil {
			return "", err
		}
		defer outFile.Close()
		resp, err := http.Get(u)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		size, err := io.Copy(outFile, resp.Body)
		if err != nil {
			return "", err
		}
		writer.Write([]byte(fmt.Sprintf(
			"Downloaded binary: %s to %s. %d bytes",
			parsedURL.String(),
			binaryLocation,
			size,
		)))

		logging.GetLogger().Debug("downloaded binary", zap.String("from", u), zap.String("to", binaryLocation), zap.Int64("bytes", size))
		outFile.Chmod(os.ModePerm)
	}

	return binaryLocation, nil
}

package main

import (
	"context"
	"fmt"
	"gopenApiToGraphQL/util"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

// getOas downloads the spec to a file if identifier is a web address or checks if the file exists and uniforms it to absolute path
func getOas(ctx context.Context, identifier string) (string, error) {
	if strings.HasPrefix(identifier, "http://") || strings.HasPrefix(identifier, "https://") {
		// download it
		return downloadOas(ctx, identifier)
	}
	// relative -> absolute filepath
	oasFile, err := util.ToAbsolutePath(identifier)
	if err != nil {
		return "", err
	}
	// check if exists
	exists, err := util.FileExists(oasFile)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("%s does not exist", oasFile)
	}

	return oasFile, nil
}

func downloadOas(ctx context.Context, url string) (string, error) {
	// create the request and fetch the data
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("GET %s - could not create request: %w", url, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("GET %s - could not fetch data: %w", url, err)
	}
	defer res.Body.Close()

	// check if server seems happy
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("GET %s - server resonded with %s", url, res.Status)
	}

	// load content to buffer
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("GET %s - could not read response body: %w", url, err)
	}

	// check if it is json or yaml

	contentType := res.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/vnd.oai.openapi+json") {
		outFile := path.Join(util.TmpDir, "openapi.json")
		err := os.WriteFile(outFile, body, 0644)
		if err != nil {
			return "", fmt.Errorf("could not write %s: %w", outFile, err)
		}
		return outFile, nil
	}

	if strings.Contains(contentType, "application/yaml") ||
		strings.Contains(contentType, "text/yaml") ||
		strings.Contains(contentType, "application/vnd.oai.openapi") {
		outFile := path.Join(util.TmpDir, "openapi.yaml")
		err := os.WriteFile(outFile, body, 0644)
		if err != nil {
			return "", fmt.Errorf("could not write %s: %w", outFile, err)
		}
		return outFile, nil
	}

	return "", fmt.Errorf("GET %s - response has no accepted content-type", url)
}

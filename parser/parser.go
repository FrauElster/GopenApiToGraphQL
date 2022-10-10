package parser

import (
	"context"
	"fmt"
	"github.com/FrauElster/gopenApiToGraphQL/util"
	"github.com/getkin/kin-openapi/openapi3"
	"io"
	"net/http"
	"os"
	"strings"
)

func Parse(oasFile string) (GqlSpec, error) {
	// get the data either from download or reading file
	oasSpec, err := getOas(context.Background(), oasFile)
	if err != nil {
		return GqlSpec{}, err
	}

	// parse the file to OAS representation
	doc, err := openapi3.NewLoader().LoadFromData(oasSpec)
	if err != nil {
		return GqlSpec{}, fmt.Errorf("could not parse OAS: %s", err)
	}
	err = doc.Validate(context.Background())
	if err != nil {
		return GqlSpec{}, fmt.Errorf("invalid OAS: %s", err)
	}

	// parse types
	gqlTypes, err := parseSchema(doc)
	if err != nil {
		return GqlSpec{}, err
	}

	// actually scalars are only types without attributes, we have to separate them
	gqlScalars := make([]GqlScalar, 0)
	gqlTypes = util.FilterSlice(gqlTypes, func(t GqlType) bool {
		if len(t.Attributes) == 0 {
			gqlScalars = append(gqlScalars, GqlScalar{Name: t.Name})
			return false
		}
		return true
	})

	queries := make([]GqlOperation, 0)
	mutations := make([]GqlOperation, 0)
	for url, path := range doc.Paths {
		if path.Get != nil {
			query, err := parseOperation(*path.Get, url, oasGet)
			if err != nil {
				return GqlSpec{}, fmt.Errorf("could not parse query: %w", err)
			}
			queries = append(queries, query)
		}
		if path.Delete != nil {
			mutation, err := parseOperation(*path.Delete, url, oasDelete)
			if err != nil {
				return GqlSpec{}, fmt.Errorf("could not parse mutation: %w", err)
			}
			mutations = append(mutations, mutation)
		}
		if path.Post != nil {
			mutation, err := parseOperation(*path.Post, url, oasPost)
			if err != nil {
				return GqlSpec{}, fmt.Errorf("could not parse mutation: %w", err)
			}
			mutations = append(mutations, mutation)
		}
		if path.Put != nil {
			mutation, err := parseOperation(*path.Put, url, oasPut)
			if err != nil {
				return GqlSpec{}, fmt.Errorf("could not parse mutation: %w", err)
			}
			mutations = append(mutations, mutation)
		}
	}

	return GqlSpec{
		Types:     gqlTypes,
		Mutations: mutations,
		Scalars:   gqlScalars,
		Queries:   queries,
	}, nil
}

// getOas downloads the spec to a file if identifier is a web address or checks if the file exists and uniforms it to absolute path
func getOas(ctx context.Context, identifier string) ([]byte, error) {
	if strings.HasPrefix(identifier, "http://") || strings.HasPrefix(identifier, "https://") {
		// download it
		return downloadOas(ctx, identifier)
	}
	// relative -> absolute filepath
	oasFile, err := util.ToAbsolutePath(identifier)
	if err != nil {
		return nil, err
	}
	// check if exists
	exists, err := util.FileExists(oasFile)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("%s does not exist", oasFile)
	}
	// load contents
	dat, err := os.ReadFile(oasFile)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", oasFile, err)
	}
	return dat, nil
}

func downloadOas(ctx context.Context, url string) ([]byte, error) {
	// create the request and fetch the data
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("GET %s - could not create request: %w", url, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s - could not fetch data: %w", url, err)
	}
	defer res.Body.Close()

	// check if server seems happy
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("GET %s - server resonded with %s", url, res.Status)
	}

	// check content type
	acceptedTypes := []string{"application/json", "application/vnd.oai.openapi+json", "application/yaml", "text/yaml", "application/vnd.oai.openapi"}
	contentType := res.Header.Get("Content-Type")
	hasValidContentType := false
	for _, acceptedType := range acceptedTypes {
		if strings.Contains(contentType, acceptedType) {
			hasValidContentType = true
			break
		}
	}
	if !hasValidContentType {
		return nil, fmt.Errorf("GET %s - response has no accepted content-type", url)
	}

	// load content to buffer
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("GET %s - could not read response body: %w", url, err)
	}
	return body, nil
}

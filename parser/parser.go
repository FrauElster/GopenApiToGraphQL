package parser

import (
	"context"
	"fmt"
	"github.com/FrauElster/gopenApiToGraphQL/util"
	"github.com/getkin/kin-openapi/openapi3"
)

func Parse(oasFile string) (GqlSpec, error) {
	// parse the file to OAS representation
	doc, err := openapi3.NewLoader().LoadFromFile(oasFile)
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

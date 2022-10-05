package parser

import (
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
	"oasToGraphql/util"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// parseOperation will parse any OpenAPI operation to a GqlOperation
func parseOperation(oasOperation openapi3.Operation, url string, kind oasOperationKind) (GqlOperation, error) {
	// converting name
	// we always want to use the operationID as the name...but it is sadly not a mandatory attribute
	// therefor we will use the url as fallback
	name := url
	if oasOperation.OperationID != "" {
		name = oasOperation.OperationID
	} else {
		log.Warnf("%s %s - has no operationID specified, using path", kind, url)
	}
	// we want to sanitize the name, because in GraphQL it has to be camelCase without special chars
	name = toCamelCase(name)

	// converting hints
	// fixme here is probably even more one could add, I just stumbled across this and is was easy enough to add
	hints := make([]string, 0)
	if oasOperation.Deprecated {
		hints = append(hints, gqlDeprecated)
	}

	// converting response
	returnType, err := parseResponse(oasOperation)
	if err != nil {
		return GqlOperation{}, fmt.Errorf("%s %s - could not convert response: %w", kind, url, err)
	}

	// converting parameters
	params, err := parseParameters(oasOperation)
	if err != nil {
		return GqlOperation{}, fmt.Errorf("%s %s - could not convert parameters: %w", kind, url, err)
	}

	return GqlOperation{
		Origin:     fmt.Sprintf("%s - %s", kind, url),
		Name:       name,
		Parameters: params,
		ReturnType: returnType,
		Hints:      hints,
	}, nil
}

var noSchemaError = errors.New("no valid schema")

var cleanNameReg = regexp.MustCompile("[^a-zA-Z0-9]")

func toCamelCase(name string) string {
	// split on every non character or integer
	parts := cleanNameReg.Split(name, -1)
	// upper each starting char for every word except the first one
	for idx := 1; idx < len(parts); idx++ {
		// first char to upper case https://stackoverflow.com/a/70259366
		r := []rune(parts[idx])
		parts[idx] = string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
	}

	return strings.Join(parts, "")
}

// maps the parameters to GqlAttribute s
func parseParameters(oasOperation openapi3.Operation) ([]GqlAttribute, error) {
	gqlParams := make([]GqlAttribute, 0, len(oasOperation.Parameters))

	for oasParamIdx := range oasOperation.Parameters {
		oasParam := oasOperation.Parameters[oasParamIdx].Value
		paramSchema := oasParam.Schema

		// as always, if we have a reference, we use it as Type
		if paramSchema.Ref != "" {
			gqlParams = append(gqlParams, GqlAttribute{
				Name:       toCamelCase(oasParam.Name),
				Type:       filepath.Base(paramSchema.Ref),
				IsRequired: oasParam.Required,
			})
			continue
		}

		// no reference, go down the anonymous rabbit hole
		typeName, err := anonymousTypeConversion(paramSchema.Value)
		if err != nil {
			return nil, fmt.Errorf("could not convert type in %s: %w", oasParam.Name, err)
		}
		gqlParams = append(gqlParams, GqlAttribute{
			Name:       toCamelCase(oasParam.Name),
			Type:       typeName,
			IsRequired: oasParam.Required,
		})
	}

	return gqlParams, nil
}

// parseResponse returns the best matching returnType as a string
func parseResponse(oasOperation openapi3.Operation) (string, error) {
	// since we can only take one for GraphQL, we have to figure out which is the best
	// Ranking:
	//		1. we absolutely prefer an OpenAPI response named "default"
	// 		2. OpenAPI response named after a http response code and is closest to "200"
	// 		3. Anything really
	bestMatch := ""
	for name := range oasOperation.Responses {
		// "default" is always the best match
		if strings.ToLower(name) == "default" {
			bestMatch = name
			break
		}

		// if nothing is set, we take anything
		if bestMatch == "" {
			bestMatch = name
			continue
		}

		// check for http codes
		code, err := strconv.Atoi(name)
		if err == nil {
			// okay just because it is a number, doesn't mean it's an HTTP Status code
			// make a quick and dirty check
			if !util.IsInSlice(code, util.HttpCodes) {
				// not a http code
				continue
			}
			// now check if the current best match is even an HTTP Status code, if not ours is better
			currentMatchCode, err := strconv.Atoi(name)
			if err != nil || !util.IsInSlice(currentMatchCode, util.HttpCodes) {
				bestMatch = name
				continue
			}

			// okay we are interested in the lowest HTTP OK we can get
			if code >= 200 && code < currentMatchCode {
				bestMatch = name
				continue
			}
		}
	}
	oasResponse := oasOperation.Responses[bestMatch]

	// check for no content
	if len(oasResponse.Value.Content) == 0 {
		return "", nil
	}

	// check for application/json
	parseJsonContent := func(jsonContent *openapi3.MediaType) (string, error) {
		// the schema really should not be nil, but my real-world test set had it sometimes, therefore this is the safety
		if jsonContent.Schema == nil {
			log.Warnf("%s response has no schema in application/json, trying another mime type", bestMatch)
			return "", noSchemaError
		}
		// if it is a named reference, we take it
		if jsonContent.Schema.Ref != "" {
			return filepath.Base(jsonContent.Schema.Ref), nil
		}
		// else it is an anonymous type
		typeName, err := anonymousTypeConversion(jsonContent.Schema.Value)
		if err != nil {
			return "", err
		}
		return typeName, nil
	}
	jsonContent := oasResponse.Value.Content.Get("application/json")
	if jsonContent != nil {
		typeName, err := parseJsonContent(jsonContent)
		if err == nil {
			return typeName, nil
		}
		if !errors.Is(err, noSchemaError) {
			return "", err
		}
		log.Warnf("%s response %s: %s", bestMatch, "application/json", noSchemaError)
	}

	// if we have a simple plain text, we go with string
	if oasResponse.Value.Content.Get("text/plain") != nil {
		return string(gqlString), nil
	}

	// todo check for xml or something else, but I don't think that is in scope rn
	log.Warnf("%s response has no supported content format, defaulting to String", bestMatch)
	return string(gqlString), nil
}

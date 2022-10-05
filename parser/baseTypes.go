package parser

import (
	"fmt"
)

type oasOperationKind string

const (
	oasGet    oasOperationKind = "GET"
	oasPut    oasOperationKind = "PUT"
	oasPost   oasOperationKind = "POST"
	oasDelete oasOperationKind = "DELETE"
)

type oasBaseType string

const (
	oasString oasBaseType = "string"
	oasFloat  oasBaseType = "number"
	oasInt    oasBaseType = "integer"
	oasBool   oasBaseType = "boolean"
)

const gqlDeprecated = "@deprecated"

type gqlBaseType string

const (
	gqlString  gqlBaseType = "String"
	gqlFloat   gqlBaseType = "Float"
	gqlInt     gqlBaseType = "Int"
	gqlBoolean gqlBaseType = "Boolean"
)

func baseTypeConversion(oas oasBaseType) (gqlBaseType, error) {
	if oas == oasString {
		return gqlString, nil
	}
	if oas == oasFloat {
		return gqlFloat, nil
	}
	if oas == oasInt {
		return gqlInt, nil
	}
	if oas == oasBool {
		return gqlBoolean, nil
	}
	return gqlString, fmt.Errorf("could not convert \"%s\" to a gqlBaseTye: not a valid oasBaseType", oas)
}

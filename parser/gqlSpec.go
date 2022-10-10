package parser

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
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

const (
	GqlTypeTemplateName   = "gqlType"
	GqlScalarTemplateName = "gqlScalar"
)

type GqlScalar struct {
	Name string
}

type GqlOperation struct {
	Origin     string
	Name       string
	Parameters []GqlAttribute
	ReturnType string
	Hints      []string
}

type GqlAttribute struct {
	Name       string
	Type       string
	IsRequired bool
	Hints      []string
}

type GqlType struct {
	Name       string
	Type       string
	Attributes []GqlAttribute
}

type GqlSpec struct {
	GenerationTime time.Time
	Types          []GqlType
	Scalars        []GqlScalar
	Mutations      []GqlOperation
	Queries        []GqlOperation
}

//go:embed gqlSchema.tmpl
var gqlSpecTemplate string

func (spec *GqlSpec) String() string {
	templ := template.Must(template.New("gqlSchema").Parse(gqlSpecTemplate))

	spec.GenerationTime = time.Now()
	buf := new(bytes.Buffer)
	err := templ.Execute(buf, spec)
	if err != nil {
		log.Fatal(err)
	}

	return buf.String()
}

var httpCodes = []int{
	http.StatusContinue,
	http.StatusSwitchingProtocols,
	http.StatusProcessing,
	http.StatusEarlyHints,
	http.StatusOK,
	http.StatusCreated,
	http.StatusAccepted,
	http.StatusNonAuthoritativeInfo,
	http.StatusNoContent,
	http.StatusResetContent,
	http.StatusPartialContent,
	http.StatusMultiStatus,
	http.StatusAlreadyReported,
	http.StatusIMUsed,
	http.StatusMultipleChoices,
	http.StatusMovedPermanently,
	http.StatusFound,
	http.StatusSeeOther,
	http.StatusNotModified,
	http.StatusUseProxy,
	http.StatusTemporaryRedirect,
	http.StatusPermanentRedirect,
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusPaymentRequired,
	http.StatusForbidden,
	http.StatusNotFound,
	http.StatusMethodNotAllowed,
	http.StatusNotAcceptable,
	http.StatusProxyAuthRequired,
	http.StatusRequestTimeout,
	http.StatusConflict,
	http.StatusGone,
	http.StatusLengthRequired,
	http.StatusPreconditionFailed,
	http.StatusRequestEntityTooLarge,
	http.StatusRequestURITooLong,
	http.StatusUnsupportedMediaType,
	http.StatusRequestedRangeNotSatisfiable,
	http.StatusExpectationFailed,
	http.StatusTeapot,
	http.StatusMisdirectedRequest,
	http.StatusUnprocessableEntity,
	http.StatusLocked,
	http.StatusFailedDependency,
	http.StatusTooEarly,
	http.StatusUpgradeRequired,
	http.StatusPreconditionRequired,
	http.StatusTooManyRequests,
	http.StatusRequestHeaderFieldsTooLarge,
	http.StatusUnavailableForLegalReasons,
	http.StatusInternalServerError,
	http.StatusNotImplemented,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
	http.StatusHTTPVersionNotSupported,
	http.StatusVariantAlsoNegotiates,
	http.StatusInsufficientStorage,
	http.StatusLoopDetected,
	http.StatusNotExtended,
	http.StatusNetworkAuthenticationRequired,
}

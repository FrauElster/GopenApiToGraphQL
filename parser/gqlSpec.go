package parser

import (
	"bytes"
	_ "embed"
	"html/template"
	"log"
	"time"
)

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

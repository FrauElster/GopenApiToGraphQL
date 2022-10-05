package parser

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"path/filepath"
)

// parseSchema converts OpenAPI schemas to GraphQL types
func parseSchema(doc *openapi3.T) ([]GqlType, error) {
	// we differentiate between
	// 		- "named" types: 	basically all schemas that are explicitly named,
	// 		- anonymous types: 	everything else, where the schema author just put the schema in line

	// now all OpenAPI schemas are per definition named types, here we just map them
	gqlTypes := make([]GqlType, 0, len(doc.Components.Schemas))
	for name, schema := range doc.Components.Schemas {
		gqlType, err := namedTypeConversion(name, schema.Value)
		if err != nil {
			return gqlTypes, fmt.Errorf("could not parse schema: %w", err)
		}
		gqlTypes = append(gqlTypes, gqlType)
	}

	return gqlTypes, nil
}

func namedTypeConversion(name string, schema *openapi3.Schema) (GqlType, error) {
	switch schema.Type {
	case "object":
		// so objects are kind of tricky, because we have to map each property
		attributes := make([]GqlAttribute, 0)
		for propertyName, property := range schema.Properties {
			// so if it is a ref, e.g. '#/components/schema/User', know we will have that component as GraphQL type
			// therefore we can just set the Component Name as type
			if property.Ref != "" {
				attributes = append(attributes, GqlAttribute{Name: propertyName, Type: filepath.Base(property.Ref)})
				continue
			}

			// okay know we have to figure out what type it is, we know it is not a reference to a component, so it is an
			// "anonymous" type. The magic happens in the anonymousTypeConversion
			typeName, err := anonymousTypeConversion(property.Value)
			if err != nil {
				return GqlType{}, err
			}
			attributes = append(attributes, GqlAttribute{Name: propertyName, Type: typeName})
		}

		return GqlType{
			Name:       name,
			Type:       schema.Type,
			Attributes: attributes,
		}, nil
	case "array":
		// actually I don't know if this can even happen, but I am too lazy to check the specs
		var typeName string
		if schema.Items.Ref != "" {
			typeName = filepath.Base(schema.Items.Ref)
		} else {
			var err error
			typeName, err = anonymousTypeConversion(schema.Items.Value)
			if err != nil {
				return GqlType{}, err
			}
		}
		typeName = fmt.Sprintf("[%s]", typeName)

		return GqlType{
			Name:       name,
			Type:       typeName,
			Attributes: []GqlAttribute{},
		}, nil
	default:
		// if it is neither a struct nor an array, it has to be a OpenAPI BaseType, e.g. "number", "integer", or "string"
		// therefore we don't have to recursively check attributes or something like that.
		// all of these BaseTypes have an equivalent GraqhQL BaseType, e.g. number -> Float
		// we will just convert the baseType over to GraphQLs version, and we are done

		// side note, such simple named types are actually GraphQL scalars. We will sort them out at a later point
		typeName, err := baseTypeConversion(oasBaseType(schema.Type))
		if err != nil {
			return GqlType{}, err
		}

		return GqlType{
			Name:       name,
			Type:       string(typeName),
			Attributes: []GqlAttribute{},
		}, nil
	}
}

func anonymousTypeConversion(schema *openapi3.Schema) (string, error) {
	// fixme add hints
	switch schema.Type {
	case "object":
		// again if it is an object, we have to check the types of its properties, ...that screams recursion

		attributes := make([]GqlAttribute, 0)
		for propertyName, property := range schema.Properties {
			// like in namedTypeConversion, if we have a reference to an OpenAPI component,
			// we can just use its name as type
			if property.Ref != "" {
				attributes = append(attributes, GqlAttribute{Name: propertyName, Type: filepath.Base(property.Ref),
					IsRequired: !property.Value.Nullable})
				continue
			}

			// else we will recursively check its type with this function again, someday, somewhere this all has to be a BaseType
			typeName, err := anonymousTypeConversion(property.Value)
			if err != nil {
				return "", err
			}
			attributes = append(attributes, GqlAttribute{
				Name: propertyName, Type: typeName, IsRequired: !property.Value.Nullable})
		}

		// because anonymous types won't be declared explicitly, we map it to a string right here, as if it was a simple
		// "String" type
		result := ""
		for idx := range attributes {
			result += fmt.Sprintf("%s: %s", attributes[idx].Name, attributes[idx].Type)
			if attributes[idx].IsRequired {
				result += "!"
			}
			result += "\n"
		}

		return fmt.Sprintf("{%s}", result), nil
	case "array":
		// again if we have a component reference, we can use it and only have to wrap it with []
		if schema.Items.Ref != "" {
			typeName := filepath.Base(schema.Items.Ref)
			return fmt.Sprintf("[%s]", typeName), nil
		}

		// else we will have to get the type of the items
		typeName, err := anonymousTypeConversion(schema.Items.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("[%s]", typeName), nil

	default:
		// it is a base type! yay, that's a root of the recursion tree :)
		typeName, err := baseTypeConversion(oasBaseType(schema.Type))
		if err != nil {
			return "", err
		}
		return string(typeName), nil
	}
}

package apinext

//go:generate oa2js -o ErrorResponse.json ../../docs/openapi.yaml ErrorResponse
//go:generate go-jsonschema -p $GOPACKAGE --tags json --only-models --output schema.go ErrorResponse.json
//go:generate rm -f ErrorResponse.json

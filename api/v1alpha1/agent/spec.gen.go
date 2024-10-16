// Package v1alpha1 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.3.0 DO NOT EDIT.
package v1alpha1

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	externalRef0 "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/getkin/kin-openapi/openapi3"
)

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"zHThlV9kCxvBzUzk/78BAAD//5sL8uBcbwAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	pathPrefix := path.Dir(pathToFile)

	for rawPath, rawFunc := range externalRef0.PathToRawSpec(path.Join(pathPrefix, "../openapi.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}

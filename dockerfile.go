package dockertesting

import (
	_ "embed"
)

// DockerfileTemplate is the embedded Dockerfile template for building test containers.
//
//go:embed Dockerfile.tmpl
var DockerfileTemplate string

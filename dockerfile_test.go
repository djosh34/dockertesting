package dockertesting

import (
	"strings"
	"testing"
)

func TestDockerfileTemplateEmbedded(t *testing.T) {
	if DockerfileTemplate == "" {
		t.Fatal("DockerfileTemplate is empty, embed failed")
	}
}

func TestDockerfileTemplateContainsGoVersion(t *testing.T) {
	if !strings.Contains(DockerfileTemplate, "ARG GO_VERSION=1.25.6") {
		t.Error("Dockerfile should contain GO_VERSION ARG with default 1.25.6")
	}
}

func TestDockerfileTemplateUsesGolangImage(t *testing.T) {
	if !strings.Contains(DockerfileTemplate, "FROM golang:${GO_VERSION}") {
		t.Error("Dockerfile should use golang:${GO_VERSION} as base image")
	}
}

func TestDockerfileTemplateSetsWorkdir(t *testing.T) {
	if !strings.Contains(DockerfileTemplate, "WORKDIR /app") {
		t.Error("Dockerfile should set WORKDIR to /app")
	}
}

func TestDockerfileTemplateCopiesContext(t *testing.T) {
	if !strings.Contains(DockerfileTemplate, "COPY . .") {
		t.Error("Dockerfile should copy build context")
	}
}

func TestDockerfileTemplateDownloadsDeps(t *testing.T) {
	if !strings.Contains(DockerfileTemplate, "go mod download") {
		t.Error("Dockerfile should run go mod download")
	}
}

func TestDockerfileTemplateEntrypoint(t *testing.T) {
	if !strings.Contains(DockerfileTemplate, "sleep") && !strings.Contains(DockerfileTemplate, "infinity") {
		t.Error("Dockerfile should have sleep infinity entrypoint")
	}
}

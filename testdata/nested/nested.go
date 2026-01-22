// Package nested provides a test fixture that demonstrates spinning up
// nested testcontainers that connect to each other via Docker network aliases.
package nested

// NestedServiceAlias is the DNS alias used for the nested nginx container.
// The test container running this code will make HTTP requests to this alias.
const NestedServiceAlias = "nested-service.test"

// NetworkEnvVar is the environment variable name that contains the Docker network name.
// When running inside dockertesting, this will be set to the network name.
const NetworkEnvVar = "TESTCONTAINERS_DOCKER_NETWORK"

// GetMessage returns a simple message for basic testing.
func GetMessage() string {
	return "Hello from nested package"
}

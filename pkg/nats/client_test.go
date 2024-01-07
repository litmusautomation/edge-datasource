package nats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	// Prepare test data
	opts := ConnectionOptions{
		Token:         "8a30b050-4a63-405b-9bf3-6dba7a44dfc3",
		Host:          "10.17.11.205",
		SkipTLSVerify: true,
	}

	// Call the function
	client, err := NewClient(opts)

	// Verify the result
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

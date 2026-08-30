package spider_test

import (
	"testing"

	"github.com/spider/spider/pkg/auth"
	"github.com/stretchr/testify/require"
)

func TestPasswordHashRoundTrip(t *testing.T) {
	// pbkdf2 120k iterations is CPU/memory heavy — keep this test sequential.
	hashed := auth.HashPassword("spider-admin")
	require.Contains(t, hashed, "pbkdf2_sha256$120000$")
	require.True(t, auth.VerifyPassword("spider-admin", hashed))
	require.False(t, auth.VerifyPassword("wrong", hashed))
}

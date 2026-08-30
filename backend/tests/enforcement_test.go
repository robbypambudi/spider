package spider_test

import (
	"testing"

	"github.com/spider/spider/pkg/apis"
	"github.com/spider/spider/pkg/security/enforcement"
	"github.com/stretchr/testify/require"
)

func TestEnforcerBlockRejects(t *testing.T) {
	e := enforcement.NewEnforcer("closed")
	require.Equal(t, enforcement.ActionForward, e.Resolve(apis.DecisionAllow))
	require.Equal(t, enforcement.ActionReject, e.Resolve(apis.DecisionBlock))
	require.Equal(t, enforcement.ActionHold, e.Resolve(apis.DecisionReview))
}

func TestFailClosedOnError(t *testing.T) {
	e := enforcement.NewEnforcer("closed")
	require.Equal(t, enforcement.ActionReject, e.Resolve(apis.DecisionError))
}

func TestFailOpenOnError(t *testing.T) {
	e := enforcement.NewEnforcer("open")
	require.Equal(t, enforcement.ActionForward, e.Resolve(apis.DecisionError))
}

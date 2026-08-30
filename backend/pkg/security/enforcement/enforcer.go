package enforcement

import "github.com/spider/spider/pkg/apis"

type FailMode string

const (
	FailModeOpen   FailMode = "open"
	FailModeClosed FailMode = "closed"
)

type Action string

const (
	ActionForward Action = "forward"
	ActionReject  Action = "reject"
	ActionHold    Action = "hold"
)

type Enforcer struct {
	FailMode FailMode
}

func NewEnforcer(failMode string) *Enforcer {
	if failMode == "open" {
		return &Enforcer{FailMode: FailModeOpen}
	}
	return &Enforcer{FailMode: FailModeClosed}
}

func (e *Enforcer) Resolve(decision apis.SecurityDecision) Action {
	switch decision {
	case apis.DecisionAllow:
		return ActionForward
	case apis.DecisionBlock:
		return ActionReject
	case apis.DecisionReview:
		return ActionHold
	case apis.DecisionError:
		if e.FailMode == FailModeOpen {
			return ActionForward
		}
		return ActionReject
	default:
		return ActionReject
	}
}

func (e *Enforcer) ShouldForward(decision apis.SecurityDecision) bool {
	return e.Resolve(decision) == ActionForward
}

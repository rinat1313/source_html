package upsertlogic

import "time"

const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionSkipped = "skipped"
)

func DecideAction(existing *time.Time, incoming time.Time) string {
	if existing == nil {
		return ActionCreated
	}
	if incoming.After(*existing) {
		return ActionUpdated
	}
	return ActionSkipped
}

func CountsTowardLimit(action string, alreadyAttachedInRun bool) bool {
	if alreadyAttachedInRun {
		return false
	}
	return action == ActionCreated || action == ActionUpdated
}

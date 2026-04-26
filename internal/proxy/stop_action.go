package proxy

import "strings"

type StopAction string

const (
	StopActionNone    StopAction = ""
	StopActionDelete  StopAction = "delete"
	StopActionDisable StopAction = "disable"
)

var stopActionAliases = map[string]StopAction{
	"delete":  StopActionDelete,
	"del":     StopActionDelete,
	"disable": StopActionDisable,
	"dis":     StopActionDisable,
	"off":     StopActionDisable,
}

func ResolveStopAction(value string) StopAction {
	action, ok := stopActionAliases[strings.ToLower(strings.TrimSpace(value))]
	if ok {
		return action
	}

	return StopActionNone
}

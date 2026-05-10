package ui

import "github.com/charmbracelet/bubbles/key"

func keyMatches(msg interface{ String() string }, binding key.Binding) bool {
	return key.Matches(msg, binding)
}

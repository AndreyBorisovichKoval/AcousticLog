// C:\_Projects_Go\AcousticLog\internal\io\eventkind.go
package io

const (
	EventKindExceeded = "EXCEEDED" // длительное превышение порога
	EventKindImpulse  = "IMPULSE"  // импульсный пик
)

func normalizeEventKind(kind string) string {
	switch kind {
	case EventKindImpulse:
		return EventKindImpulse
	default:
		return EventKindExceeded
	}
}

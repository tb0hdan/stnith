package failsafes

// Failsafe is the interface that all failsafes must implement.
type Failsafe interface {
	Trigger() error
}

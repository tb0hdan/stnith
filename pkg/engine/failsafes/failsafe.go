package failsafes

// FailsafeInterface is the interface that all failsafes must implement.
type FailsafeInterface interface {
	Trigger() error
}

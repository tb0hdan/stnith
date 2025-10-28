package disablers

// DisablerInterface is the interface that all disablers must implement.
type DisablerInterface interface {
	Disable() error
}

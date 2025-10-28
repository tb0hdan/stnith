package disablers

// Disabler is the interface that all disablers must implement.
type Disabler interface {
	Disable() error
}

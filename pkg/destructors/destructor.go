package destructors

type Destructor interface {
	Destroy() error
}


package customerr

type LimitError struct {
	Err string
}

func (e *LimitError) Error() string {
	return e.Err
}

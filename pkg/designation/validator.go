package designation

type DesignationValidator interface {
	New(raw string) (Designation, error)
	Validate(d Designation) (bool, error)
}

// DefaultDesignationValidator is the default implementation of DesignationValidator.
type DefaultDesignationValidator struct{}

func (v *DefaultDesignationValidator) New(raw string) (Designation, error) {
	return New(raw)
}

func (v *DefaultDesignationValidator) Validate(d Designation) (bool, error) {
	return d.Validate()
}

package services

// User registration service
type Register struct {
}

type Registrar interface {
	Regisger() error
}

func NewRegister() *Register {
	return &Register{}
}

func (r *Register) IsRegistred() bool {

	return true
}

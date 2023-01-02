package repository

type Repository interface {
	Get(...interface{}) (interface{}, error)
	GetAll(interface{}) ([]interface{}, error)
	Delete() error
}

package repository

type Repository interface {
	Get(...interface{}) (interface{}, error)
	GetAll(interface{}) ([]interface{}, error)
	Update(...interface{}) (interface{}, error)
	Create(...interface{}) (interface{}, error)
	Delete(interface{}) error
}

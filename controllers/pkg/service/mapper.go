package service

type Mapper interface {
	to(interface{}) interface{}
}

type HAAuditMapper struct{}

func (H HAAuditMapper) from() interface{} {
	//TODO implement me
	panic("implement me")
}

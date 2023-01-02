package service

type Mapper interface {
	from(interface{}) interface{}
	to(interface{}) interface{}
}

type HAAuditMapper struct{}

func (H HAAuditMapper) from() interface{} {
	//TODO implement me
	panic("implement me")
}

func (H HAAuditMapper) to(i interface{}) interface{} {
	//TODO implement me
	panic("implement me")
}

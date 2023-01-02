package service

type Service interface {
	init() error
	run() error
}

package main

type SelfRegisteringResource interface {
	Register(string)
}

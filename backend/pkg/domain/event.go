package domain

type Broker interface {
	EmitAll(*Emit)
}

type Emit struct {
	Topic   string
	Event   string
	Payload interface{}
}

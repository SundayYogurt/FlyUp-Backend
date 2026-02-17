package interfaces

type ConsumerHandler interface {
	HandleMessage(message string) error
}

type ProducerHandler interface {
	PublishMessage(key, value []byte) error
}

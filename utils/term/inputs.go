package term

import (
	"bytes"
)

type logEvent interface {
	isLogEvent() bool
}

var (
	_ logEvent = dynamicLogEvent{}
	_ logEvent = persistentLogEvent{}
	_ logEvent = controlLogEvent{}
)

type dynamicLogEvent struct {
	data  []byte
	input *dynamicInputDefinition
}

func (d dynamicLogEvent) isLogEvent() bool {
	return true
}

type dynamicInputDefinition struct {
	prefix       string
	prefixLength int
}

type dynamicInput struct {
	channel    chan logEvent
	definition *dynamicInputDefinition
}

func (d *dynamicInput) Write(data []byte) (int, error) {
	d.channel <- dynamicLogEvent{
		data:  bytes.Clone(data),
		input: d.definition,
	}
	return len(data), nil
}

type persistentLogEvent struct {
	data     []byte
	isStderr bool
}

func (p persistentLogEvent) isLogEvent() bool {
	return true
}

type persistentInput struct {
	channel  chan logEvent
	isStderr bool
}

func (p *persistentInput) Write(data []byte) (int, error) {
	p.channel <- persistentLogEvent{
		data:     bytes.Clone(data),
		isStderr: p.isStderr,
	}
	return len(data), nil
}

type controlLogEvent struct {
	eventType       string
	responseChannel chan<- struct{}
}

func newClearEvent() (controlLogEvent, <-chan struct{}) {
	channel := make(chan struct{})
	return controlLogEvent{
		eventType:       "clear",
		responseChannel: channel,
	}, channel
}

func newPersistEvent() (controlLogEvent, <-chan struct{}) {
	channel := make(chan struct{})
	return controlLogEvent{
		eventType:       "persist",
		responseChannel: channel,
	}, channel
}

func (c controlLogEvent) isLogEvent() bool {
	return true
}

func (c *controlLogEvent) isClearEvent() bool {
	return c.eventType == "clear"
}

func (c *controlLogEvent) isPersistEvent() bool {
	return c.eventType == "persist"
}

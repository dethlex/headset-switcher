package headset

import (
	"strings"
)

type State int

const (
	SUnknown State = iota
	SListen
	SSpeak
)

func (d State) String() string {
	return [...]string{"", "a2dp_sink", "headset_head_unit"}[d]
}

func ToState(state string) State {
	s, f := map[string]State{
		"a2dp_sink":         SListen,
		"headset_head_unit": SSpeak,
	}[state]

	if !f {
		return SUnknown
	}
	return s
}

type Headset struct {
	state    State
	name     string
	cardName string
	sinkName string
}

func NewHeadset(name, mac string) *Headset {
	return &Headset{
		SUnknown,
		name,
		"bluez_card." + strings.ReplaceAll(mac, ":", "_"),
		"bluez_sink." + strings.ReplaceAll(mac, ":", "_"),
	}
}

func (h *Headset) GetStateName() string {

	return h.name + ": " + [...]string{"Unknown", "Listen", "Speak"}[h.state]
}

func (h *Headset) GetSinkName() string {
	return h.sinkName
}

func (h *Headset) GetCardName() string {
	return h.cardName
}

func (h *Headset) GetState() State {
	return h.state
}

func (h *Headset) SetState(state State) {
	h.state = state
}

type HeadsetMap map[string]*Headset

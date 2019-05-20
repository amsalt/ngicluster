package consts

const (
	// ExtraMsgID for encode&decode extra message.
	ExtraMsgID = 61000
)

// ExtraMsg represents a extra message for relay which stores some information.
type ExtraMsg struct {
	Params map[string]interface{}
}

// system reserved protocol ids is from 60000
const (
	SystemIdentifySelf = 60000
)

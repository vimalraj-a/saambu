package models

// Element is one interactive DOM element discovered by the capture session's
// JS-side inventory walk (a, button, input, textarea, select, [role=button]).
type Element struct {
	Tag      string `bson:"tag" json:"tag"`
	Text     string `bson:"text,omitempty" json:"text,omitempty"`
	Selector string `bson:"selector" json:"selector"`
}

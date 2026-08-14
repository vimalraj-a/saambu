package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ChangeResponseType string

const (
	ChangeResponseClarification ChangeResponseType = "clarification"
	ChangeResponseUpdated       ChangeResponseType = "updated"
)

// ChangeRequest is one turn of the chat-style refinement loop that mediates
// every edit to a TestScript.
type ChangeRequest struct {
	ID            bson.ObjectID      `bson:"_id,omitempty" json:"id"`
	TestScriptID  bson.ObjectID      `bson:"testScriptId" json:"testScriptId"`
	Message       string             `bson:"message" json:"message"`
	ResponseType  ChangeResponseType `bson:"responseType" json:"responseType"`
	ResponseText  string             `bson:"responseText,omitempty" json:"responseText,omitempty"`
	PreviousSteps []string           `bson:"previousSteps,omitempty" json:"previousSteps,omitempty"`
	NewSteps      []string           `bson:"newSteps,omitempty" json:"newSteps,omitempty"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
}

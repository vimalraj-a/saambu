package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ScriptStatus string

const (
	ScriptStatusDraft                ScriptStatus = "draft"
	ScriptStatusVerifying            ScriptStatus = "verifying"
	ScriptStatusResolvedPass         ScriptStatus = "resolved_pass"
	ScriptStatusResolvedKnownFailure ScriptStatus = "resolved_known_failure"
)

// TestScript is the NL spec for one flow. Its Steps are mutated only through
// ReviewChangeRequest — there is no direct-edit path.
type TestScript struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CaptureID bson.ObjectID `bson:"captureId" json:"captureId"`
	Steps     []string      `bson:"steps" json:"steps"`
	Version   int           `bson:"version" json:"version"`
	Status    ScriptStatus  `bson:"status" json:"status"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}

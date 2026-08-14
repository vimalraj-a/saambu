package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// GeneratedTest is the final Playwright artifact produced from a verified
// (or locked-as-known-failure) ExecutionRun.
type GeneratedTest struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	TestScriptID   bson.ObjectID `bson:"testScriptId" json:"testScriptId"`
	SourceRunID    bson.ObjectID `bson:"sourceRunId,omitempty" json:"sourceRunId,omitempty"`
	Code           string        `bson:"code" json:"code"`
	ExpectedToFail bool          `bson:"expectedToFail" json:"expectedToFail"`
	CreatedAt      time.Time     `bson:"createdAt" json:"createdAt"`
}

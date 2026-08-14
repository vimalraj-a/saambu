package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ExecutionRun is one attempt at driving the browser through a TestScript
// version. Only the test-step portion is stored here — the prerequisite
// replay that precedes it is re-run from the Capture's PrerequisiteSteps but
// not persisted per-run (it's already visible on the Capture itself).
type ExecutionRun struct {
	ID            bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	TestScriptID  bson.ObjectID   `bson:"testScriptId" json:"testScriptId"`
	ScriptVersion int             `bson:"scriptVersion" json:"scriptVersion"`
	Steps         []StepExecution `bson:"steps" json:"steps"`
	OverallPassed bool            `bson:"overallPassed" json:"overallPassed"`
	CreatedAt     time.Time       `bson:"createdAt" json:"createdAt"`
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Capture is the starting state for one flow, reached after running the
// prerequisite steps (if any) starting from URL. Screenshot/HTML/Elements
// reflect the page the session landed on after that run, not the raw URL.
type Capture struct {
	ID                bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	URL               string          `bson:"url" json:"url"`
	Title             string          `bson:"title,omitempty" json:"title,omitempty"`
	PrerequisiteText  string          `bson:"prerequisiteText,omitempty" json:"prerequisiteText,omitempty"`
	PrerequisiteSteps []string        `bson:"prerequisiteSteps,omitempty" json:"prerequisiteSteps,omitempty"`
	PrerequisiteRun   []StepExecution `bson:"prerequisiteRun,omitempty" json:"prerequisiteRun,omitempty"`
	ScreenshotBase64  string          `bson:"screenshotBase64" json:"screenshotBase64"`
	HTMLSnapshot      string          `bson:"htmlSnapshot,omitempty" json:"htmlSnapshot,omitempty"`
	Elements          []Element       `bson:"elements,omitempty" json:"elements,omitempty"`
	CreatedAt         time.Time       `bson:"createdAt" json:"createdAt"`
}

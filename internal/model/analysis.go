package model

import (
	"github.com/edulinq/autograder/internal/timestamp"
)

// A key for pairwise analysis.
// Should always be an ordered (lexicographically) pair of full submissions IDs.
type PairwiseKey [2]string

const PAIRWISE_KEY_DELIM string = "||"

type IndividualAnalysis struct {
	AnalysisTimestamp timestamp.Timestamp `json:"analysis-timestamp"`

	SubmissionStartTime timestamp.Timestamp `json:"submission-start-time"`
	Files               []string            `json:"files"`

	LinesOfCode int     `json:"lines-of-code"`
	Score       float64 `json:"score"`

	LinesOfCodeDelta float64 `json:"lines-of-code-delta"`
	ScoreDelta       float64 `json:"score-delta"`

	LinesOfCodeVelocity float64 `json:"lines-of-code-velocity"`
	ScoreVelocity       float64 `json:"score-velocity"`
}

type PairWiseAnalysis struct {
	AnalysisTimestamp timestamp.Timestamp `json:"analysis-timestamp"`
	SubmissionIDs     PairwiseKey         `json:"submission-ids"`

	Similarities   map[string][]*FileSimilarity `json:"similarities"`
	UnmatchedFiles [][2]string                  `json:"unmatched-files"`
}

type FileSimilarity struct {
	Filename string         `json:"filename"`
	Tool     string         `json:"tool"`
	Version  string         `json:"version"`
	Options  map[string]any `json:"options,omitempty"`
	Score    float64        `json:"score"`
}

func NewPairwiseKey(fullSubmissionID1 string, fullSubmissionID2 string) PairwiseKey {
	return PairwiseKey([2]string{
		min(fullSubmissionID1, fullSubmissionID2),
		max(fullSubmissionID1, fullSubmissionID2),
	})
}

func (this *PairwiseKey) String() string {
	return this[0] + PAIRWISE_KEY_DELIM + this[1]
}
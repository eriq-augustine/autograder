package main

import (
	"fmt"
	"testing"

	"github.com/edulinq/autograder/internal/cmd"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	cmd.CMDServerTestingMain(suite)
}

func TestPeekBase(test *testing.T) {
	testCases := []struct {
		cmd.CommonCMDTestCases
		targetEmail      string
		courseID         string
		assignmentID     string
		targetSubmission string
	}{
		{
			CommonCMDTestCases: cmd.CommonCMDTestCases{
				ExpectedStdout: SUBMISSION_1697406272,
			},
			targetEmail:  "course-student@test.edulinq.org",
			courseID:     "course101",
			assignmentID: "hw0",
		},
		{
			CommonCMDTestCases: cmd.CommonCMDTestCases{
				ExpectedStdout: SUBMISSION_1697406272,
			},
			targetEmail:      "course-student@test.edulinq.org",
			courseID:         "course101",
			assignmentID:     "hw0",
			targetSubmission: "1697406272",
		},
		{
			CommonCMDTestCases: cmd.CommonCMDTestCases{
				ExpectedStdout: SUBMISSION_1697406272,
			},
			targetEmail:      "course-student@test.edulinq.org",
			courseID:         "course101",
			assignmentID:     "hw0",
			targetSubmission: "course101::hw0::student@test.com::1697406272",
		},
		{
			CommonCMDTestCases: cmd.CommonCMDTestCases{
				ExpectedStdout: NO_SUBMISSION,
			},
			targetEmail:  "course-admin@test.edulinq.org",
			courseID:     "course101",
			assignmentID: "hw0",
		},
		{
			CommonCMDTestCases: cmd.CommonCMDTestCases{
				ExpectedStdout: INCORRECT_SUBMISSION,
			},
			targetEmail:      "course-student@test.edulinq.org",
			courseID:         "course101",
			assignmentID:     "hw0",
			targetSubmission: "ZZZ",
		},
		{
			CommonCMDTestCases: cmd.CommonCMDTestCases{
				ExpectedExitCode:        2,
				ExpectedStderrSubstring: `"Could not find course: 'ZZZ'."`,
			},
			targetEmail:  "course-student@test.edulinq.org",
			courseID:     "ZZZ",
			assignmentID: "hw0",
		},
		{
			CommonCMDTestCases: cmd.CommonCMDTestCases{
				ExpectedExitCode:        2,
				ExpectedStderrSubstring: `"Could not find assignment: 'zzz'."`,
			},
			targetEmail:  "course-student@test.edulinq.org",
			courseID:     "course101",
			assignmentID: "zzz",
		},
	}

	for i, testCase := range testCases {
		args := []string{
			testCase.targetEmail,
			testCase.courseID,
			testCase.assignmentID,
			testCase.targetSubmission,
		}

		cmd.RunCommonCMDTests(test, main, args, testCase.CommonCMDTestCases, fmt.Sprintf("Case %d: ", i))
	}
}

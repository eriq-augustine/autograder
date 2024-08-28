package lms

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	lmstest "github.com/edulinq/autograder/internal/lms/backend/test"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestUploadScores(test *testing.T) {
	defer db.ResetForTesting()
	// Reset the LMS adapter.
	defer func() {
		lmstest.SetFailUpdateAssignmentScores(false)
	}()

	testCases := []struct {
		email      string
		permError  bool
		failUpdate bool
		addUsers   map[string]*model.CourseUser
		scores     []ScoreEntry
		expected   *UploadScoresResponse
	}{
		// Normal.
		{
			"course-grader", false, false, nil,
			[]ScoreEntry{
				ScoreEntry{"course-student@test.edulinq.org", 10},
			},
			&UploadScoresResponse{
				Count:             1,
				ErrorCount:        0,
				UnrecognizedUsers: []RowEntry{},
				NoLMSIDUsers:      []RowEntry{},
			},
		},
		{
			"course-grader", false, false, nil,
			[]ScoreEntry{
				ScoreEntry{"course-student@test.edulinq.org", 10},
				ScoreEntry{"course-grader@test.edulinq.org", 0},
				ScoreEntry{"course-admin@test.edulinq.org", -10},
				ScoreEntry{"course-owner@test.edulinq.org", 12.34},
			},
			&UploadScoresResponse{
				Count:             4,
				ErrorCount:        0,
				UnrecognizedUsers: []RowEntry{},
				NoLMSIDUsers:      []RowEntry{},
			},
		},

		// Permissions.
		{"course-other", true, false, nil, nil, nil},
		{"course-student", true, false, nil, nil, nil},

		// Upload fails.
		{
			"course-grader", false, true, nil,
			[]ScoreEntry{
				ScoreEntry{"course-student@test.edulinq.org", 10},
			},
			nil,
		},

		// Bad scores.
		{
			"course-grader", false, false,
			map[string]*model.CourseUser{
				"no-lms-id@test.edulinq.org": &model.CourseUser{"no-lms-id@test.edulinq.org", nil, model.CourseRoleStudent, nil},
			},
			[]ScoreEntry{
				ScoreEntry{"zzz@test.edulinq.org", 10},
				ScoreEntry{"no-lms-id@test.edulinq.org", 20},
				ScoreEntry{"abc@test.edulinq.org", 30},
				ScoreEntry{"course-student@test.edulinq.org", 10},
			},
			&UploadScoresResponse{
				Count:      1,
				ErrorCount: 3,
				UnrecognizedUsers: []RowEntry{
					RowEntry{0, "zzz@test.edulinq.org"},
					RowEntry{2, "abc@test.edulinq.org"},
				},
				NoLMSIDUsers: []RowEntry{
					RowEntry{1, "no-lms-id@test.edulinq.org"},
				},
			},
		},

		// Upload will pass, but never gets called.
		{
			"course-grader", false, false, nil,
			[]ScoreEntry{
				ScoreEntry{"zzz@test.edulinq.org", 10},
			},
			&UploadScoresResponse{
				Count:      0,
				ErrorCount: 1,
				UnrecognizedUsers: []RowEntry{
					RowEntry{0, "zzz@test.edulinq.org"},
				},
				NoLMSIDUsers: []RowEntry{},
			},
		},

		// Upload will fail, but never gets called.
		{
			"course-grader", false, true, nil,
			[]ScoreEntry{
				ScoreEntry{"zzz@test.edulinq.org", 10},
			},
			&UploadScoresResponse{
				Count:      0,
				ErrorCount: 1,
				UnrecognizedUsers: []RowEntry{
					RowEntry{0, "zzz@test.edulinq.org"},
				},
				NoLMSIDUsers: []RowEntry{},
			},
		},
	}

	TEST_COURSE_ID := "course-with-lms"

	for i, testCase := range testCases {
		db.ResetForTesting()

		if testCase.addUsers != nil {
			db.UpsertCourseUsers(db.MustGetCourse(TEST_COURSE_ID), testCase.addUsers)
		}

		fields := map[string]any{
			"course-id": TEST_COURSE_ID,
			// ID does not matter, test LMS will accept all ids.
			"assignment-lms-id": "foo",
			"scores":            testCase.scores,
		}

		lmstest.SetFailUpdateAssignmentScores(testCase.failUpdate)

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`lms/upload/scores`), fields, nil, testCase.email)
		if !response.Success {
			expectedLocator := ""
			if testCase.permError {
				expectedLocator = "-020"
			} else if testCase.failUpdate {
				expectedLocator = "-406"
			}

			if expectedLocator == "" {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			} else {
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, expectedLocator, response.Locator)
				}
			}

			continue
		}

		var responseContent UploadScoresResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(testCase.expected, &responseContent) {
			test.Errorf("Case %d: Unexpected result. Expected: '%+v', actual: '%+v'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent))
			continue
		}
	}
}

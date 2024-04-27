package grader

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/edulinq/autograder/common"
	"github.com/edulinq/autograder/config"
	"github.com/edulinq/autograder/db"
	"github.com/edulinq/autograder/model"
)

var SUBMISSION_RELPATH string = filepath.Join("test-submissions", "solution");

func TestRejectSubmissionMaxAttempts(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    assignment := db.MustGetTestAssignment();

    // Set the max submissions to zero.
    maxValue := 0
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{Max: &maxValue};

    // Make a submission that should be rejected.
    submitForRejection(test, assignment, "other@test.com", false, &RejectMaxAttempts{0});
}

func TestRejectSubmissionMaxAttemptsInfinite(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    assignment := db.MustGetTestAssignment();

    // Set the max submissions to empty (infinite).
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{};

    // All submissions should pass.
    submitForRejection(test, assignment, "other@test.com", false, nil);

    // Set the max submissions to nagative (infinite).
    maxValue := -1
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{Max: &maxValue};

    // All submissions should pass.
    submitForRejection(test, assignment, "other@test.com", false, nil);
}

func TestRejectSubmissionMaxWindowAttempts(test *testing.T) {
    testMaxWindowAttemps(test, "other@test.com", true);
}

func TestRejectSubmissionMaxWindowAttemptsAdmin(test *testing.T) {
    testMaxWindowAttemps(test, "grader@test.com", false);
}

func testMaxWindowAttemps(test *testing.T, user string, expectReject bool) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    assignment := db.MustGetTestAssignment();
    duration := common.DurationSpec{Days: 1000};

    // Set the submission limit window to 1 attempt in a large duration.
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{
        Window: &model.SubmittionLimitWindow{
            AllowedAttempts: 1,
            Duration: duration,
        },
    };

    // Make a submission that should pass.
    result, _, _ := submitForRejection(test, assignment, user, false, nil);

    expectedTime, err := result.Info.GradingStartTime.Time();
    if (err != nil) {
        test.Fatalf("Failed to parse expected time: '%v'.", err);
    }

    // Make a submission that should be rejected.
    var reason RejectReason;
    if (expectReject) {
        reason = &RejectWindowMax{1, duration, expectedTime};
    }

    submitForRejection(test, assignment, user, false, reason);
}


func TestRejectSubmissionLateAcknowledgmentOverdueEmptyPolicy(test *testing.T) {
    // if late policy is not set, can submit overdue assignment without late acknowledgment

    db.ResetForTesting();
    defer db.ResetForTesting();

    assignment := db.MustGetTestAssignment();
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{};
    assignment.LatePolicy = &model.LateGradingPolicy{Type: model.EmptyPolicy}

    assignment.DueDate = common.TimestampFromTime(time.Now().AddDate(-1, 0, 0)) // was due a year ago
    submitForRejection(test, assignment, "other@test.com", false, nil)
}

func TestRejectSubmissionLateAcknowledgmentOverdue(test *testing.T) {
    // late submission is rejected without acknowledgment

    db.ResetForTesting();
    defer db.ResetForTesting();

    assignment := db.MustGetTestAssignment();
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{};
    assignment.LatePolicy = &model.LateGradingPolicy{Type: model.LateDays}

    assignment.DueDate = common.TimestampFromTime(time.Now().AddDate(-1, 0, 0))
    submitForRejection(test, assignment, "other@test.com", false, &RejectMissingLateAcknowledgment{})
}

func TestRejectSubmissionLateAcknowledgmentOverdueWithAck(test *testing.T) {
    // late submission is accepted with acknowledgment

    db.ResetForTesting();
    defer db.ResetForTesting();

    assignment := db.MustGetTestAssignment();
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{};
    assignment.LatePolicy = &model.LateGradingPolicy{Type: model.LateDays}

    assignment.DueDate = common.TimestampFromTime(time.Now().AddDate(-1, 0, 0))
    submitForRejection(test, assignment, "other@test.com", true, nil)
}

func TestRejectSubmissionLateAcknowledgmentNotOverdue(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    assignment := db.MustGetTestAssignment();
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{};
    assignment.LatePolicy = &model.LateGradingPolicy{Type: model.LateDays}

    assignment.DueDate = common.TimestampFromTime(time.Now().AddDate(+1, 0, 0))
    submitForRejection(test, assignment, "other@test.com", false, nil) // assignment is not overdue so can submit without acknowledgment
}

func submitForRejection(test *testing.T, assignment *model.Assignment, user string, lateAcknowledgment bool, expectedRejection RejectReason) (
        *model.GradingResult, RejectReason, error) {
    // Disable testing mode to check for rejection.
    config.TESTING_MODE.Set(false);
    defer config.TESTING_MODE.Set(true);

    submissionPath := filepath.Join(assignment.GetSourceDir(), SUBMISSION_RELPATH);

    if assignment.SubmissionLimit != nil {
        err := assignment.SubmissionLimit.Validate();
        if (err != nil) {
            test.Fatalf("Failed to validate submission limit: '%v'.", err);
        }
    }

    result, reject, err := GradeDefault(assignment, submissionPath, user, TEST_MESSAGE, lateAcknowledgment);
    if (err != nil) {
        test.Fatalf("Failed to grade assignment: '%v'.", err);
    }

    if (expectedRejection == nil) {
        // Submission should go through.

        if (reject != nil) {
            test.Fatalf("Submission was rejected: '%s'.", reject.String());
        }

        if (result == nil) {
            test.Fatalf("Did not get a grading result.");
        }
    } else {
        // Submission should be rejected.

        if (result != nil) {
            test.Fatalf("Should not get a grading result.");
        }

        if (reject == nil) {
            test.Fatalf("Submission was not rejected when it should have been.");
        }

        if (!reflect.DeepEqual(expectedRejection, reject)) {
            test.Fatalf("Did not get the expected rejection. Expected: '%+v', Actual: '%+v'.", expectedRejection, reject);
        }
    }

    return result, reject, err;
}

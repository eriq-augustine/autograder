package submission

import (
    "path/filepath"
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/util"
    "github.com/eriq-augustine/autograder/model"
)

var SUBMISSION_RELPATH string = filepath.Join("test-submissions", "solution", "submission.py");

func TestSubmit(test *testing.T) {
    testSubmissions, err := grader.GetTestSubmissions(config.COURSES_ROOT.Get());
    if (err != nil) {
        test.Fatalf("Failed to get test submissions in '%s': '%v'.", config.COURSES_ROOT.Get(), err);
    }

    for i, testSubmission := range testSubmissions {
        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/submit`), nil, testSubmission.Files, model.RoleStudent);
        if (!response.Success) {
            test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            continue;
        }

        var responseContent SubmitResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (!responseContent.GradingSucess) {
            test.Errorf("Case %d: Response is not a grading success when it should be: '%v'.", i, responseContent);
            continue;
        }

        if (responseContent.Rejected) {
            test.Errorf("Case %d: Response is rejected when it should not be: '%v'.", i, responseContent);
            continue;
        }

        if (responseContent.RejectReason != "") {
            test.Errorf("Case %d: Response has a reject reason when it should not: '%v'.", i, responseContent);
            continue;
        }

        if (!responseContent.GradingInfo.Equals(*testSubmission.TestSubmission.GradingInfo, !testSubmission.TestSubmission.IgnoreMessages)) {
            test.Errorf("Case %d: Actual output:\n---\n%v\n---\ndoes not match expected output:\n---\n%v\n---\n.",
                    i, responseContent.GradingInfo, testSubmission.TestSubmission.GradingInfo);
            continue;
        }

        // Fetch the most recent submission from the DB and ensure it matches.
        submission, err := db.GetSubmissionResult(testSubmission.Assignment, "student@test.com", "");
        if (err != nil) {
            test.Errorf("Case %d: Failed to get submission: '%v'.", i, err);
            continue;
        }

        if (!responseContent.GradingInfo.Equals(*submission, !testSubmission.TestSubmission.IgnoreMessages)) {
            test.Errorf("Case %d: Actual output:\n---\n%v\n---\ndoes not match database value:\n---\n%v\n---\n.",
                    i, responseContent.GradingInfo, submission);
            continue;
        }
    }
}

func TestRejectSubmissionMaxAttempts(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    // Note that we are using a submission from a different assignment.
    assignment := db.MustGetTestAssignment();
    paths := []string{filepath.Join(assignment.GetSourceDir(), SUBMISSION_RELPATH)};

    fields := map[string]any{
        "course-id": "course101-with-zero-limit",
        "assignment-id": "hw0",
    };

    response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/submit`), fields, paths, model.RoleStudent);
    if (!response.Success) {
        test.Fatalf("Response is not a success when it should be: '%v'.", response);
    }

    var responseContent SubmitResponse;
    util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

    if (responseContent.GradingSucess) {
        test.Fatalf("Response is a grading success when it should not be: '%v'.", responseContent);
    }

    if (!responseContent.Rejected) {
        test.Fatalf("Response is not rejected when it should be: '%v'.", responseContent);
    }

    if (responseContent.RejectReason == "") {
        test.Fatalf("Response does not have a reject reason when it should: '%v'.", responseContent);
    }

    expected := (&grader.RejectMaxAttempts{0}).String();
    if (expected != responseContent.RejectReason) {
        test.Fatalf("Did not get the expected rejection reason. Expected: '%s', Actual: '%s'.",
            expected, responseContent.RejectReason);
    }
}

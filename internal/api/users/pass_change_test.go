package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestPassChange(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		newPass  string
		expected PassChangeResponse
	}{
		{"spooky", PassChangeResponse{true, false}},
		{"admin", PassChangeResponse{true, true}},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		fields := map[string]any{
			"user-email": "admin@test.com",
			"user-pass":  util.Sha256HexFromString("admin"),
			"new-pass":   util.Sha256HexFromString(testCase.newPass),
		}

		response := core.SendTestAPIRequest(test, core.NewEndpoint(`users/pass/change`), fields)
		if !response.Success {
			test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			continue
		}

		var responseContent PassChangeResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !reflect.DeepEqual(testCase.expected, responseContent) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent))
			continue
		}

		user, err := db.GetServerUser("admin@test.com", true)
		if err != nil {
			test.Errorf("Case %d: Failed to get saved user: '%v'.", i, err)
			continue
		}

		success, err := user.Auth(util.Sha256HexFromString(testCase.newPass))
		if err != nil {
			test.Errorf("Case %d: Failed to auth user: '%v'.", i, err)
			continue
		}

		if !success {
			test.Errorf("Case %d: The new password fails to auth after the change.", i)
			continue
		}
	}
}

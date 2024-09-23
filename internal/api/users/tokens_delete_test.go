package users

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestTokensDelete(test *testing.T) {
	db.ResetForTesting()
	db.MustClearTestUserTokens()
	defer db.ResetForTesting()

	email := "course-admin@test.edulinq.org"
	user := db.MustGetServerUser(email)

	// Add a token.

	_, _, err := user.CreateRandomToken("test", model.TokenSourceServer)
	if err != nil {
		test.Fatalf("Failed to create token: '%v'.", err)
	}

	err = db.UpsertUser(user)
	if err != nil {
		test.Fatalf("Could not upsert user: '%v'.", err)
	}

	// Re-fetch and ensure the token exists.
	user = db.MustGetServerUser(email)

	initialTokenCount := len(user.Tokens)
	if initialTokenCount == 0 {
		test.Fatalf("Test user has no tokens.")
	}

	args := map[string]any{
		"token-id": user.Tokens[0].ID,
	}

	response := core.SendTestAPIRequest(test, core.NewEndpoint("users/tokens/delete"), args)
	if !response.Success {
		test.Fatalf("Response not successful: '%s'.", util.MustToJSONIndent(response))
	}

	var responseContent TokensDeleteResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	if !responseContent.Found {
		test.Fatalf("Could not find token to delete.")
	}

	user = db.MustGetServerUser(email)

	newTokenCount := len(user.Tokens)

	if newTokenCount != (initialTokenCount - 1) {
		test.Fatalf("Incorrect token count. Expected: %d, Found: %d.", (initialTokenCount - 1), newTokenCount)
	}
}

func TestTokensDeleteNoTokens(test *testing.T) {
	db.ResetForTesting()
	db.MustClearTestUserTokens()
	defer db.ResetForTesting()

	email := "course-admin@test.edulinq.org"
	user := db.MustGetServerUser(email)

	if len(user.Tokens) != 0 {
		test.Fatalf("Test user has tokens.")
	}

	args := map[string]any{
		"token-id": "abc123",
	}

	response := core.SendTestAPIRequest(test, core.NewEndpoint("users/tokens/delete"), args)
	if !response.Success {
		test.Fatalf("Response not successful: '%s'.", util.MustToJSONIndent(response))
	}

	var responseContent TokensDeleteResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	if responseContent.Found {
		test.Fatalf("Found token to delete (when there should not be one).")
	}

	user = db.MustGetServerUser(email)

	if len(user.Tokens) != 0 {
		test.Fatalf("User somehow gained tokens...")
	}
}

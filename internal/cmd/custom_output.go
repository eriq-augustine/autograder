package cmd

import (
	"strings"

	"github.com/edulinq/autograder/internal/api/core"
	courseUsers "github.com/edulinq/autograder/internal/api/courses/users"
	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/util"
)

var EndpointCustomFormatters = map[string]CustomResponseFormatter{
	"users/list":         mustListServerUsersTable,
	"courses/users/list": mustListCourseUsersTable,
}

func mustListCourseUsersTable(response core.APIResponse) string {
	var responseContent courseUsers.ListResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	var courseUsersTable strings.Builder

	headers := []string{"email", "name", "role", "lms-id"}
	courseUsersTable.WriteString(strings.Join(headers, "\t") + "\n")

	for i, user := range responseContent.Users {
		if i > 0 {
			courseUsersTable.WriteString("\n")
		}

		courseUsersTable.WriteString(user.Email)
		courseUsersTable.WriteString("\t")
		courseUsersTable.WriteString(user.Name)
		courseUsersTable.WriteString("\t")
		courseUsersTable.WriteString(user.Role.String())
		courseUsersTable.WriteString("\t")
		courseUsersTable.WriteString(user.LMSID)
	}

	return courseUsersTable.String()
}

func mustListServerUsersTable(response core.APIResponse) string {
	var responseContent users.ListResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	var serverUsersTable strings.Builder

	headers := []string{"email", "name", "server-role", "courses"}
	serverUsersTable.WriteString(strings.Join(headers, "\t") + "\n")

	for i, user := range responseContent.Users {
		if i > 0 {
			serverUsersTable.WriteString("\n")
		}

		serverUsersTable.WriteString(user.Email)
		serverUsersTable.WriteString("\t")
		serverUsersTable.WriteString(user.Name)
		serverUsersTable.WriteString("\t")
		serverUsersTable.WriteString(user.Role.String())
		serverUsersTable.WriteString("\t")
		serverUsersTable.WriteString(util.MustToJSON(user.Courses))
	}

	return serverUsersTable.String()
}

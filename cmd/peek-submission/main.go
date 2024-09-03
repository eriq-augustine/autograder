package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/assignments/submissions"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	TargetEmail      string `help:"Email of the user to fetch." arg:""`
	CourseID         string `help:"ID of the course." arg:""`
	AssignmentID     string `help:"ID of the assignment." arg:""`
	TargetSubmission string `help:"ID of the submission. Defaults to latest submission." arg:"" optional:""`
	Verbose          bool   `help:"Print the entire response." short:"v"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Fetch a submission for a specific assignment and user."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
	}

	socketPath, err := common.GetUnixSocketPath()
	if err != nil {
		log.Fatal("Failed to get the unix socket path.", err)
	}

	connection, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatal("Failed to dial the unix socket.", err)
	}
	defer connection.Close()

	request := submissions.FetchUserPeekRequest{
		APIRequestAssignmentContext: core.APIRequestAssignmentContext{
			APIRequestCourseUserContext: core.APIRequestCourseUserContext{
				CourseID: args.CourseID,
			},
			AssignmentID: args.AssignmentID,
		},
		TargetUser: core.TargetCourseUserSelfOrGrader{
			TargetCourseUser: core.TargetCourseUser{
				Email: args.TargetEmail,
			},
		},
		TargetSubmission: args.TargetSubmission,
	}

	requestMap := map[string]interface{}{
		api.ENDPOINT_KEY: core.NewEndpoint(`courses/assignments/submissions/fetch/user/peek`),
		api.REQUEST_KEY:  request,
	}

	jsonRequest := util.MustToJSONIndent(requestMap)
	jsonBytes := []byte(jsonRequest)
	err = util.WriteToUnixSocket(connection, jsonBytes)
	if err != nil {
		log.Fatal("Failed to write the request to the unix socket.", err)
	}

	responseBuffer, err := util.ReadFromUnixSocket(connection)
	if err != nil {
		log.Fatal("Failed to read the response from the unix socket.", err)
	}

	var response core.APIResponse
	err = json.Unmarshal(responseBuffer, &response)
	if err != nil {
		log.Fatal("Failed to unmarshal the API response.", err)
	}

	if !response.Success {
		message := "Request to the autograder failed."
		if response.Message != "" {
			message = fmt.Sprintf("Failed to complete operation: %s", response.Message)
		}

		log.Error("API request was not successful.", log.NewAttr("message", message), log.NewAttr("http-status", response.HTTPStatus))
	}

	if args.Verbose {
		fmt.Println(util.MustToJSONIndent(response))
	} else {
		var responseContent submissions.FetchUserPeekResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
		fmt.Println(util.MustToJSONIndent(responseContent))
	}
}
package db

import (
	"errors"
	"fmt"

	"github.com/edulinq/autograder/internal/model"
)

// See Backend.
func GetServerUsers() (map[string]*model.ServerUser, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetServerUsers()
}

// See Backend.
func GetCourseUsers(course *model.Course) (map[string]*model.CourseUser, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetCourseUsers(course)
}

// See Backend.
func GetServerUser(email string, includeTokens bool) (*model.ServerUser, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetServerUser(email, includeTokens)
}

// See Backend.
func UpsertUsers(users map[string]*model.ServerUser) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.UpsertUsers(users)
}

// Get a specific course user.
// Returns nil if the user does not exist or is not enrolled in the course.
func GetCourseUser(course *model.Course, email string) (*model.CourseUser, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	serverUser, err := backend.GetServerUser(email, false)
	if err != nil {
		return nil, err
	}

	if serverUser == nil {
		return nil, nil
	}

	return serverUser.GetCourseUser(course.ID)
}

// Convenience function for UpsertUsers() with a single user.
func UpsertUser(user *model.ServerUser) error {
	users := map[string]*model.ServerUser{user.Email: user}
	return UpsertUsers(users)
}

// Convenience function for UpsertUsers() with course users.
func UpsertCourseUsers(course *model.Course, users map[string]*model.CourseUser) error {
	serverUsers := make(map[string]*model.ServerUser, len(users))

	var userErrors error = nil
	for email, user := range users {
		serverUser, err := user.GetServerUser(course.ID)
		if err != nil {
			userErrors = errors.Join(userErrors, fmt.Errorf("Invalid user '%s': '%w'.", email, err))
		} else {
			serverUsers[email] = serverUser
		}
	}

	if userErrors != nil {
		return fmt.Errorf("Found errors when processing users: '%w'.", userErrors)
	}

	return UpsertUsers(serverUsers)
}

// Convenience function for UpsertCourseUsers() with a single user.
func UpsertCourseUser(course *model.Course, user *model.CourseUser) error {
	users := map[string]*model.CourseUser{user.Email: user}
	return UpsertCourseUsers(course, users)
}

// Delete a user from the server.
// Returns a boolean indicating if the user exists.
// If true, then the user exists and was removed.
// If false (and the error is nil), then the user did not exist.
func DeleteUser(email string) (bool, error) {
	if backend == nil {
		return false, fmt.Errorf("Database has not been opened.")
	}

	user, err := GetServerUser(email, false)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	return true, backend.DeleteUser(email)
}

// Remove a user from the course (but leave on the server).
// Returns booleans indicating if the user exists and was enrolled in the course.
func RemoveUserFromCourse(course *model.Course, email string) (bool, bool, error) {
	if backend == nil {
		return false, false, fmt.Errorf("Database has not been opened.")
	}

	user, err := GetServerUser(email, false)
	if err != nil {
		return false, false, err
	}

	if user == nil {
		return false, false, nil
	}

	_, exists := user.Roles[course.ID]
	if !exists {
		return true, false, nil
	}

	return true, true, backend.RemoveUserFromCourse(course, email)
}
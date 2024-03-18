package main

import (
    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/email"
    "github.com/edulinq/autograder/log"
)

var args struct {
    config.ConfigArgs
    To []string `help:"Email recipents." required:""`
    Subject string `help:"Email subject." required:""`
    Body string `help:"Email body." required:""`
    Course string `help:"Course ID." default:""`
}

func main() {
    kong.Parse(&args,
        kong.Description("Send an email."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    course, err := db.GetCourse(args.Course);
    if (err != nil) {
        log.Fatal("Failed to get course: '%s', '%w'.", args.Course, err);
    }

    emailTo, err := db.ResolveUsers(course, args.To);
    if (err != nil) {
        log.Fatal("Failed to resolve users: '%s', '%w'.", course.GetName(), err);
    }

    err = email.Send(emailTo, args.Subject, args.Body, false);
    if (err != nil) {
        log.Fatal("Could not send email.", err);
    }
}

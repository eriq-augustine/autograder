package common

// Note that this file is largely a copy of db/test.go.
// The content is repeated to avoid an import cycle.

import (
	"os"
	"testing"

	"github.com/eriq-augustine/autograder/config"
	"github.com/eriq-augustine/autograder/log"
	"github.com/eriq-augustine/autograder/util"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
    // Run inside a func so defers will run before os.Exit().
    code := func() int {
        config.MustEnableUnitTestingMode();
		log.SetLevelFatal();
		
        defer CleanupTestingMain();

        return suite.Run();
    }();

    os.Exit(code);
}

func CleanupTestingMain() {
    // Remove any temp directories.
    err := util.RemoveRecordedTempDirs();
    if (err != nil) {
        log.Error("Error when removing temp dirs.", err);
    }
}

// Copyright 2019 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package mtail_test

import (
	"os"
	"path/filepath"
	"testing"

	"flashcat.cloud/categraf/inputs/mtail/internal/mtail"
	"flashcat.cloud/categraf/inputs/mtail/internal/testutil"
)

func TestPollLogPathPatterns(t *testing.T) {
	testutil.SkipIfShort(t)
	tmpDir := testutil.TestTempDir(t)

	logDir := filepath.Join(tmpDir, "logs")
	testutil.FatalIfErr(t, os.Mkdir(logDir, 0o700))
	testutil.Chdir(t, logDir)

	m, stopM := mtail.TestStartServer(t, 0, mtail.LogPathPatterns(logDir+"/files/*/log/*log"))
	defer stopM()

	logCountCheck := m.ExpectExpvarDeltaWithDeadline("log_count", 1)
	lineCountCheck := m.ExpectExpvarDeltaWithDeadline("lines_total", 1)

	logFile := filepath.Join(logDir, "files", "a", "log", "a.log")
	testutil.FatalIfErr(t, os.MkdirAll(filepath.Dir(logFile), 0o700))

	f := testutil.TestOpenFile(t, logFile)
	defer f.Close()
	m.PollWatched(1)

	logCountCheck()

	testutil.WriteString(t, f, "line 1\n")
	m.PollWatched(1)
	lineCountCheck()
}

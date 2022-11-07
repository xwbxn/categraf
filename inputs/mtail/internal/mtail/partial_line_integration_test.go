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

func TestPartialLineRead(t *testing.T) {
	testutil.SkipIfShort(t)

	tmpDir := testutil.TestTempDir(t)

	logDir := filepath.Join(tmpDir, "logs")
	progDir := filepath.Join(tmpDir, "progs")
	err := os.Mkdir(logDir, 0o700)
	testutil.FatalIfErr(t, err)
	err = os.Mkdir(progDir, 0o700)
	testutil.FatalIfErr(t, err)

	logFile := filepath.Join(logDir, "log")

	f := testutil.TestOpenFile(t, logFile)
	defer f.Close()

	m, stopM := mtail.TestStartServer(t, 1, mtail.ProgramPath(progDir), mtail.LogPathPatterns(logDir+"/log"))
	defer stopM()

	lineCountCheck := m.ExpectExpvarDeltaWithDeadline("lines_total", 2)

	testutil.WriteString(t, f, "line 1\n")
	m.PollWatched(1)

	testutil.WriteString(t, f, "line ")
	m.PollWatched(1)

	testutil.WriteString(t, f, "2\n")
	m.PollWatched(1)

	lineCountCheck()
}

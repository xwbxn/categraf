// Copyright 2019 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package mtail_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"flashcat.cloud/categraf/inputs/mtail/internal/mtail"
	"flashcat.cloud/categraf/inputs/mtail/internal/testutil"
)

func TestMultipleLinesInOneWrite(t *testing.T) {
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

	m.PollWatched(1) // Force sync to EOF

	{
		lineCountCheck := m.ExpectExpvarDeltaWithDeadline("lines_total", 1)
		n, err := f.WriteString("line 1\n")
		testutil.FatalIfErr(t, err)
		log.Printf("Wrote %d bytes", n)
		m.PollWatched(1)
		lineCountCheck()
	}

	{
		lineCountCheck := m.ExpectExpvarDeltaWithDeadline("lines_total", 2)
		n, err := f.WriteString("line 2\nline 3\n")
		testutil.FatalIfErr(t, err)
		log.Printf("Wrote %d bytes", n)
		m.PollWatched(1)
		lineCountCheck()
	}
}

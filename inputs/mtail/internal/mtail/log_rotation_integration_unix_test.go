// Copyright 2019 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

//go:build unix
// +build unix

package mtail_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"flashcat.cloud/categraf/inputs/mtail/internal/mtail"
	"flashcat.cloud/categraf/inputs/mtail/internal/testutil"
)

// TestLogRotation is a unix-specific test because on Windows, files cannot be removed
// or renamed while there is an open read handle on them. Instead, log rotation would
// have to be implemented by copying and then truncating the original file. That test
// case is already covered by TestLogTruncation.
func TestLogRotation(t *testing.T) {
	testutil.SkipIfShort(t)

	for _, tc := range []bool{false, true} {
		tc := tc
		name := "disabled"
		if tc {
			name = "enabled"
		}
		t.Run(fmt.Sprintf("race simulation %s", name), func(t *testing.T) {
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

			logOpensTotalCheck := m.ExpectMapExpvarDeltaWithDeadline("log_opens_total", logFile, 1)
			logLinesTotalCheck := m.ExpectMapExpvarDeltaWithDeadline("log_lines_total", logFile, 3)

			testutil.WriteString(t, f, "line 1\n")
			m.PollWatched(1)

			testutil.WriteString(t, f, "line 2\n")
			m.PollWatched(1)

			logClosedCheck := m.ExpectMapExpvarDeltaWithDeadline("log_closes_total", logFile, 1)
			logCompletedCheck := m.ExpectExpvarDeltaWithDeadline("log_count", -1)
			log.Println("rename")
			err = os.Rename(logFile, logFile+".1")
			testutil.FatalIfErr(t, err)
			if tc {
				m.PollWatched(0)    // simulate race condition with this poll.
				logClosedCheck()    // sync when filestream closes fd
				m.PollWatched(0)    // invoke the GC
				logCompletedCheck() // sync to when the logstream is removed from tailer
			}
			log.Println("create")
			f = testutil.TestOpenFile(t, logFile)
			m.PollWatched(1)
			testutil.WriteString(t, f, "line 1\n")
			m.PollWatched(1)

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				logLinesTotalCheck()
			}()
			go func() {
				defer wg.Done()

				logOpensTotalCheck()
			}()
			wg.Wait()
		})
	}
}

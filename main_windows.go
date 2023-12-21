//go:build windows

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/chai2010/winsvc"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	"flashcat.cloud/categraf/agent"
	"flashcat.cloud/categraf/config"
	"flashcat.cloud/categraf/pkg/pprof"
)

var (
	pprofStart          uint32
	flagWinSvcName      = flag.String("win-service-name", "categraf", "Set windows service name")
	flagWinSvcDesc      = flag.String("win-service-desc", "Categraf", "Set windows service description")
	flagWinSvcInstall   = flag.Bool("win-service-install", false, "Install windows service")
	flagWinSvcUninstall = flag.Bool("win-service-uninstall", false, "Uninstall windows service")
	flagWinSvcStart     = flag.Bool("win-service-start", false, "Start windows service")
	flagWinSvcStop      = flag.Bool("win-service-stop", false, "Stop windows service")
)

func runAgent(ag *agent.Agent) {
	if !winsvc.IsAnInteractiveSession() {
		initLog("categraf.log")

		if err := winsvc.RunAsService(*flagWinSvcName, ag.Start, ag.Stop, false); err != nil {
			log.Fatalln("F! failed to run windows service:", err)
		}
		return
	}

	ag.Start()
	go profile()
	handleSignal(ag)
}

func doOSsvc() {
	// install service
	if *flagWinSvcInstall {
		if err := installService(appPath, *flagWinSvcName, *flagWinSvcDesc); err != nil {
			log.Fatalln("F! failed to install service:", *flagWinSvcName, "error:", err)
		}
		fmt.Println("done")
		os.Exit(0)
	}

	// uninstall service
	if *flagWinSvcUninstall {
		if err := winsvc.RemoveService(*flagWinSvcName); err != nil {
			log.Fatalln("F! failed to uninstall service:", *flagWinSvcName, "error:", err)
		}
		fmt.Println("done")
		os.Exit(0)
	}

	// start service
	if *flagWinSvcStart {
		if err := winsvc.StartService(*flagWinSvcName); err != nil {
			log.Fatalln("F! failed to start service:", *flagWinSvcName, "error:", err)
		}
		fmt.Println("done")
		os.Exit(0)
	}

	// stop service
	if *flagWinSvcStop && runtime.GOOS == "windows" {
		if err := winsvc.StopService(*flagWinSvcName); err != nil {
			log.Fatalln("F! failed to stop service:", *flagWinSvcName, "error:", err)
		}
		fmt.Println("done")
		os.Exit(0)
	}
}

func profile() {
	// TODO: replace with windows event
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			file := filepath.Join(config.Config.ConfigDir, ".pprof")
			if _, err := os.Stat(file); err == nil {
				if !atomic.CompareAndSwapUint32(&pprofStart, 0, 1) {
					return
				}
				go pprof.Go()
			}
		}
	}
}

func installService(appPath string, name string, desc string, params ...string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err == nil {
		return fmt.Errorf("winsvc.InstallService: service %s already exists", name)
	}
	defer s.Close()

	serviceType := windows.SERVICE_WIN32_OWN_PROCESS
	s, err = m.CreateService(name, appPath,
		mgr.Config{
			DisplayName:      name,
			Description:      desc,
			StartType:        mgr.StartAutomatic,
			ServiceType:      uint32(serviceType),
			DelayedAutoStart: true,
		},
		params...,
	)
	if err != nil {
		return err
	}

	if err := s.SetRecoveryActions([]mgr.RecoveryAction{
		{
			Type:  mgr.ServiceRestart,
			Delay: 0,
		},
	}, 0); err != nil {
		return err
	}

	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("winsvc.InstallService: InstallAsEventCreate failed, err = %v", err)
	}
	return nil
}

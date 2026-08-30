package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spider/spider/pkg/config"
	"github.com/spider/spider/pkg/workerctl"
)

func workerCommand(settings *config.Settings, base string, userToken *string) *cobra.Command {
	cmd := &cobra.Command{Use: "worker", Short: "Manage SPIDER cluster workers"}
	cmd.AddCommand(workerJoinCommand(settings))
	cmd.AddCommand(workerStopCommand())
	cmd.AddCommand(workerPsCommand())
	cmd.AddCommand(workerListCommand(base, userToken))
	return cmd
}

func workerJoinCommand(settings *config.Settings) *cobra.Command {
	var (
		apiURL    string
		workerTok string
		id        string
		site      string
		interval  int
		detach    bool
		logFile   string
	)

	cmd := &cobra.Command{
		Use:   "join",
		Short: "Register this machine as a worker in the SPIDER cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedID := workerctl.ResolveWorkerID(id)

			if detach {
				childArgs := []string{
					"worker", "join",
					"--api", apiURL,
					"--worker-token", workerTok,
					"--id", resolvedID,
					"--heartbeat-interval", strconv.Itoa(interval),
				}
				if site != "" {
					childArgs = append(childArgs, "--site", site)
				}
				pid, err := workerctl.Detach(resolvedID, childArgs, logFile)
				if err != nil {
					return fmt.Errorf("start background worker: %w", err)
				}
				logPath := logFile
				if logPath == "" {
					logPath, _ = workerctl.DefaultLogFilePath(resolvedID)
				}
				fmt.Printf("worker %s started in background (pid %d)\n", resolvedID, pid)
				fmt.Printf("logs: %s\n", logPath)
				fmt.Printf("stop with: spider worker stop %s\n", resolvedID)
				return nil
			}

			if err := workerctl.WritePIDFile(resolvedID, os.Getpid()); err != nil {
				return fmt.Errorf("write pid file: %w", err)
			}
			defer func() { _ = workerctl.RemovePIDFile(resolvedID) }()

			runSettings := *settings
			runSettings.APIBaseURL = apiURL
			runSettings.WorkerToken = workerTok
			runSettings.WorkerHeartbeatInterval = interval

			var sitePtr *string
			if site != "" {
				sitePtr = &site
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			fmt.Printf("joining %s as worker %q (ctrl+c to leave)\n", apiURL, resolvedID)
			return workerctl.Run(ctx, workerctl.RunOptions{Settings: &runSettings, WorkerID: resolvedID, Site: sitePtr})
		},
	}

	cmd.Flags().StringVar(&apiURL, "api", settings.APIBaseURL, "SPIDER API base URL")
	cmd.Flags().StringVar(&workerTok, "worker-token", settings.WorkerToken, "Worker auth token (X-Spider-Worker-Token)")
	cmd.Flags().StringVar(&id, "id", "", "Worker ID override (default: auto-generated / .spider-worker-id)")
	cmd.Flags().StringVar(&site, "site", "", "Optional site/location label")
	cmd.Flags().IntVar(&interval, "heartbeat-interval", settings.WorkerHeartbeatInterval, "Heartbeat interval in seconds")
	cmd.Flags().BoolVarP(&detach, "detach", "d", false, "Run the worker in the background instead of blocking this terminal")
	cmd.Flags().StringVar(&logFile, "log-file", "", "Log file for --detach (default: ~/.spider/worker/<id>.log)")
	return cmd
}

func workerStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <worker-id>",
		Short: "Stop a worker started locally with `spider worker join`",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := workerctl.Stop(args[0]); err != nil {
				return err
			}
			fmt.Printf("worker %s stopped\n", args[0])
			return nil
		},
	}
}

func workerPsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ps",
		Short: "List workers started locally via `spider worker join`",
		RunE: func(cmd *cobra.Command, args []string) error {
			workers, err := workerctl.ListLocal()
			if err != nil {
				return err
			}
			if len(workers) == 0 {
				fmt.Println("no local workers")
				return nil
			}
			for _, w := range workers {
				status := "stopped"
				if w.Running {
					status = "running"
				}
				fmt.Printf("%s\tpid=%d\t%s\t%s\n", w.WorkerID, w.PID, status, w.LogFile)
			}
			return nil
		},
	}
}

func workerListCommand(base string, userToken *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List workers registered with the cluster (requires --token)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return apiGet(base+"/api/v1/workers", *userToken, os.Stdout)
		},
	}
}

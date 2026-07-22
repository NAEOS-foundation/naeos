package main

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/api"
	"github.com/NAEOS-foundation/naeos/internal/dashboard"
	ws "github.com/NAEOS-foundation/naeos/internal/websocket"
)

var (
	dashPort string
)

func newDashboardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Start NAEOS web dashboard",
		Long:  `Start the NAEOS web dashboard for monitoring and managing projects.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dash, err := dashboard.New()
			if err != nil {
				return err
			}

			wsServer := ws.NewServer()
			wsServer.SetAllowedOrigins([]string{"*"})
			go wsServer.Run()

			mux := api.NewServer(dashPort, &api.AuthConfig{Enabled: false})
			mux.SetWebSocketServer(wsServer)
			mux.Router.HandleFunc("/", dash.ServeHTTP)
			mux.Router.HandleFunc("/ws", wsServer.HandleWebSocket)

			al := dashboard.NewActivityLog(500)
			ch := dashboard.NewComponentHealth()
			dh := dashboard.NewAPIHandler(dash, al, ch, dashboard.DefaultConfig())
			mux.Router.Handle("/api/stats", dh)
			mux.Router.Handle("/api/activity", dh)
			mux.Router.Handle("/api/health", dh)

			broadcaster := ws.NewEventBroadcaster(wsServer)
			al.SetLogCallback(func(entry dashboard.LogEntry) {
				level := string(entry.Level)
				if level == "" {
					level = "info"
				}
				broadcaster.LogMessage(level, entry.Message)
			})

			observer := ws.NewWSObserver(broadcaster)
			mux.SetPipelineObserver(observer)

			go func() {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()
				for range ticker.C {
					stats := dashboard.GetStats()
					wsServer.Broadcast("stats_update", stats)
				}
			}()

			go func() {
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()
				for range ticker.C {
					al.Add(dashboard.LevelInfo, "Dashboard running")
				}
			}()

			ch.Set("API Server", dashboard.Healthy, "Running")
			ch.Set("Parser", dashboard.Healthy, "Ready")
			ch.Set("Compiler", dashboard.Healthy, "Ready")
			ch.Set("MCP Server", dashboard.Degraded, "Stopped")

			return mux.Start()
		},
	}

	cmd.Flags().StringVarP(&dashPort, "port", "p", "3000", "Dashboard port")

	return cmd
}

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/websocket"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

var (
	wsPort       string
	wsConfigPath string
	wsInput      string
	wsInputFile  string
)

func newWebSocketCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ws",
		Short: "Start WebSocket server for real-time updates",
		Long:  `Start WebSocket server for real-time dashboard updates and event streaming.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			server := websocket.NewServer()
			go server.Run()

			broadcaster := websocket.NewEventBroadcaster(server)
			observer := websocket.NewWSObserver(broadcaster)

			http.HandleFunc("/ws", server.HandleWebSocket)
			http.HandleFunc("/ws/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"status":"healthy","clients":%d}`, server.ClientCount())
			})

			if wsInput != "" || wsInputFile != "" {
				inputValue, err := loadInput(wsInput, wsInputFile)
				if err != nil {
					return err
				}

				cfg, err := loadPipelineConfig(wsConfigPath, cliVerbose, nil, cliDryRun)
				if err != nil {
					return err
				}
				cfg.Observer = observer

				p, err := pipeline.New(*cfg)
				if err != nil {
					return err
				}

				broadcaster.LogMessage("info", "pipeline started")
				result, err := p.Run(inputValue)
				if err != nil {
					broadcaster.LogMessage("error", fmt.Sprintf("pipeline failed: %v", err))
					return err
				}
				broadcaster.LogMessage("info", fmt.Sprintf("pipeline complete: %d artifacts", len(result.Artifacts)))
			}

			fmt.Printf("WebSocket server starting on ws://localhost%s/ws\n", wsPort)
			srv := &http.Server{
				Addr:              wsPort,
				ReadHeaderTimeout: 10 * time.Second,
				ReadTimeout:       15 * time.Second,
				WriteTimeout:      15 * time.Second,
				IdleTimeout:       60 * time.Second,
			}
			return srv.ListenAndServe()
		},
	}

	cmd.Flags().StringVarP(&wsPort, "port", "p", ":8081", "WebSocket server port")
	cmd.Flags().StringVar(&wsConfigPath, "config", "", "path to config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&wsInput, "input", "", "specification input to process")
	cmd.Flags().StringVar(&wsInputFile, "input-file", "", "path to a specification file")

	return cmd
}

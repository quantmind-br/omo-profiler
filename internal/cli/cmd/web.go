package cmd

import (
	"github.com/diogenes/omo-profiler/internal/web"
	"github.com/spf13/cobra"
)

var (
	webHost   string
	webPort   int
	webNoOpen bool
)

var WebCmd = &cobra.Command{
	Use:   "web",
	Short: "Launch the web UI for managing profiles",
	Long:  "Starts a local web server (default http://127.0.0.1:4747) with a browser UI for managing profiles in ~/.omo/omo.json.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return web.Serve(web.Options{Host: webHost, Port: webPort, Open: !webNoOpen})
	},
}

func init() {
	WebCmd.Flags().StringVar(&webHost, "host", "127.0.0.1", "Host to bind")
	WebCmd.Flags().IntVar(&webPort, "port", 4747, "Port to listen on")
	WebCmd.Flags().BoolVar(&webNoOpen, "no-open", false, "Do not open the browser automatically")
}

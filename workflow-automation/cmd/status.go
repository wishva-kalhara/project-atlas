package cmd

import (
    "fmt"
    "os"
    "text/tabwriter"

    "github.com/spf13/cobra"
    "workflow-automation/internal/config"
    "workflow-automation/internal/data"
)

var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show current scheduling state for Facebook reels",
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, err := config.Load()
        if err != nil {
            return err
        }
        repo := data.NewRepository(cfg.FbReelsDataPath)
        state, err := repo.Load()
        if err != nil {
            return err
        }
        // Pretty table output
        w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
        fmt.Fprintln(w, "SLUG\tPAGE\tCAPTION\tVIDEO PATH\tTRIED\tSUCCESS")
        for _, p := range state.Posts {
            fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\t%t\n",
                state.FbPageSlug,
                state.PageName,
                p.Caption,
                p.VideoPath,
                p.SchedulingInfo.Tried,
                p.SchedulingInfo.IsSuccess,
            )
        }
        w.Flush()
        return nil
    },
}

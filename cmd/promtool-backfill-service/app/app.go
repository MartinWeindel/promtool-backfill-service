package app

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type promtoolBackfillService struct {
}

func NewCommand(log logrus.FieldLogger) *cobra.Command {
	app := &promtoolBackfillService{}
	cmd = &cobra.Command{
		Use:   "promtool-backfill-service",
		Short: "Prometheus promtool backfill service",
		RunE:  app.run,
	}

	return cmd
}

func (svc *promtoolBackfillService) run(cmd *cobra.Command, args []string) error {
	return nil
}

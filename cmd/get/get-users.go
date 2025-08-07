package get

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/user"
	"github.com/spf13/cobra"
)

type GetTopicsFlags struct {
	OutputFormat string
}

func newGetUsersCmd() *cobra.Command {

	var flags user.GetUsersFlags

	var cmdGetUsers = &cobra.Command{
		Use:   "users",
		Short: "list available users",
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&user.Operation{}).GetUsers(flags)
		},
	}

	cmdGetUsers.Flags().StringVarP(&flags.OutputFormat, "output", "o", flags.OutputFormat, "output format. One of: json|yaml|wide|compact")

	return cmdGetUsers
}

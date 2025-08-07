package create

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/k8s"
	"github.com/deviceinsight/kafkactl/v5/internal/user"
	"github.com/spf13/cobra"
)

func newCreateUserCmd() *cobra.Command {

	var flags user.CreateUsersFlags

	var cmdCreateUser = &cobra.Command{
		Use:   "user USER",
		Short: "create a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if internal.IsKubernetesEnabled() {
				return k8s.NewOperation().Run(cmd, args)
			}
			return (&user.Operation{}).CreateUsers(args[0], flags)
		},
	}

	cmdCreateUser.Flags().Int32VarP(&flags.Iterations, "iterations", "i", 4096, "number of iterations")
	cmdCreateUser.Flags().StringVarP(&flags.Password, "password", "p", "", "password")
	cmdCreateUser.Flags().StringVarP(&flags.ScramMechanism, "mechanism", "m", "SCRAM-SHA-512", "scram mechanism (SCRAM-SHA-256, SCRAM-SHA-512)")

	return cmdCreateUser
}

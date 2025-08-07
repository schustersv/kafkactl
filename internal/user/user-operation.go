package user

import (
	"sort"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type Operation struct {
}

type CreateUsersFlags struct {
	Password       string
	ScramMechanism string
	Iterations     int32
}

type DeleteUsersFlags struct {
	ScramMechanism string
}

type GetUsersFlags struct {
	OutputFormat string
}

type User struct {
	Name           string
	ScramMechanism string `json:"scramMechanism" yaml:"scramMechanism"`
	Iterations     int32  `json:"iterations" yaml:"iterations"`
}

func (operation *Operation) GetUsers(flags GetUsersFlags) error {
	var (
		err     error
		context internal.ClientContext
		admin   sarama.ClusterAdmin
		users   []*sarama.DescribeUserScramCredentialsResult
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	if users, err = admin.DescribeUserScramCredentials([]string{}); err != nil {
		return errors.Wrap(err, "failed to read topics")
	}

	tableWriter := output.CreateTableWriter()

	switch flags.OutputFormat {
	case "json", "yaml":

	case "", "wide":
		if err := tableWriter.WriteHeader("USER", "SCRAM MECHANISM", "ITERATIONS"); err != nil {
			return err
		}
	case "compact":
		tableWriter.Initialize()
	default:
		return errors.Errorf("unknown outputFormat: %s", flags.OutputFormat)
	}

	userList := make([]User, 0, len(users))
	for _, u := range users {
		c := u.CredentialInfos
		for _, scram := range c {
			user := User{
				Name:           u.User,
				ScramMechanism: scram.Mechanism.String(),
				Iterations:     scram.Iterations,
			}
			userList = append(userList, user)
		}
	}

	sort.Slice(userList, func(i, j int) bool {
		return userList[i].Name < userList[j].Name
	})

	switch flags.OutputFormat {
	case "json", "yaml":
		return output.PrintObject(userList, flags.OutputFormat)
	case "compact":
		for _, u := range userList {
			if err := tableWriter.Write(u.Name); err != nil {
				return err
			}
		}
	default:
		for _, u := range userList {
			if err := tableWriter.Write(u.Name, u.ScramMechanism, strconv.Itoa(int(u.Iterations))); err != nil {
				return err
			}
		}
	}

	if flags.OutputFormat == "wide" || flags.OutputFormat == "compact" || flags.OutputFormat == "" {
		if err := tableWriter.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func (operation *Operation) ListUserNames() ([]string, error) {
	var (
		err     error
		context internal.ClientContext
		admin   sarama.ClusterAdmin
		users   []*sarama.DescribeUserScramCredentialsResult
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return nil, err
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return nil, errors.Wrap(err, "failed to create cluster admin")
	}

	if users, err = admin.DescribeUserScramCredentials([]string{}); err != nil {
		return nil, errors.Wrap(err, "failed to read topics")
	}

	userList := make([]string, 0, len(users))
	for _, u := range users {
		userList = append(userList, u.User)
	}

	return userList, nil
}

func (operation *Operation) CreateUsers(user string, flags CreateUsersFlags) error {

	var (
		err            error
		context        internal.ClientContext
		admin          sarama.ClusterAdmin
		newUser        sarama.AlterUserScramCredentialsUpsert
		scramMechanism sarama.ScramMechanismType
		result         []*sarama.AlterUserScramCredentialsResult
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	if flags.Iterations < 1 {
		return errors.Errorf("iterations must be greater than 0")
	}

	switch flags.ScramMechanism {
	case "SCRAM-SHA-256":
		scramMechanism = sarama.SCRAM_MECHANISM_SHA_256
	case "SCRAM-SHA-512":
		scramMechanism = sarama.SCRAM_MECHANISM_SHA_512
	default:
		return errors.Errorf("unknown scram mechanism: %s", flags.ScramMechanism)
	}

	if flags.Password == "" {
		return errors.Errorf("password must not be empty")
	}

	newUser = sarama.AlterUserScramCredentialsUpsert{
		Name:       user,
		Mechanism:  scramMechanism,
		Iterations: flags.Iterations,
		Password:   []byte(flags.Password),
	}
	if result, err = admin.UpsertUserScramCredentials([]sarama.AlterUserScramCredentialsUpsert{newUser}); err != nil {
		return errors.Wrapf(err, "failed to create user %s", user)
	}

	for _, r := range result {
		if r.ErrorCode != sarama.ErrNoError {
			return errors.Errorf("user creation resulted in error for %s: %s", r.User, r.ErrorCode.Error())
		}
	}

	output.Infof("user created: %s", user)

	return nil
}

func (operation *Operation) DeleteUser(user string, flags CreateUsersFlags) error {

	var (
		err            error
		context        internal.ClientContext
		admin          sarama.ClusterAdmin
		scramMechanism sarama.ScramMechanismType
		result         []*sarama.AlterUserScramCredentialsResult
		deleteUser     sarama.AlterUserScramCredentialsDelete
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	switch flags.ScramMechanism {
	case "SCRAM-SHA-256":
		scramMechanism = sarama.SCRAM_MECHANISM_SHA_256
	case "SCRAM-SHA-512":
		scramMechanism = sarama.SCRAM_MECHANISM_SHA_512
	default:
		return errors.Errorf("unknown scram mechanism: %s", flags.ScramMechanism)
	}

	deleteUser = sarama.AlterUserScramCredentialsDelete{
		Name:      user,
		Mechanism: scramMechanism,
	}
	if result, err = admin.DeleteUserScramCredentials([]sarama.AlterUserScramCredentialsDelete{deleteUser}); err != nil {
		return errors.Wrapf(err, "failed to delete user %s", user)
	}

	for _, r := range result {
		if r.ErrorCode != sarama.ErrNoError {
			return errors.Errorf("user deletion resulted in error for %s: %s", r.User, r.ErrorCode.Error())
		}
	}

	output.Infof("user deleted: %s", user)

	return nil
}

func CompleteUserNames(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {

	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	users, err := (&Operation{}).ListUserNames()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return users, cobra.ShellCompDirectiveNoFileComp
}

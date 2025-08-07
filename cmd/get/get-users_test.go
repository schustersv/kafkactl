package get_test

import (
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestGetUsersIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	// Create test users
	testutil.CreateUser(t, "test-user-a", "password123", "SCRAM-SHA-256", 4096)
	testutil.CreateUser(t, "test-user-b", "password456", "SCRAM-SHA-512", 8192)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("get", "users"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := kafkaCtl.GetStdOutLines()

	testutil.AssertContains(t, "USER|SCRAM MECHANISM|ITERATIONS", outputLines)
	testutil.AssertContains(t, "test-user-a|SCRAM-SHA-256|4096", outputLines)
	testutil.AssertContains(t, "test-user-b|SCRAM-SHA-512|8192", outputLines)
}

package database_test

import (
	"encoder/framework/database"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDbTest(t *testing.T) {
	database := database.NewDbTest()

	require.NotNil(t, database)
}

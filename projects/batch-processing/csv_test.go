package main

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadInputCSV(t *testing.T) {
	t.Run("with an invalid csv filepath", func(t *testing.T) {
		inputFilepath := "invalid-path/invalid.csv"

		inputCSVRows, err := ReadInputCSV(inputFilepath)

		expectedError := errors.New("🔴 error: failed to open the input csv file")
		require.Error(t, expectedError, err)
		require.Contains(t, err.Error(), expectedError.Error())

		require.Nil(t, inputCSVRows)
	})

	t.Run("with a valid csv filepath but an empty csv file", func(t *testing.T) {
		temporaryDirectory := t.TempDir()

		file, err := os.CreateTemp(temporaryDirectory, "empty-*.csv")
		require.NoError(t, err)
		defer file.Close()

		inputCSVRows, err := ReadInputCSV(file.Name())

		expectedError := errors.New("🔴 error: the input csv file is empty")
		require.Error(t, err)
		require.Contains(t, err.Error(), expectedError.Error())

		require.Nil(t, inputCSVRows)
	})

	t.Run("with a valid csv filepath but invalid csv rows", func(t *testing.T) {
		temporaryDirectory := t.TempDir()

		file, err := os.CreateTemp(temporaryDirectory, "invalid-*.csv")
		require.NoError(t, err)
		defer file.Close()

		_, err = file.WriteString("test-heading-1,test-heading-2\nrow-1-value-1,row-1-value-2,row-1-value-3\nrow-2-value-1,row-2-value-2\n")
		require.NoError(t, err)

		inputCSVRows, err := ReadInputCSV(file.Name())
		expectedError := errors.New("🔴 error: failed to read all the input csv rows")
		require.Error(t, err)
		require.Contains(t, err.Error(), expectedError.Error())

		require.Nil(t, inputCSVRows)
	})

	t.Run("with a valid csv filepath and valid csv rows", func(t *testing.T) {
		temporaryDirectory := t.TempDir()

		file, err := os.CreateTemp(temporaryDirectory, "valid-*.csv")
		require.NoError(t, err)
		defer file.Close()

		_, err = file.WriteString("test-heading-1,test-heading-2\nrow-1-value-1,row-1-value-2\nrow-2-value-1,row-2-value-2\n")
		require.NoError(t, err)

		inputCSVRows, err := ReadInputCSV(file.Name())
		require.NoError(t, err)

		expectedCSVRows := [][]string{
			{"test-heading-1", "test-heading-2"},
			{"row-1-value-1", "row-1-value-2"},
			{"row-2-value-1", "row-2-value-2"},
		}
		require.Equal(t, expectedCSVRows, inputCSVRows)
	})
}

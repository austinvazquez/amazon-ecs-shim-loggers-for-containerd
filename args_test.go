// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// +build unit

package main

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

const (
	testContainerID   = "test-container-id"
	testContainerName = "test-container-name"
	testLogDriver     = "test-log-driver"
)

// TestGetGlobalArgs tests getGlobalArgs with/without correct settings of
// flag values.
func TestGetGlobalArgs(t *testing.T) {
	t.Run("NoError", testGetGlobalArgsNoError)
	t.Run("WithError", testGetGlobalArgsWithError)
}

// testGetGlobalArgsNoError is a sub-test of TestGetGlobalArgs. It tests
// getGlobalArgs function with no returned errors
func testGetGlobalArgsNoError(t *testing.T) {
	// Unset all keys used for this test
	defer viper.Reset()

	viper.Set(containerIDKey, testContainerID)
	viper.Set(containerNameKey, testContainerName)
	viper.Set(logDriverTypeKey, testLogDriver)

	args, err := getGlobalArgs()
	require.NoError(t, err)
	assert.Equal(t, args.ContainerID, testContainerID)
	assert.Equal(t, args.ContainerName, testContainerName)
	assert.Equal(t, args.LogDriver, testLogDriver)
	assert.Equal(t, args.Mode, blockingMode)
	assert.Equal(t, args.MaxBufferSize, 0)
	assert.Equal(t, *args.CleanupTime, time.Duration(5*time.Second))
}

// testGetGlobalArgsWithError is a sub-test of TestGetGlobalArgs. It tests
// getGlobalArgs function with returned errors when required flag values are
// missed. Since the logic in getGlobalArgs is examining required flag values
// one by one in this order:
//     container-id, container-name, log-driver,
// in this test, we will set flag one by one in the previous order to examine if
// each value check worked as expected.
func testGetGlobalArgsWithError(t *testing.T) {
	// Unset all keys used for this test
	defer viper.Reset()

	testCasesWithError := []struct {
		keyToSet string
		valToSet string
	}{
		{containerIDKey, testContainerID},
		{containerNameKey, testContainerName},
		{logDriverTypeKey, testLogDriver},
	}

	for _, tc := range testCasesWithError {
		_, err := getGlobalArgs()
		require.Error(t, err)
		require.Contains(t,
			err.Error(),
			fmt.Sprintf("%s is required", tc.keyToSet),
		)
		viper.Set(tc.keyToSet, tc.valToSet)
	}
}

// TestGetModeAndMaxBufferSize tests getModeAndMaxBufferSize with/without correct
// settings of mode.
func TestGetModeAndMaxBufferSize(t *testing.T) {
	t.Run("NoError", testGetModeAndMaxBufferSizeNoError)
	t.Run("WithError", testGetModeAndMaxBufferSizeWithError)
}

// testGetModeAndMaxBufferSizeNoError is a sub-test of TestGetModeAndMaxBufferSize.
// It tests different valid values set for mode option.
func testGetModeAndMaxBufferSizeNoError(t *testing.T) {
	// Unset all keys used for this test
	defer viper.Reset()

	testCasesNoError := []struct {
		mode               string
		expectedMode       string
		expectedBufferSize int
	}{
		{"", blockingMode, 0},
		{blockingMode, blockingMode, 0},
		{nonBlockingMode, nonBlockingMode, int(math.Pow(2, 20))},
	}

	for _, tc := range testCasesNoError {
		viper.Set(modeKey, tc.mode)
		mode, maxBufferSize, err := getModeAndMaxBufferSize()
		require.NoError(t, err)
		require.Equal(t, tc.expectedMode, mode)
		require.Equal(t, tc.expectedBufferSize, maxBufferSize)
	}
}

// testGetModeAndMaxBufferSizeNoError is a sub-test of TestGetModeAndMaxBufferSize. It
// tests invalid values set for mode option.
func testGetModeAndMaxBufferSizeWithError(t *testing.T) {
	// Unset all keys used for this test
	defer viper.Reset()

	viper.Set(modeKey, "test-mode")
	_, _, err := getModeAndMaxBufferSize()
	require.Error(t, err)
}

// TestGetMaxBufferSize tests getMaxBufferSize with/without valid setting max buffer
// size options.
func TestGetMaxBufferSize(t *testing.T) {
	t.Run("NoError", testGetMaxBufferSizeNoError)
	t.Run("WithError", testGetMaxBufferSizeWithError)
}

// testGetMaxBufferSizeNoError is a sub-test of TestGetMaxBufferSize. It tests
// getMaxBufferSize with multiple valid user-set values.
func testGetMaxBufferSizeNoError(t *testing.T) {
	// Unset all keys used for this test
	defer viper.Reset()

	testCasesNoError := []struct {
		bufferSize         string
		expectedBufferSize int
	}{
		{"", int(math.Pow(2, 20))},
		{"2m", int(math.Pow(2, 21))},
		{"4k", int(math.Pow(2, 12))},
		{"1234", 1234},
	}

	for _, tc := range testCasesNoError {
		viper.Set(maxBufferSizeKey, tc.bufferSize)
		size, err := getMaxBufferSize()
		require.NoError(t, err)
		require.Equal(t, tc.expectedBufferSize, size)
	}
}

// testGetMaxBufferSizeWithError is a sub-test of TestGetMaxBufferSize. It tests
// getMaxBufferSize with multiple invalid user-set values.
func testGetMaxBufferSizeWithError(t *testing.T) {
	// Unset all keys used for this test
	defer viper.Reset()

	testCasesWithError := []struct {
		bufferSize string
	}{
		{"3q"},
		{"-1"},
	}

	for _, tc := range testCasesWithError {
		viper.Set(maxBufferSizeKey, tc.bufferSize)
		_, err := getMaxBufferSize()
		require.Error(t, err)
	}
}

// TestGetCleanupTime tests getCleanupTime with/without valid setting cleanup time options
func TestGetCleanupTime(t *testing.T) {
	t.Run("NoError", testGetCleanupTimeNoError)
	t.Run("WithError", testGetCleanupTimeWithError)
}

// testGetCleanupTimeNoError is a sub-test of TestGetCleanupTime. It tests getCleanupTime
// with multiple valid user-set values.
func testGetCleanupTimeNoError(t *testing.T) {
	// Unset all keys used for this test
	defer viper.Reset()

	testCasesNoError := []struct {
		cleanupTime         string
		expectedCleanupTime time.Duration
	}{
		{"3s", time.Duration(3 * time.Second)},
		{"10s", time.Duration(10 * time.Second)},
		{"12s", time.Duration(12 * time.Second)},
	}

	for _, tc := range testCasesNoError {
		viper.Set(cleanupTimeKey, tc.cleanupTime)
		cleanupTime, err := getCleanupTime()
		require.NoError(t, err)
		require.Equal(t, tc.expectedCleanupTime, *cleanupTime)
	}
}

// testGetCleanupTimeWithError is a sub-test of TestGetCleanupTime. It tests getCleanupTime
// with multiple invalid user-set values.
func testGetCleanupTimeWithError(t *testing.T) {
	// Unset all keys used for this test
	defer viper.Reset()

	testCasesWithError := []struct {
		cleanupTime string
	}{
		{"3"},
		{"15s"},
	}

	for _, tc := range testCasesWithError {
		viper.Set(cleanupTimeKey, tc.cleanupTime)
		_, err := getCleanupTime()
		require.Error(t, err)
	}
}
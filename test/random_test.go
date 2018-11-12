package test

import (
	"fmt"
	"github.com/orbs-network/go-mock"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestWithRandLogsCorrectSeedAndTestName(t *testing.T) {
	randPreference.mode = randPrefInvokeClock
	nlMock := namedLoggerMock{name: "MockName1"}

	var loggedSeed int64
	nlMock.When("Log", mock.Any).Call(func(message string) {
		var err error
		tokens := strings.Split(message, " ")
		loggedSeed, err = strconv.ParseInt(tokens[2], 0, 64)
		require.NoError(t, err, "expected third word in log message to be an int64 random seed")
		require.Equal(t, "(MockName1)", tokens[3], "expected fourth word in log message to be the name of the test")
	}).Times(1)

	rand1 := NewRand(&nlMock).Uint64()

	expectedRand1 := rand.New(rand.NewSource(loggedSeed)).Uint64()

	require.Equal(t, expectedRand1, rand1, "expected NewRand() to log the correct random seed")
}


func TestWithExplicitRand(t *testing.T) {
	randPreference.seed = 1
	randPreference.mode = randPrefExplicit

	nlMock := namedLoggerMock{name: "MockName"}
	nlMock.When("Log", "random seed 1 (MockName)").Times(2)

	rand1 := NewRand(&nlMock).Uint64()
	rand2 := NewRand(&nlMock).Uint64()

	expectedRand := rand.New(rand.NewSource(1)).Uint64()

	require.Equal(t, expectedRand, rand1, "expected explicit random seed to produce identical random values")
	require.Equal(t, expectedRand, rand2, "expected explicit random seed to produce identical random values")

	_, err := nlMock.Verify()
	require.NoError(t, err, "expected NewRand to log the explicit random value on each separate invocation")
}

func TestWithLaunchClock(t *testing.T) {
	launchClock := time.Now().UTC().UnixNano()
	randPreference.seed = launchClock
	randPreference.mode = randPrefLaunchClock

	nlMock := namedLoggerMock{name: "MockName"}
	nlMock.When("Log", mock.Any).Call(func(message string){
		require.EqualValues(t, fmt.Sprintf("random seed %v (MockName)", randPreference.seed), message, "expected NewRand to log the launch clock")
	}).Times(2)


	rand1 := NewRand(&nlMock).Uint64()
	rand2 := NewRand(&nlMock).Uint64()

	expectedRand := rand.New(rand.NewSource(launchClock)).Uint64()

	require.Equal(t, expectedRand, rand1, "expected launch-clock random seed to produce identical values on each invocation")
	require.Equal(t, expectedRand, rand2, "expected launch-clock random seed to produce identical values on each invocation")
	require.Equal(t, launchClock, randPreference.seed, "expected seed to not change when calling NewRand in LaunchClock mode")

	_, err := nlMock.Verify()
	require.NoError(t, err)
}

func TestWithInvocationClock(t *testing.T) {
	randPreference.mode = randPrefInvokeClock

	nlMock := namedLoggerMock{name: "MockName"}
	var loggedSeeds []int64
	nlMock.When("Log", mock.Any).Call(func(message string){
		seed, _ := strconv.ParseInt(strings.Split(message, " ")[2], 0, 64)
		loggedSeeds = append(loggedSeeds, seed)
	}).Times(2)

	rand1 := NewRand(&nlMock).Uint64()
	rand2 := NewRand(&nlMock).Uint64()

	require.Equal(t, rand.New(rand.NewSource(loggedSeeds[0])).Uint64(), rand1)
	require.Equal(t, rand.New(rand.NewSource(loggedSeeds[1])).Uint64(), rand2)

	require.NotEqual(t, loggedSeeds[0], loggedSeeds[1], "expected seed values to be different on two NewRand invocations")

	require.True(t,time.Now().UTC().UnixNano() - loggedSeeds[0] < int64(1*time.Millisecond))

	_, err := nlMock.Verify()
	require.NoError(t, err)
}

type namedLoggerMock struct {
	mock.Mock
	name string
}

func (t *namedLoggerMock) Log(args ...interface{}) {
	t.Mock.Called(args...)
}

func (t *namedLoggerMock) Name() string {
	return t.name
}
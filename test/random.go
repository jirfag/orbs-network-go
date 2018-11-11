package test

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type namedLogger interface {
	Log(args ...interface{})
	Name() string
}

type randMode int
const (
	randPrefInvokeClock randMode = iota
	randPrefLaunchClock
	randPrefExplicit
)

type randomPreference struct {
	mode randMode	// default value is randPrefInvokeClock
	seed int64		// applicable only in mode != randPrefInvokeClock
}

var randPreference randomPreference

func (i *randomPreference) String() string {
	var preference string
	switch i.mode {
	case randPrefInvokeClock:
		preference = "clock at invocation (default)"
	case randPrefLaunchClock:
		preference = fmt.Sprintf("launchClock: %v", i.seed)
	case randPrefExplicit:
		preference = fmt.Sprintf("explicit seed: %v", i.seed)
	}
	return preference
}

func (i *randomPreference) Set(value string) error {
	if value == "launchClock" {
		i.mode = randPrefLaunchClock
		i.seed = time.Now().UTC().UnixNano()
		return nil
	}
	i.mode = randPrefExplicit
	v, err := strconv.ParseInt(value, 0, 64)
	i.seed = v
	return err
}

func init() {
	flag.Var(&randPreference, "test.randSeed",
		"Specify a random seed for tests, or 'launchClock' to use" +
		" the same arbitrary value in each test invocation")
}

func WithRand(t namedLogger, f func(r *rand.Rand)) {
	var newSeed int64
	if randPreference.mode == randPrefInvokeClock {
		newSeed = time.Now().UTC().UnixNano()
	} else {
		newSeed = randPreference.seed
	}
	t.Log(fmt.Sprintf("random seed %v (%s)", newSeed, t.Name()))
	f(rand.New(rand.NewSource(newSeed)))
}
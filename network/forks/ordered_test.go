package forks

import (
	"math"
	"testing"

	"github.com/prysmaticlabs/prysm/config/params"
	types "github.com/prysmaticlabs/prysm/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/testing/require"
)

func TestOrderedConfigSchedule(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	for name, getCfg := range params.KnownConfigs {
		cfg := getCfg()
		t.Run(name.String(), func(t *testing.T) {
			prevVersion := [4]byte{0, 0, 0, 0}
			// epoch 0 is genesis, and it's a uint so can't make it -1
			// so we use a pointer to detect the boundary condition and skip it
			var prevEpoch *types.Epoch
			for _, fse := range NewOrderedSchedule(cfg) {
				// copy loop variable so we can take the address of fields
				f := fse
				if prevEpoch == nil {
					prevEpoch = &f.Epoch
					prevVersion = f.Version
					continue
				}
				if *prevEpoch > f.Epoch {
					t.Errorf("Epochs out of order! %#x/%d before %#x/%d", f.Version, f.Epoch, prevVersion, prevEpoch)
				}
				prevEpoch = &f.Epoch
				prevVersion = f.Version
			}
		})
	}

	bc := testForkVersionScheduleBCC()
	ofs := NewOrderedSchedule(bc)
	for i := range ofs {
		if ofs[i].Epoch != types.Epoch(math.Pow(2, float64(i))) {
			t.Errorf("expected %dth element of list w/ epoch=%d, got=%d. list=%v", i, types.Epoch(2^i), ofs[i].Epoch, ofs)
		}
	}
}

func TestVersionForEpoch(t *testing.T) {
	bc := testForkVersionScheduleBCC()
	ofs := NewOrderedSchedule(bc)
	testCases := []struct {
		name    string
		version [4]byte
		epoch   types.Epoch
		err     error
	}{
		{
			name:    "found between versions",
			version: [4]byte{2, 1, 2, 3},
			epoch:   types.Epoch(7),
		},
		{
			name:    "found at end",
			version: [4]byte{4, 1, 2, 3},
			epoch:   types.Epoch(100),
		},
		{
			name:    "found at start",
			version: [4]byte{0, 1, 2, 3},
			epoch:   types.Epoch(1),
		},
		{
			name:    "found at boundary",
			version: [4]byte{1, 1, 2, 3},
			epoch:   types.Epoch(2),
		},
		{
			name:  "not found before",
			epoch: types.Epoch(0),
			err:   ErrVersionNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := ofs.VersionForEpoch(tc.epoch)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.err)
			}
			require.Equal(t, tc.version, v)
		})
	}
}

func testForkVersionScheduleBCC() *params.BeaconChainConfig {
	return &params.BeaconChainConfig{
		ForkVersionSchedule: map[[4]byte]types.Epoch{
			{1, 1, 2, 3}: types.Epoch(2),
			{0, 1, 2, 3}: types.Epoch(1),
			{4, 1, 2, 3}: types.Epoch(16),
			{3, 1, 2, 3}: types.Epoch(8),
			{2, 1, 2, 3}: types.Epoch(4),
		},
	}
}

func TestPrevious(t *testing.T) {
	cfg := testForkVersionScheduleBCC()
	os := NewOrderedSchedule(cfg)
	unreal := [4]byte{255, 255, 255, 255}
	_, err := os.Previous(unreal)
	require.ErrorIs(t, err, ErrVersionNotFound)
	// first element has no previous, should return appropriate error
	_, err = os.Previous(os[0].Version)
	require.ErrorIs(t, err, ErrNoPreviousVersion)
	// work up the list from the second element to the last, make sure each result matches the previous element
	// this test of course relies on TestOrderedConfigSchedule to be correct!
	prev := os[0].Version
	for i := 1; i < len(os); i++ {
		p, err := os.Previous(os[i].Version)
		require.NoError(t, err)
		require.Equal(t, prev, p)
		prev = os[i].Version
	}
}

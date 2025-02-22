package slots

import (
	"math"
	"testing"
	"time"

	"github.com/prysmaticlabs/prysm/config/params"
	types "github.com/prysmaticlabs/prysm/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/testing/assert"
	"github.com/prysmaticlabs/prysm/testing/require"
	prysmTime "github.com/prysmaticlabs/prysm/time"
)

func TestSlotsSinceGenesis(t *testing.T) {
	type args struct {
		genesis time.Time
	}
	tests := []struct {
		name string
		args args
		want types.Slot
	}{
		{
			name: "pre-genesis",
			args: args{
				genesis: prysmTime.Now().Add(1 * time.Hour), // 1 hour in the future
			},
			want: 0,
		},
		{
			name: "post-genesis",
			args: args{
				genesis: prysmTime.Now().Add(-5 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SinceGenesis(tt.args.genesis); got != tt.want {
				t.Errorf("SinceGenesis() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAbsoluteValueSlotDifference(t *testing.T) {
	type args struct {
		x types.Slot
		y types.Slot
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "x_<_y",
			args: args{
				x: types.Slot(3),
				y: types.Slot(4),
			},
			want: 1,
		},
		{
			name: "x_>_y",
			args: args{
				x: types.Slot(100),
				y: types.Slot(4),
			},
			want: 96,
		},
		{
			name: "x_==_y",
			args: args{
				x: types.Slot(100),
				y: types.Slot(100),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AbsoluteValueSlotDifference(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("AbsoluteValueSlotDifference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiplySlotBy(t *testing.T) {
	type args struct {
		times int64
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "multiply by 1",
			args: args{
				times: 1,
			},
			want: time.Duration(12) * time.Second,
		},
		{
			name: "multiply by 2",
			args: args{
				times: 2,
			},
			want: time.Duration(24) * time.Second,
		},
		{
			name: "multiply by 10",
			args: args{
				times: 10,
			},
			want: time.Duration(120) * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MultiplySlotBy(tt.args.times); got != tt.want {
				t.Errorf("MultiplySlotBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEpochStartSlot_OK(t *testing.T) {
	tests := []struct {
		epoch     types.Epoch
		startSlot types.Slot
		error     bool
	}{
		{epoch: 0, startSlot: 0 * params.BeaconConfig().SlotsPerEpoch, error: false},
		{epoch: 1, startSlot: 1 * params.BeaconConfig().SlotsPerEpoch, error: false},
		{epoch: 10, startSlot: 10 * params.BeaconConfig().SlotsPerEpoch, error: false},
		{epoch: 1 << 58, startSlot: 1 << 63, error: false},
		{epoch: 1 << 59, startSlot: 1 << 63, error: true},
		{epoch: 1 << 60, startSlot: 1 << 63, error: true},
	}
	for _, tt := range tests {
		ss, err := EpochStart(tt.epoch)
		if !tt.error {
			require.NoError(t, err)
			assert.Equal(t, tt.startSlot, ss, "EpochStart(%d)", tt.epoch)
		} else {
			require.ErrorContains(t, "start slot calculation overflow", err)
		}
	}
}

func TestEpochEndSlot_OK(t *testing.T) {
	tests := []struct {
		epoch     types.Epoch
		startSlot types.Slot
		error     bool
	}{
		{epoch: 0, startSlot: 1*params.BeaconConfig().SlotsPerEpoch - 1, error: false},
		{epoch: 1, startSlot: 2*params.BeaconConfig().SlotsPerEpoch - 1, error: false},
		{epoch: 10, startSlot: 11*params.BeaconConfig().SlotsPerEpoch - 1, error: false},
		{epoch: 1 << 59, startSlot: 1 << 63, error: true},
		{epoch: 1 << 60, startSlot: 1 << 63, error: true},
		{epoch: math.MaxUint64, startSlot: 0, error: true},
	}
	for _, tt := range tests {
		ss, err := EpochEnd(tt.epoch)
		if !tt.error {
			require.NoError(t, err)
			assert.Equal(t, tt.startSlot, ss, "EpochStart(%d)", tt.epoch)
		} else {
			require.ErrorContains(t, "start slot calculation overflow", err)
		}
	}
}

func TestIsEpochStart(t *testing.T) {
	epochLength := params.BeaconConfig().SlotsPerEpoch

	tests := []struct {
		slot   types.Slot
		result bool
	}{
		{
			slot:   epochLength + 1,
			result: false,
		},
		{
			slot:   epochLength - 1,
			result: false,
		},
		{
			slot:   epochLength,
			result: true,
		},
		{
			slot:   epochLength * 2,
			result: true,
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.result, IsEpochStart(tt.slot), "IsEpochStart(%d)", tt.slot)
	}
}

func TestIsEpochEnd(t *testing.T) {
	epochLength := params.BeaconConfig().SlotsPerEpoch

	tests := []struct {
		slot   types.Slot
		result bool
	}{
		{
			slot:   epochLength + 1,
			result: false,
		},
		{
			slot:   epochLength,
			result: false,
		},
		{
			slot:   epochLength - 1,
			result: true,
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.result, IsEpochEnd(tt.slot), "IsEpochEnd(%d)", tt.slot)
	}
}

func TestSlotsSinceEpochStarts(t *testing.T) {
	tests := []struct {
		slots       types.Slot
		wantedSlots types.Slot
	}{
		{slots: 0, wantedSlots: 0},
		{slots: 1, wantedSlots: 1},
		{slots: params.BeaconConfig().SlotsPerEpoch - 1, wantedSlots: params.BeaconConfig().SlotsPerEpoch - 1},
		{slots: params.BeaconConfig().SlotsPerEpoch + 1, wantedSlots: 1},
		{slots: 10*params.BeaconConfig().SlotsPerEpoch + 2, wantedSlots: 2},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.wantedSlots, SinceEpochStarts(tt.slots))
	}
}

func TestRoundUpToNearestEpoch_OK(t *testing.T) {
	tests := []struct {
		startSlot     types.Slot
		roundedUpSlot types.Slot
	}{
		{startSlot: 0 * params.BeaconConfig().SlotsPerEpoch, roundedUpSlot: 0},
		{startSlot: 1*params.BeaconConfig().SlotsPerEpoch - 10, roundedUpSlot: 1 * params.BeaconConfig().SlotsPerEpoch},
		{startSlot: 10*params.BeaconConfig().SlotsPerEpoch - (params.BeaconConfig().SlotsPerEpoch - 1), roundedUpSlot: 10 * params.BeaconConfig().SlotsPerEpoch},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.roundedUpSlot, RoundUpToNearestEpoch(tt.startSlot), "RoundUpToNearestEpoch(%d)", tt.startSlot)
	}
}

func TestSlotToTime(t *testing.T) {
	type args struct {
		genesisTimeSec uint64
		slot           types.Slot
	}
	tests := []struct {
		name      string
		args      args
		want      time.Time
		wantedErr string
	}{
		{
			name: "slot_0",
			args: args{
				genesisTimeSec: 0,
				slot:           0,
			},
			want: time.Unix(0, 0),
		},
		{
			name: "slot_1",
			args: args{
				genesisTimeSec: 0,
				slot:           1,
			},
			want: time.Unix(int64(1*params.BeaconConfig().SecondsPerSlot), 0),
		},
		{
			name: "slot_12",
			args: args{
				genesisTimeSec: 500,
				slot:           12,
			},
			want: time.Unix(500+int64(12*params.BeaconConfig().SecondsPerSlot), 0),
		},
		{
			name: "overflow",
			args: args{
				genesisTimeSec: 500,
				slot:           math.MaxUint64,
			},
			wantedErr: "is in the far distant future",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToTime(tt.args.genesisTimeSec, tt.args.slot)
			if tt.wantedErr != "" {
				assert.ErrorContains(t, tt.wantedErr, err)
			} else {
				assert.NoError(t, err)
				assert.DeepEqual(t, tt.want, got)
			}
		})
	}
}

func TestVerifySlotTime(t *testing.T) {
	type args struct {
		genesisTime   int64
		slot          types.Slot
		timeTolerance time.Duration
	}
	tests := []struct {
		name      string
		args      args
		wantedErr string
	}{
		{
			name: "Past slot",
			args: args{
				genesisTime: prysmTime.Now().Add(-1 * 5 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second).Unix(),
				slot:        3,
			},
		},
		{
			name: "within tolerance",
			args: args{
				genesisTime: prysmTime.Now().Add(-1 * 5 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second).Add(20 * time.Millisecond).Unix(),
				slot:        5,
			},
		},
		{
			name: "future slot",
			args: args{
				genesisTime: prysmTime.Now().Add(-1 * 5 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second).Unix(),
				slot:        6,
			},
			wantedErr: "could not process slot from the future",
		},
		{
			name: "max future slot",
			args: args{
				genesisTime: prysmTime.Now().Add(-1 * 5 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second).Unix(),
				slot:        types.Slot(MaxSlotBuffer + 6),
			},
			wantedErr: "exceeds max allowed value relative to the local clock",
		},
		{
			name: "evil future slot",
			args: args{
				genesisTime: prysmTime.Now().Add(-1 * 24 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second).Unix(), // 24 slots in the past
				// Gets multiplied with slot duration, and results in an overflow. Wraps around to a valid time.
				// Lower than max signed int. And chosen specifically to wrap to a valid slot 24
				slot: types.Slot((^uint64(0))/params.BeaconConfig().SecondsPerSlot) + 24,
			},
			wantedErr: "is in the far distant future",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyTime(uint64(tt.args.genesisTime), tt.args.slot, tt.args.timeTolerance)
			if tt.wantedErr != "" {
				assert.ErrorContains(t, tt.wantedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSlotClock_HandlesBadSlot(t *testing.T) {
	genTime := prysmTime.Now().Add(-1 * time.Duration(MaxSlotBuffer) * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second).Unix()

	assert.NoError(t, ValidateClock(types.Slot(MaxSlotBuffer), uint64(genTime)), "unexpected error validating slot")
	assert.NoError(t, ValidateClock(types.Slot(2*MaxSlotBuffer), uint64(genTime)), "unexpected error validating slot")
	assert.ErrorContains(t, "which exceeds max allowed value relative to the local clock", ValidateClock(types.Slot(2*MaxSlotBuffer+1), uint64(genTime)), "no error from bad slot")
	assert.ErrorContains(t, "which exceeds max allowed value relative to the local clock", ValidateClock(1<<63, uint64(genTime)), "no error from bad slot")
}

func TestPrevSlot(t *testing.T) {
	tests := []struct {
		name string
		slot types.Slot
		want types.Slot
	}{
		{
			name: "no underflow",
			slot: 0,
			want: 0,
		},
		{
			name: "slot 1",
			slot: 1,
			want: 0,
		},
		{
			name: "slot 2",
			slot: 2,
			want: 1,
		},
		{
			name: "max",
			slot: math.MaxUint64,
			want: math.MaxUint64 - 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrevSlot(tt.slot); got != tt.want {
				t.Errorf("PrevSlot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncCommitteePeriod(t *testing.T) {
	tests := []struct {
		epoch  types.Epoch
		wanted uint64
	}{
		{epoch: 0, wanted: 0},
		{epoch: 0, wanted: 0 / uint64(params.BeaconConfig().EpochsPerSyncCommitteePeriod)},
		{epoch: 1, wanted: 1 / uint64(params.BeaconConfig().EpochsPerSyncCommitteePeriod)},
		{epoch: 1000, wanted: 1000 / uint64(params.BeaconConfig().EpochsPerSyncCommitteePeriod)},
	}
	for _, test := range tests {
		require.Equal(t, test.wanted, SyncCommitteePeriod(test.epoch))
	}
}

func TestSyncCommitteePeriodStartEpoch(t *testing.T) {
	tests := []struct {
		epoch  types.Epoch
		wanted types.Epoch
	}{
		{epoch: 0, wanted: 0},
		{epoch: params.BeaconConfig().EpochsPerSyncCommitteePeriod + 1, wanted: params.BeaconConfig().EpochsPerSyncCommitteePeriod},
		{epoch: params.BeaconConfig().EpochsPerSyncCommitteePeriod*2 + 100, wanted: params.BeaconConfig().EpochsPerSyncCommitteePeriod * 2},
		{epoch: params.BeaconConfig().EpochsPerSyncCommitteePeriod*params.BeaconConfig().EpochsPerSyncCommitteePeriod + 1, wanted: params.BeaconConfig().EpochsPerSyncCommitteePeriod * params.BeaconConfig().EpochsPerSyncCommitteePeriod},
	}
	for _, test := range tests {
		e, err := SyncCommitteePeriodStartEpoch(test.epoch)
		require.NoError(t, err)
		require.Equal(t, test.wanted, e)
	}
}

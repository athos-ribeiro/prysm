package v2_test

import (
	"context"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	coreState "github.com/prysmaticlabs/prysm/beacon-chain/core/transition"
	v2 "github.com/prysmaticlabs/prysm/beacon-chain/state/v2"
	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/testing/assert"
	"github.com/prysmaticlabs/prysm/testing/util"
)

func FuzzV2StateHashTreeRoot(f *testing.F) {
	gState, _ := util.DeterministicGenesisStateAltair(f, 100)
	output, err := gState.MarshalSSZ()
	assert.NoError(f, err)
	f.Add(output[:100], uint64(10))
	f.Fuzz(func(t *testing.T, diffBuffer []byte, slotsToTransition uint64) {
		for i := 0; i < len(diffBuffer); i += 9 {
			if i+8 >= len(diffBuffer) {
				return
			}
			num := bytesutil.BytesToUint64BigEndian(diffBuffer[i : i+8])
			num %= uint64(len(diffBuffer))
			output[num] ^= diffBuffer[i+8]
		}
		pbState := &ethpb.BeaconStateAltair{}
		err := pbState.UnmarshalSSZ(output)
		if err != nil {
			return
		}
		slotsToTransition %= 100
		stateObj, err := v2.InitializeFromProtoUnsafe(pbState)
		assert.NoError(t, err)
		for stateObj.Slot() < types.Slot(slotsToTransition) {
			stateObj, err = coreState.ProcessSlots(context.Background(), stateObj, stateObj.Slot()+1)
			assert.NoError(t, err)
			stateObj.Copy()
		}
		assert.NoError(t, err)
		innerState := stateObj.InnerStateUnsafe().(*ethpb.BeaconStateAltair)
		rt, pbErr := innerState.HashTreeRoot()
		newRt, err := stateObj.HashTreeRoot(context.Background())
		assert.Equal(t, pbErr != nil, err != nil)
		if err == nil {
			assert.Equal(t, rt, newRt)
		}
	})
}
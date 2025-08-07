package query

import (
	"fmt"
	"math/bits"
)

const (
	// ビットセット制限
	MaxComponentTypes = 64

	// デフォルト値
	DefaultBitSetCapacity = MaxComponentTypes

	// エラーメッセージ
	ErrInvalidComponentType   = "invalid component type"
	ErrComponentLimitExceeded = "component type limit exceeded"
)

// String returns a string representation of the bitset
func (b ComponentBitSet) String() string {
	return fmt.Sprintf("0b%064b", uint64(b))
}

// Count returns the number of set bits
func (b ComponentBitSet) Count() int {
	return bits.OnesCount64(uint64(b))
}

// IsEmpty returns true if no bits are set
func (b ComponentBitSet) IsEmpty() bool {
	return b == 0
}

// IsFull returns true if all bits are set (within component limit)
func (b ComponentBitSet) IsFull() bool {
	return b == (1<<len(componentTypeToBitPosition))-1
}

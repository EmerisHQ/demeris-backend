// Code generated by "stringer -type=chainStatus"; DO NOT EDIT.

package chainwatch

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[starting-0]
	_ = x[running-1]
	_ = x[relayerConnecting-2]
	_ = x[done-3]
}

const _chainStatus_name = "startingrunningrelayerConnectingdone"

var _chainStatus_index = [...]uint8{0, 8, 15, 32, 36}

func (i chainStatus) String() string {
	if i >= chainStatus(len(_chainStatus_index)-1) {
		return "chainStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _chainStatus_name[_chainStatus_index[i]:_chainStatus_index[i+1]]
}
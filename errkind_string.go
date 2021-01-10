// Code generated by "stringer -linecomment -type ErrKind"; DO NOT EDIT.

package cep

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[CEPNotFound-1]
	_ = x[ContextCanceled-2]
	_ = x[InvalidCEP-4]
	_ = x[Timeout-8]
	_ = x[UnmarshalErr-16]
	_ = x[Other-32]
}

const (
	_ErrKind_name_0 = "CEP Not FoundContext Canceled"
	_ErrKind_name_1 = "Invalid CEP"
	_ErrKind_name_2 = "Timeout"
	_ErrKind_name_3 = "Unmarshal error"
	_ErrKind_name_4 = "Other"
)

var (
	_ErrKind_index_0 = [...]uint8{0, 13, 29}
)

func (i ErrKind) String() string {
	switch {
	case 1 <= i && i <= 2:
		i -= 1
		return _ErrKind_name_0[_ErrKind_index_0[i]:_ErrKind_index_0[i+1]]
	case i == 4:
		return _ErrKind_name_1
	case i == 8:
		return _ErrKind_name_2
	case i == 16:
		return _ErrKind_name_3
	case i == 32:
		return _ErrKind_name_4
	default:
		return "ErrKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
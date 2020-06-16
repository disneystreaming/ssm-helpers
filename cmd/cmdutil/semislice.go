package cmdutil

import (
	"strings"
)

// -- semiSlice Value
type semiSliceValue struct {
	value   *[]string
	changed bool
}

func newSemiSliceValue(val []string, p *[]string) *semiSliceValue {
	ssv := new(semiSliceValue)
	ssv.value = p
	*ssv.value = val
	return ssv
}

func (s *semiSliceValue) Set(val string) error {
	v := readAsSSV(val)
	if !s.changed {
		*s.value = v
	} else {
		*s.value = append(*s.value, v...)
	}
	s.changed = true
	return nil
}

func (s *semiSliceValue) Type() string {
	return "stringSlice"
}

func (s *semiSliceValue) String() string {
	str, _ := writeAsSSV(*s.value)
	return "[" + str + "]"
}

func (s *semiSliceValue) Append(val string) error {
	*s.value = append(*s.value, val)
	return nil
}

func (s *semiSliceValue) Replace(val []string) error {
	*s.value = val
	return nil
}

func (s *semiSliceValue) GetSlice() []string {
	return *s.value
}

func writeAsSSV(vals []string) (string, error) {
	return strings.TrimSuffix(strings.Join(vals, ";"), ";"), nil
}

func readAsSSV(val string) []string {
	if val == "" {
		return []string{}
	}

	return strings.Split(val, ";")
}

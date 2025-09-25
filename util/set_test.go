package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestSet(t *testing.T) {
	s1 := Set{}
	s1.Add("red")
	s1.Add("orange")
	s1.Add("yellow")
	s1.Add("yellow")
	s1.Add("red")
	s1.Add("red")
	s1.Add("red")
	s1.Add("green")
	assert.Equal(t, len(s1), 4)

	s2 := Set{}
	s2.Add("green")
	s2.Add("green")
	s2.Add("green")
	s2.Add("orange")
	s2.Add("yellow")
	s2.Add("yellow")
	s2.Add("orange")
	s2.Add("red")
	s2.Add("orange")
	s2.Add("red")
	s2.Add("red")
	s2.Add("green")
	assert.Equal(t, len(s2), 4)
	assert.DeepEqual(t, s1, s2)
	assert.Assert(t, s2.Contains("green"))

	s1.Remove("orange")
	s1.Remove("green")
	assert.Equal(t, len(s1), 2)
	assert.Assert(t, !s1.Contains("green"))

	s1.Remove("green")
	assert.Equal(t, len(s1), 2)

	s3 := Set{}
	s3.Add("green", "red", "orange", "red", "yellow", "yellow", "green")
	assert.Equal(t, len(s3), 4)
	assert.DeepEqual(t, s2, s3)

	slice := s3.ToSlice()
	assert.Equal(t, len(slice), 4)
	assert.Assert(t, ContainsAny(slice, []string{"green", "red", "orange", "yellow"}))

	s4 := Set{}
	slice = s4.ToSlice()
	assert.Equal(t, len(slice), 0)
}

package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIntersect(t *testing.T) {
	int_tests := []struct {
		a        []int
		b        []int
		expected []int
	}{
		{[]int{1, 2, 3}, []int{2, 4, 6}, []int{2}},
	}

	string_tests := []struct {
		a        []string
		b        []string
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"b", "i", "n"}, []string{"b"}},
	}

	for _, tt := range int_tests {
		res := Intersect(tt.a, tt.b)
		if !cmp.Equal(res, tt.expected) {
			t.Errorf("Intersect(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, res)
		}
	}
	for _, tt := range string_tests {
		res := Intersect(tt.a, tt.b)
		if !cmp.Equal(res, tt.expected) {
			t.Errorf("Intersect(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, res)
		}
	}
}

func TestUnion(t *testing.T) {
	int_tests := []struct {
		a        []int
		b        []int
		expected []int
	}{
		{[]int{1, 2, 3}, []int{2, 4, 6}, []int{1, 2, 3, 4, 6}},
	}

	string_tests := []struct {
		a        []string
		b        []string
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"b", "i", "n"}, []string{"a", "b", "c", "i", "n"}},
	}

	for _, tt := range int_tests {
		res := Union(tt.a, tt.b)
		if !cmp.Equal(res, tt.expected) {
			t.Errorf("Union(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, res)
		}
	}
	for _, tt := range string_tests {
		res := Union(tt.a, tt.b)
		if !cmp.Equal(res, tt.expected) {
			t.Errorf("Union(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, res)
		}
	}
}

func TestDifference(t *testing.T) {
	int_tests := []struct {
		a        []int
		b        []int
		expected []int
	}{
		{[]int{1, 2, 3}, []int{2, 4, 6}, []int{1, 3}},
	}

	string_tests := []struct {
		a        []string
		b        []string
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"b", "i", "n"}, []string{"a", "c"}},
	}

	for _, tt := range int_tests {
		res := Difference(tt.a, tt.b)
		if !cmp.Equal(res, tt.expected) {
			t.Errorf("Difference(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, res)
		}
	}
	for _, tt := range string_tests {
		res := Difference(tt.a, tt.b)
		if !cmp.Equal(res, tt.expected) {
			t.Errorf("Difference(%v, %v): expected %v, got %v", tt.a, tt.b, tt.expected, res)
		}
	}
}

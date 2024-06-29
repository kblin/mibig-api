/*
Copyright © 2024 Technical University of Denmark - written by Kai Blin <kblin@biosustain.dtu.dk>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils

import (
	"cmp"
	"slices"
)

func sortSlice[T cmp.Ordered](s []T) {
	slices.SortFunc(s, func(a, b T) int {
		return cmp.Compare(a, b)
	})
}

func Intersect[T comparable](a, b []T) []T {
	hash := make(map[T]bool)
	res := make([]T, 0)

	for _, i := range a {
		hash[i] = true
	}
	for _, i := range b {
		if _, ok := hash[i]; ok {
			res = append(res, i)
		}
	}

	return res
}

func Union[T cmp.Ordered](a, b []T) []T {
	hash := make(map[T]bool)
	res := make([]T, 0)

	for _, i := range a {
		hash[i] = true
	}
	for _, i := range b {
		hash[i] = true
	}

	for key := range hash {
		res = append(res, key)
	}

	sortSlice(res)
	return res
}

func Difference[T comparable](a, b []T) []T {
	hash := make(map[T]bool)
	res := make([]T, 0)

	for _, i := range b {
		hash[i] = true
	}
	for _, i := range a {
		if _, ok := hash[i]; !ok {
			res = append(res, i)
		}
	}

	return res
}

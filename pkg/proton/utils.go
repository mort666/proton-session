package proton

import "slices"

func Filter[S ~[]E, E any](s S, keep func(E) bool) S {
	return slices.DeleteFunc(slices.Clone(s), func(e E) bool {
		return !keep(e)
	})
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

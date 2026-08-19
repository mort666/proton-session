/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-19 22:00:08
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-19 22:00:08
 */

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

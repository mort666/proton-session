/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-19 22:02:17
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-19 22:02:17
 */

package proton

type Status int

const (
	StatusUp Status = iota
	StatusDown
)

func (s Status) String() string {
	switch s {
	case StatusUp:
		return "up"

	case StatusDown:
		return "down"

	default:
		return "unknown"
	}
}

type StatusObserver func(Status)

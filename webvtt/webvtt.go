package webvtt

import "github.com/thiagopnts/caps"

func NewReader(ignoreTimingErrors bool) caps.CaptionReader {
	return &Reader{
		ignoreTimingErrors,
	}
}

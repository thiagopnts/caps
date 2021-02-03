package webvtt

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/thiagopnts/caps"
)

var timingPattern = regexp.MustCompile("^(.+?) --> (.+)")
var timestampPattern = regexp.MustCompile(`^(\d+):(\d{2})(:\d{2})?\.(\d{3})`)
var voiceSpanPattern = regexp.MustCompile(`<v(\\.\\w+)* ([^>]*)>`)
var otherSpanPattern = regexp.MustCompile("</?([cibuv]|ruby|rt|lang).*?>")
var webvttTiming = "-->"

func microseconds(h, m, s, f string) (float64, error) {
	hh, err := strconv.ParseFloat(h, 64)
	if err != nil {
		return 0, err
	}
	mm, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return 0, err
	}
	ss, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	ff, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return 0, err
	}
	return (hh*3600+mm*60+ss)*1000000 + ff*1000, nil
}

type Reader struct {
	ignoreTimingErrors bool
}

func (r *Reader) Read(content []byte) (*caps.CaptionSet, error) {
	return r.ReadString(string(content))
}

func (r *Reader) ReadString(content string) (*caps.CaptionSet, error) {
	captionSet := caps.NewCaptionSet()
	captions, err := r.parse(caps.SplitLines(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing webvtt: %w", err)
	}
	captionSet.SetCaptions(caps.DefaultLang, captions)
	if captionSet.IsEmpty() {
		return nil, fmt.Errorf("empty caption file")
	}
	return captionSet, nil
}

func (r *Reader) parse(lines []string) ([]*caps.Caption, error) {
	captions := []*caps.Caption{}
	foundTiming := false
	var caption *caps.Caption
	var err error
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, webvttTiming) {
			foundTiming = true
			timingLine := i
			lastStartTime := 0.0
			if len(captions) != 0 {
				lastStartTime = captions[len(captions)-1].Start
			}
			caption, err = r.parseTimingLine(line, lastStartTime)
			if err != nil {
				return nil, fmt.Errorf("error parsing timing line %d: %w", timingLine, err)
			}
		} else if line == "" {
			if foundTiming {
				foundTiming = false
				if caption != nil && !caption.IsEmpty() {
					captions = append(captions, caption)
				}
				caption = nil
			}
		} else {
			if foundTiming {
				if caption != nil && !caption.IsEmpty() {
					caption.Nodes = append(caption.Nodes, caps.NewLineBreak())
				}
				caption.Nodes = append(caption.Nodes, caps.NewCaptionText(r.removeStyles(line)))
			}
		}
	}
	if caption != nil && !caption.IsEmpty() {
		captions = append(captions, caption)
	}
	return captions, nil
}

func (r *Reader) validateTimings(caption *caps.Caption, lastStartTime float64) error {
	// FIXME: we might need to use a *float64 for Start/End so we can check for unset values.
	//if caption.Start == nil {
	//}
	if caption.Start > caption.End {
		return fmt.Errorf("end timestamp is not greater than start timestamp")
	}

	if caption.Start < lastStartTime {
		return fmt.Errorf("start timestamp is not greater to start timestamp of previous cue")
	}
	return nil
}

func (r *Reader) removeStyles(line string) string {
	partialResult := voiceSpanPattern.ReplaceAllString(line, "\\2: ")
	return otherSpanPattern.ReplaceAllString(partialResult, "")
}

func (r *Reader) parseTimingLine(line string, lastStartTime float64) (*caps.Caption, error) {
	matches := timingPattern.FindAllString(line, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid timing format")
	}
	start, err := r.parseTimestamp(string(matches[1]))
	if err != nil {
		return nil, err
	}
	end, err := r.parseTimestamp(string(matches[2]))
	if err != nil {
		return nil, err
	}
	caption := &caps.Caption{Start: start, End: end}
	if !r.ignoreTimingErrors {
		r.validateTimings(caption, lastStartTime)
	}
	return caption, nil
}

func (r *Reader) parseTimestamp(input string) (float64, error) {
	matches := timestampPattern.FindAllString(input, -1)
	if len(matches) < 4 {
		return 0, nil
	}
	if matches[2] != "" {
		return microseconds(matches[0], matches[1], strings.ReplaceAll(matches[2], ":", ""), matches[3])
	}
	return microseconds("0", matches[0], matches[1], matches[3])
}

func (r *Reader) Detect(content []byte) bool {
	return r.DetectString(string(content))
}

func (r *Reader) DetectString(content string) bool {
	return strings.Contains(content, "WEBVTT")
}

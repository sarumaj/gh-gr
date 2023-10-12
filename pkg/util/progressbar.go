package util

import (
	"flag"
	"fmt"
	"io"

	progressbar "github.com/schollz/progressbar/v3"
)

type Progressbar struct {
	*progressbar.ProgressBar
	w io.Writer
}

type ProgressbarOption = progressbar.Option

var (
	ClearOnFinish            = progressbar.OptionClearOnFinish
	EnableColorCodes         = progressbar.OptionEnableColorCodes
	FullWidth                = progressbar.OptionFullWidth
	OnCompletion             = progressbar.OptionOnCompletion
	SetDescription           = progressbar.OptionSetDescription
	SetElapsedTime           = progressbar.OptionSetElapsedTime
	SetItsString             = progressbar.OptionSetItsString
	SetPredictTime           = progressbar.OptionSetPredictTime
	SetRenderBlankState      = progressbar.OptionSetRenderBlankState
	SetTheme                 = progressbar.OptionSetTheme
	SetVisibility            = progressbar.OptionSetVisibility
	SetWidth                 = progressbar.OptionSetWidth
	SetWriter                = progressbar.OptionSetWriter
	ShowBytes                = progressbar.OptionShowBytes
	ShowCount                = progressbar.OptionShowCount
	ShowDescriptionAtLineEnd = progressbar.OptionShowDescriptionAtLineEnd
	ShowElapsedTimeOnFinish  = progressbar.OptionShowElapsedTimeOnFinish
	ShowIts                  = progressbar.OptionShowIts
	SpinnerCustom            = progressbar.OptionSpinnerCustom
	SpinnerType              = progressbar.OptionSpinnerType
	Throttle                 = progressbar.OptionThrottle
	UseANSICodes             = progressbar.OptionUseANSICodes
)

func (p *Progressbar) Add(i int) *Progressbar {
	_ = p.ProgressBar.Add(i)
	return p
}

func (p *Progressbar) ChangeMax(max int) *Progressbar {
	p.ProgressBar.ChangeMax64(int64(max))
	return p
}

func (p *Progressbar) Clear() *Progressbar {
	_ = p.ProgressBar.Clear()
	return p
}

func (p *Progressbar) Describe(format string, a ...any) *Progressbar {
	p.ProgressBar.Describe(fmt.Sprintf(format, a...))
	return p
}

func (p *Progressbar) Inc() *Progressbar {
	_ = p.Add(1)
	return p
}

func NewProgressbar(m int, options ...ProgressbarOption) *Progressbar {
	c := Console()
	p := &Progressbar{w: c.Stdout()}

	if options == nil {
		options = []ProgressbarOption{
			EnableColorCodes(c.ColorsEnabled()),
			SetWidth(20),
			ShowCount(),
			ShowElapsedTimeOnFinish(),
			ClearOnFinish(),
		}
	}

	if interactive := c.IsTerminal(true, false, false); !interactive || flag.Lookup("test.v") != nil {
		p.w = io.Discard

	} else {
		options = append(options, UseANSICodes(true))

	}

	options = append(options, progressbar.OptionSetWriter(p.w))
	p.ProgressBar = progressbar.NewOptions64(int64(m), options...)

	return p
}

package util

import (
	"flag"
	"fmt"
	"io"
	"os"

	term "github.com/cli/go-gh/v2/pkg/term"
	progressbar "github.com/schollz/progressbar/v3"
)

type Progressbar struct {
	*progressbar.ProgressBar
	w               io.Writer
	msgOnCompletion string
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

func (p *Progressbar) Print() {
	fmt.Fprint(p.w, p.msgOnCompletion)
}

func (p *Progressbar) SetMessageOnCompletion(msg string) *Progressbar {
	p.msgOnCompletion = msg
	return p
}

func NewProgressbar(m int, options ...ProgressbarOption) *Progressbar {
	p := &Progressbar{w: os.Stdout}

	if options == nil {
		options = []ProgressbarOption{
			EnableColorCodes(UseColors()),
			SetWidth(20),
			ShowCount(),
			ShowElapsedTimeOnFinish(),
			ClearOnFinish(),
			OnCompletion(p.Print),
		}
	}

	if interactive := term.IsTerminal(os.Stdout); !interactive || flag.Lookup("test.v") != nil {
		p.w = io.Discard
	} else if interactive {
		options = append(options, UseANSICodes(true))
	}

	options = append(options, progressbar.OptionSetWriter(p.w))
	p.ProgressBar = progressbar.NewOptions64(int64(m), options...)

	return p
}

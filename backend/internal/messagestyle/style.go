package messagestyle

import (
	"fmt"
	"regexp"
	"strings"
)

var allowedFamilies = map[string]struct{}{"Arial":{},"Tahoma":{},"Verdana":{},"Georgia":{},"Times New Roman":{},"Courier New":{}}
var colorPattern = regexp.MustCompile("^#[0-9A-Fa-f]{6}$")

type Style struct {
	FontFamily string `json:"fontFamily"`
	FontSize int `json:"fontSize"`
	Color string `json:"color"`
	Bold bool `json:"bold"`
	Italic bool `json:"italic"`
	Underline bool `json:"underline"`
}

func DefaultStyle() Style { return Style{FontFamily:"Arial",FontSize:14,Color:"#FFFFFF"} }

func (s Style) Normalize() (Style,error) {
	s.FontFamily = strings.TrimSpace(s.FontFamily)
	if _,ok := allowedFamilies[s.FontFamily]; !ok { return Style{},fmt.Errorf("fontFamily is not supported") }
	if s.FontSize < 10 || s.FontSize > 28 { return Style{},fmt.Errorf("fontSize must be between 10 and 28") }
	s.Color = strings.ToUpper(strings.TrimSpace(s.Color))
	if !colorPattern.MatchString(s.Color) { return Style{},fmt.Errorf("color must be a six-digit hexadecimal color") }
	return s,nil
}

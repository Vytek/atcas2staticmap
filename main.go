//main2.go

package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"image/color"

	"github.com/gocarina/gocsv"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
)

type Record struct {
	TIME string `csv:"TIME"`
	LAT  string `csv:"LAT"`
	LON  string `csv:"LON"`
}

func StringToFloat64(data string) float64 {
	data = strings.Replace(data, ",", ".", -1)
	if s, err := strconv.ParseFloat(data, 64); err == nil {
		return s
	} else {
		fmt.Println(err.Error())
		return 0.0
	}
}

func IntToString(n int) string {
	return strconv.Itoa(n)
}

// NOTE: this isn't multi-Unicode-codepoint aware, like specifying skintone or
//
//	gender of an emoji: https://unicode.org/emoji/charts/full-emoji-modifiers.html
func substr(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}

// TextMarker is an MapObject that displays a text and has a pointy tip:
//
//	+------------+
//	| text label |
//	+----\  /----+
//	      \/
type TextMarker struct {
	sm.MapObject
	Position   s2.LatLng
	Text       string
	TextWidth  float64
	TextHeight float64
	TipSize    float64
}

// NewTextMarker creates a new TextMarker
func NewTextMarker(pos s2.LatLng, text string) *TextMarker {
	s := new(TextMarker)
	s.Position = pos
	s.Text = text
	s.TipSize = 16.0

	d := &font.Drawer{
		Face: basicfont.Face7x13,
	}
	s.TextWidth = float64(d.MeasureString(s.Text) >> 6)
	s.TextHeight = 13.0
	return s
}

// ExtraMarginPixels returns the left, top, right, bottom pixel margin of the TextMarker object.
func (s *TextMarker) ExtraMarginPixels() (float64, float64, float64, float64) {
	w := math.Max(4.0+s.TextWidth, 2*s.TipSize)
	h := s.TipSize + s.TextHeight + 4.0
	return w * 0.5, h, w * 0.5, 0.0
}

// Bounds returns the bounding rectangle of the TextMarker object, which is just the tip position.
func (s *TextMarker) Bounds() s2.Rect {
	r := s2.EmptyRect()
	r = r.AddPoint(s.Position)
	return r
}

// Draw draws the object.
func (s *TextMarker) Draw(gc *gg.Context, trans *sm.Transformer) {
	if !sm.CanDisplay(s.Position) {
		return
	}

	w := math.Max(4.0+s.TextWidth, 2*s.TipSize)
	h := s.TextHeight + 4.0
	x, y := trans.LatLngToXY(s.Position)
	gc.ClearPath()
	gc.SetLineWidth(1)
	gc.SetLineCap(gg.LineCapRound)
	gc.SetLineJoin(gg.LineJoinRound)
	gc.LineTo(x, y)
	gc.LineTo(x-s.TipSize, y-s.TipSize)
	gc.LineTo(x-w*0.5, y-s.TipSize)
	gc.LineTo(x-w*0.5, y-s.TipSize-h)
	gc.LineTo(x+w*0.5, y-s.TipSize-h)
	gc.LineTo(x+w*0.5, y-s.TipSize)
	gc.LineTo(x+s.TipSize, y-s.TipSize)
	gc.LineTo(x, y)
	gc.SetColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	gc.FillPreserve()
	gc.SetColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
	gc.Stroke()

	gc.SetRGBA(0.0, 0.0, 0.0, 1.0)
	gc.DrawString(s.Text, x-s.TextWidth*0.5, y-s.TipSize-4.0)
}

type TrackInput struct {
	FilePath string
	Color    color.RGBA
}

// Modifica la funzione per accettare []TrackInput invece di []string
func CreateTracksImage(tracks []TrackInput, outputImage string) error {
	ctx := sm.NewContext()
	ctx.SetSize(1920, 1080)

	for _, track := range tracks {
		file, err := os.Open(track.FilePath)
		if err != nil {
			return fmt.Errorf("errore apertura file %s: %w", track.FilePath, err)
		}

		var records []Record
		if err := gocsv.UnmarshalFile(file, &records); err != nil {
			file.Close()
			return fmt.Errorf("errore parsing file %s: %w", track.FilePath, err)
		}
		file.Close()

		if len(records) == 0 {
			continue
		}

		v := len(records) - 1
		path := make([]s2.LatLng, 0, len(records)+2)

		//Flight
		var StrToShow string
		flight := substr(track.FilePath, 0, 4)
		if flight == "1136" {
			StrToShow = "DC9 ITAVIA IH870 (Bologna-Palermo) 1136"
		} else if flight == "1133" {
			StrToShow = "Fokker 28 ITAVIA IH779 (Bergamo-Roma) 1133"
		} else if flight == "1132" {
			StrToShow = "ATI BM300 (Trieste-Roma) 1132"
		} else if flight == "0225" {
			StrToShow = "B727 I-DIRU (Pezzopane) Alitalia AZ865 (Tunisi-Fiumicino) 0225"
		} else if flight == "0226" {
			StrToShow = "Bea Tours KT881 (Malta-Londra) 0226"
		} else if flight == "1235" {
			StrToShow = "Boing 720 Air Malta KM153 (Londra-Malta) 1235"
		} else if flight == "4200" {
			StrToShow = "TF-104G (20-4 MM54230) Bergamini/Moretti (Intermedia 8)"
		} else {
			StrToShow = "Unknown"
		}

		for i, record := range records {
			lat := StringToFloat64(record.LAT)
			lon := StringToFloat64(record.LON)
			pos := s2.LatLngFromDegrees(lat, lon)

			if i == 0 {
				FirstText := NewTextMarker(pos, record.TIME+" "+StrToShow)
				ctx.AddObject(FirstText)
				first := sm.NewMarker(pos, color.RGBA{0, 255, 0, 255}, 8.0)
				ctx.AddObject(first)
				path = append(path, first.Position)
			}
			if i == v {
				LastText := NewTextMarker(pos, record.TIME)
				ctx.AddObject(LastText)
				last := sm.NewMarker(pos, color.RGBA{255, 0, 0, 255}, 8.0)
				ctx.AddObject(last)
				path = append(path, last.Position)
			}
			path = append(path, pos)
		}
		// Usa il colore specificato per la traccia
		ctx.AddObject(sm.NewPath(path, track.Color, 4.0))
	}

	img, err := ctx.Render()
	if err != nil {
		return err
	}

	if err := gg.SavePNG(outputImage, img); err != nil {
		return err
	}
	return nil
}

func HexToRGBA(hex string) color.RGBA {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b uint8 = 0, 0, 0
	if len(hex) == 6 {
		ri, _ := strconv.ParseUint(hex[0:2], 16, 8)
		gi, _ := strconv.ParseUint(hex[2:4], 16, 8)
		bi, _ := strconv.ParseUint(hex[4:6], 16, 8)
		r, g, b = uint8(ri), uint8(gi), uint8(bi)
	}
	return color.RGBA{r, g, b, 255}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: main.go file1.csv:#RRGGBB file2.csv:#RRGGBB ...")
		return
	}
	var tracks []TrackInput
	for _, arg := range os.Args[1:] {
		parts := strings.Split(arg, ":")
		if len(parts) != 2 {
			fmt.Printf("Argomento non valido: %s. Usa file.csv:#RRGGBB\n", arg)
			continue
		}
		tracks = append(tracks, TrackInput{
			FilePath: parts[0],
			Color:    HexToRGBA(parts[1]),
		})
	}
	if len(tracks) == 0 {
		fmt.Println("Nessuna traccia valida fornita.")
		return
	}
	err := CreateTracksImage(tracks, "tutte_le_tracce.png")
	if err != nil {
		panic(err)
	}
}

package main

import (
	"flag"
	"math"
	"os"
	"time"

	"codello.dev/ultrastar"
	"codello.dev/ultrastar/txt"
	"github.com/brianvoe/gofakeit/v6"
)

var (
	duet   bool
	output string
)

func init() {
	flag.BoolVar(&duet, "duet", false, "generate a duet instead of a solo song")
	flag.StringVar(&output, "output", "", "write the generated song to this file")
	flag.Parse()
}

func main() {
	song := ultrastar.NewSong()
	song.Title = gofakeit.BookTitle()
	song.Artist = gofakeit.BookAuthor()
	song.Genre = gofakeit.BookGenre()
	song.Edition = ""
	song.Creator = gofakeit.Username()
	song.Language = gofakeit.Language()
	song.Year = gofakeit.Number(1930, 2050)
	song.Comment = gofakeit.SentenceSimple()

	song.Gap = time.Duration(gofakeit.Number(0, 30000)) * time.Millisecond

	song.MusicP1 = genMusic()

	if duet {
		song.DuetSinger2 = gofakeit.Name()
		song.DuetSinger1 = gofakeit.Name()
		song.MusicP2 = genMusic()
	}

	song.SetBPM(ultrastar.BPM(math.Round(gofakeit.Float64Range(320, 1000)*100) / 100))

	file := os.Stdout
	var err error
	if output != "" {
		if file, err = os.Create(output); err != nil {
			println(err)
			os.Exit(1)
		}
	}
	_ = txt.WriteSong(file, song)
	if err = file.Close(); err != nil {
		println(err)
		os.Exit(1)
	}
}

func genMusic() *ultrastar.Music {
	m := ultrastar.NewMusic()
	beat := ultrastar.Beat(gofakeit.Number(0, 1000))
	numLines := gofakeit.Number(5, 20)
	for l := 0; l < numLines; l++ {
		numNotes := gofakeit.Number(4, 8)
		rap := gofakeit.Number(1, 30) == 15
		for n := 0; n < numNotes; n++ {
			beat += ultrastar.Beat(gofakeit.Number(1, 20))
			note := genNote(rap)
			note.Start = beat
			beat += note.Duration
			m.Notes = append(m.Notes, note)
		}
		if l < numLines-1 {
			beat += ultrastar.Beat(gofakeit.Number(1, 20))
			note := ultrastar.Note{
				Type:  ultrastar.NoteTypeLineBreak,
				Start: beat,
			}
			m.Notes = append(m.Notes, note)
		}
	}
	return m
}

func genNote(rap bool) (n ultrastar.Note) {
	if gofakeit.Number(0, 10) == 2 {
		if rap {
			n.Type = ultrastar.NoteTypeGoldenRap
		} else {
			n.Type = ultrastar.NoteTypeGolden
		}
	} else {
		if rap {
			n.Type = ultrastar.NoteTypeRap
		} else {
			n.Type = ultrastar.NoteTypeRegular
		}
	}
	n.Duration = ultrastar.Beat(gofakeit.Number(1, 12))
	n.Pitch = ultrastar.Pitch(gofakeit.Number(0, 16))
	n.Text = gofakeit.Word()
	return
}

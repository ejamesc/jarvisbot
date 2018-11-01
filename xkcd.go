package jarvisbot

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/tucnak/telebot"
)

// Constants

// URL for looking up xkcd comic titles to their url / comic img src
var URL = "http://www.explainxkcd.com/wiki/index.php?title=List_of_all_comics_(full)&printable=yes"
var titleWordWeight = 400
var titleIntextWeight = 200
var transcriptWeight = 30
var textWordWeight = 8
var textWeight = 1
var comics []XKCDComic

// XKCDComic is a struct defined for comics scraped from URL
type XKCDComic struct {
	Number      int      `json:"number"`
	Title       string   `json:"title"`
	TitleText   string   `json:"titletext"`
	TitleFields []string `json:"-"`
	URL         string   `json:"url"`
	Image       string   `json:"image"`
	Date        string   `json:"date"`
	Text        string   `json:"-"`
}

func (j *JarvisBot) SearchXkcd(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyTo: *msg.Message, ReplyMarkup: telebot.ReplyMarkup{ForceReply: true, Selective: true}}
		j.SendMessage(msg.Chat, "/xkcd: Get the comic image from xkcd \nHere are some commands to try: \n/xkcd standards", so)
		return
	}

	quitRepeat := j.RepeatChatAction(msg, telebot.UploadingPhoto)
	defer func() { quitRepeat <- true }()

	results := make([]XKCDComic, 5)
	rawQuery := ""
	for _, v := range msg.Args {
		rawQuery = rawQuery + v + " "
	}

	for _, term := range strings.Fields(rawQuery) {
		term = strings.ToLower(term)
		type ScoreRecord struct {
			Index int
			Score int
		}

		scores := make([]ScoreRecord, len(comics))

		for i, comic := range comics {
			if comic.Number == 1190 || comic.Number == 1608 || comic.Number == 1037 || comic.Number == 1335 || comic.Number == 1663 {
				continue
			}

			if len(results) >= 5 {
				break
			}

			scores[i].Index = i
			scores[i].Score = titleWordWeight * stringSliceCount(term, comic.TitleFields)
			scores[i].Score += titleIntextWeight * strings.Count(strings.ToLower(comic.Title), term)

			lx := strings.Index(comic.Text, "==Transcript==")
			rx := len(comic.Text)
			if lx < 0 {
				lx = 0
			} else {
				rx = strings.Index(comic.Text[lx+14:], "==")
				if rx < 0 {
					rx = len(comic.Text)
				} else {
					rx += lx + 14
				}
			}

			scores[i].Score += transcriptWeight * strings.Count(comic.Text[lx:rx], term)
			scores[i].Score += textWordWeight * stringSliceCount(term, strings.Fields(strings.ToLower(comic.Text)))
			scores[i].Score += textWeight * strings.Count(strings.ToLower(comic.Text), term)
		}

		sort.Slice(scores, func(i, j int) bool {
			return scores[i].Score > scores[j].Score
		})

		results = append(results, comics[scores[0].Index])
	}
	if len(results) > 0 {
		if imgUrl, err := url.Parse(results[0].URL); err == nil {
			j.sendPhotoFromURL(imgUrl, msg)
			return
		}
	}
	j.SendMessage(msg.Chat, "My xkcd search returned nothing. \U0001F622", &telebot.SendOptions{ReplyTo: *msg.Message})

}

func (j *JarvisBot) CrawlXkcd(pwd string) {
	doc, _ := goquery.NewDocument(URL)

	tmpComics := make([]XKCDComic, 0)

	mux := &sync.Mutex{}

	var wg sync.WaitGroup
	doc.Find("tr").Each(func(i int, row *goquery.Selection) {
		wg.Add(1)

		go func() {
			defer wg.Done()

			comic := XKCDComic{}
			explanationURL := ""

			row.Find("td").Each(func(j int, col *goquery.Selection) {
				text := strings.TrimSpace(col.Text())

				switch j {
				case 0:
					comic.URL = text
					comic.Number, _ = strconv.Atoi(text[strings.Index(text, "/")+1:])

				case 1:
					comic.Title = strings.TrimSpace(text[:strings.Index(text, "(create)")-1])
					comic.TitleFields = strings.Fields(comic.Title)

					explanationURL, _ = col.Find("a").Attr("href")
					explanationURL = "http://www.explainxkcd.com" + explanationURL[:15] + "?action=edit&title=" + explanationURL[16:]

					exp, err := goquery.NewDocument(explanationURL)
					if err == nil {
						comic.Text = exp.Find("textarea").Text()
					}

				case 3:
					comic.Image = "https://imgs.xkcd.com/comics/" + strings.Replace(text, " ", "_", -1)

				case 4:
					comic.Date = text
				}
			})

			index := strings.Index(comic.Text, "titletext = ")
			if index > 0 {
				comic.TitleText = comic.Text[index+12:]
				comic.TitleText = comic.TitleText[:strings.Index(comic.TitleText, "}")-1]
			}

			mux.Lock()
			tmpComics = append(tmpComics, comic)
			mux.Unlock()
		}()
	})
	wg.Wait()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(tmpComics)

	f, _ := os.Create(path.Join(pwd, "comics.bin"))
	f.Write(buf.Bytes())
	f.Close()

	comics = tmpComics
}

func (j *JarvisBot) LoadComics(pwd string) {
	b, err := ioutil.ReadFile(path.Join(pwd, "comics.bin"))
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	comics = make([]XKCDComic, 0)
	dec.Decode(&comics)

	fmt.Printf("Loaded %d comics\n", len(comics))
}

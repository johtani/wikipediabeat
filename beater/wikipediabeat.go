package beater

import (
	"compress/bzip2"
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"errors"
	"github.com/dustin/go-wikiparse"
	"github.com/johtani/wikipediabeat/config"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Wikipediabeat struct {
	beatConfig *config.Config
	done       chan struct{}
	period     time.Duration
	name       string
}

var specialTitleRE, redirectCheckRE, categoryRE, nowikiRE, commentRE, fileRE *regexp.Regexp
var plainText1RE, plainText2RE, plainText3RE, plainText4RE, plainText5RE, plainText6RE *regexp.Regexp

func init() {
	specialTitleRE = regexp.MustCompile(`^Wikipedia:`)
	redirectCheckRE = regexp.MustCompile(`#REDIRECT\s+\[\[(.*?)\]\]`)
	categoryRE = regexp.MustCompile(`\[\[[Cc]ategory:([^\|\]]+)`)
	nowikiRE = regexp.MustCompile(`(?ms)<nowiki>.*</nowiki>`)
	commentRE = regexp.MustCompile(`(?ms)<!--.*-->`)
	fileRE = regexp.MustCompile(`\[ファイル:([^\|\]]+)`)
	plainText1RE = regexp.MustCompile(`<ref>(.*?)</ref>`)
	plainText2RE = regexp.MustCompile(`\{\{(.*?)\}\}`)
	plainText3RE = regexp.MustCompile(`\[\[(.*?):(.*?)\]\]`)
	plainText4RE = regexp.MustCompile(`\[\[(.*?)\]\]`)
	plainText5RE = regexp.MustCompile(`\s(.*?)\|(\w+\s)`)
	plainText6RE = regexp.MustCompile(`\[(.*?)\]`)
}

func findFiles(text string) []string {
	cleaned := nowikiRE.ReplaceAllString(commentRE.ReplaceAllString(text, ""), "")
	matches := fileRE.FindAllStringSubmatch(cleaned, 20)
	returnValue := []string{}
	for _, x := range matches {
		returnValue = append(returnValue, x[1])
	}
	return returnValue
}

func imageURL(text string) string {
	//TODO implemantation
	files := findFiles(text)
	image := ""
	if len(files) > 0 {

	}
	return image
}

func baseURL(text string) string {
	return text[0 : strings.LastIndex(text, "/")+1]
}

func findCategories(text string) []string {
	cleaned := nowikiRE.ReplaceAllString(commentRE.ReplaceAllString(text, ""), "")
	matches := categoryRE.FindAllStringSubmatch(cleaned, -1)
	returnValue := make([]string, 0, len(matches))
	for _, x := range matches {
		returnValue = append(returnValue, x[1])
	}
	return returnValue
}

func filteringLinks(text string) []string {
	plainLinks := wikiparse.FindLinks(text)
	var returnValue []string
	for _, value := range plainLinks {
		if !strings.Contains(value, ":") {
			returnValue = append(returnValue, value)
		}
	}
	return returnValue
}

func plainText(text string) string {
	plainText := text
	//plainText = strings.Replace(plainText, "\n", " ", -1)
	plainText = strings.Replace(plainText, "&gt;", ">", -1)
	plainText = strings.Replace(plainText, "&lt;", "<", -1)
	plainText = plainText1RE.ReplaceAllString(plainText, " ")
	plainText = plainText2RE.ReplaceAllString(plainText, " ")
	plainText = plainText3RE.ReplaceAllString(plainText, " ")
	plainText = plainText4RE.ReplaceAllString(plainText, "$1")
	plainText = plainText5RE.ReplaceAllString(plainText, "$2")
	plainText = plainText6RE.ReplaceAllString(plainText, " ")
	plainText = strings.Replace(plainText, "'", "", -1)
	return plainText
}

// Creates beater
func New() *Wikipediabeat {
	return &Wikipediabeat{
		done: make(chan struct{}),
	}
}

/// *** Beater interface methods ***///

func (bt *Wikipediabeat) Config(b *beat.Beat) error {

	// Load beater beatConfig
	err := cfgfile.Read(&bt.beatConfig, "")
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	return nil
}

func (bt *Wikipediabeat) Setup(b *beat.Beat) error {

	name := bt.beatConfig.Wikipediabeat.Name
	if name == "" {
		return errors.New("no name in config file")
	}

	bt.name = name

	return nil
}

func (bt *Wikipediabeat) Run(b *beat.Beat) error {
	logp.Info("wikipediabeat is running! Hit CTRL-C to stop it.")

	//read file
	f, err := os.Open(bt.name)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	z := bzip2.NewReader(f)

	p, err := wikiparse.NewParser(z)
	if err != nil {
		log.Fatalf("Error setting up new page parser: %v", err)
	}

	logp.Info("Got site info: %+v", p.SiteInfo())

	//parse each pages
	counter := 1
	baseURL := baseURL(p.SiteInfo().Base)

	for err == nil {
		var page *wikiparse.Page
		page, err = p.Next()
		//if needed, filtering to index or not
		if !specialTitleRE.MatchString(page.Title) && len(page.Revisions) > 0 {
			pageTime, err2 := time.Parse("2006-01-02T15:04:05Z", page.Revisions[0].Timestamp)
			if err == nil && err2 == nil {
				event := common.MapStr{
					"@timestamp": common.Time(time.Now()),
					"updated":    common.Time(pageTime),
					"type":       b.Name,
					"title":      page.Title,
					"text":       plainText(page.Revisions[0].Text),
					"category":   findCategories(page.Revisions[0].Text),
					"link":       filteringLinks(page.Revisions[0].Text),
					"url":        baseURL + url.QueryEscape(strings.Replace(page.Title, " ", "_", -1)),
					"image":      imageURL(page.Revisions[0].Text),
				}
				b.Events.PublishEvent(event)
				logp.Info("Event sent")
			} else if err2 != nil {
				err = err2
			}
		}
		counter++
	}
	logp.Info("wikipediabeat ended parse %l documents", counter)
	return err
}

func (bt *Wikipediabeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (bt *Wikipediabeat) Stop() {
	close(bt.done)
}

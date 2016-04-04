package beater

import (
	"compress/bzip2"
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/johtani/wikipediabeat/config"
	"github.com/dustin/go-wikiparse"
	"errors"
	"os"
	"log"
	"regexp"
	"strings"
)

type Wikipediabeat struct {
	beatConfig *config.Config
	done       chan struct{}
	period     time.Duration
	name       string
}

var specialTitleRE, redirectCheckRE, categoryRE, nowikiRE, commentRE *regexp.Regexp
var plainText1RE, plainText2RE, plainText3RE, plainText4RE, plainText5RE, plainText6RE *regexp.Regexp

func init() {
	specialTitleRE = regexp.MustCompile(`^Wikipedia:`)
	redirectCheckRE = regexp.MustCompile(`#REDIRECT\s+\[\[(.*?)\]\]`)
	categoryRE = regexp.MustCompile(`\[\[[Cc]ategory:([^\|\]]+)`)
	nowikiRE = regexp.MustCompile(`(?ms)<nowiki>.*</nowiki>`)
	commentRE = regexp.MustCompile(`(?ms)<!--.*-->`)
	plainText1RE = regexp.MustCompile(`<ref>(.*?)</ref>`)
	plainText2RE = regexp.MustCompile(`\{\{(.*?)\}\}`)
	plainText3RE = regexp.MustCompile(`\[\[(.*?):(.*?)\]\]`)
	plainText4RE = regexp.MustCompile(`\[\[(.*?)\]\]`)
	plainText5RE = regexp.MustCompile(`\s(.*?)\|(\w+\s)`)
	plainText6RE = regexp.MustCompile(`\[(.*?)\]`)
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

	for err == nil && counter < 5 {
		var page *wikiparse.Page
		page, err = p.Next()
		//if needed, filtering to index or not
		if !specialTitleRE.MatchString(page.Title) {
			if err == nil {
				event := common.MapStr{
					"@timestamp": page.Revisions[0].Timestamp,
					"type": b.Name,
					"title": page.Title,
					"text": plainText(page.Revisions[0].Text),
					"category": findCategories(page.Revisions[0].Text),
					"link": filteringLinks(page.Revisions[0].Text),
					"": nil,
				}

				b.Events.PublishEvent(event)
				logp.Info("Event sent")
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

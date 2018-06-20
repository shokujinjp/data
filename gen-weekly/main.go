package main

import (
	"context"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/oauth2/google"

	"github.com/ChimeraCoder/anaconda"
	"github.com/gocarina/gocsv"
	vision "google.golang.org/api/vision/v1"
)

const (
	createAtFormat = "Mon Jan 02 15:04:05 -0700 2006"
	dayFormat      = "2006-01-02"
	idFormat       = "20060102"
)

var (
	re = regexp.MustCompile(`(9|15)\.(.*?)(\d+)円`)
)

type Record struct {
	Id          string `csv:"id"`
	Name        string `csv:"name"`
	Price       string `csv:"price"`
	Category    string `csv:"category"`
	DayStart    string `csv:"day_start"`
	DayEnd      string `csv:"day_end"`
	CanWeekday  string `csv:"can_weekday"`
	Description string `csv:"description"`
}

func (r *Record) MarshalString() string {
	return r.Id + "," +
		r.Name + "," +
		r.Price + "," +
		r.Category + "," +
		r.DayStart + "," +
		r.DayEnd + "," +
		r.CanWeekday + "," +
		r.Description

}

func initialize() (*vision.Service, *anaconda.TwitterApi, error) {
	// create vision service
	saJson := os.Getenv("SA_JSON")

	vcfg, err := google.JWTConfigFromJSON(
		[]byte(saJson), vision.CloudPlatformScope)
	if err != nil {
		return nil, nil, err
	}

	vclient := vcfg.Client(context.Background())

	svc, err := vision.New(vclient)
	if err != nil {
		return nil, nil, err
	}

	// create twitter client
	api := anaconda.NewTwitterApiWithCredentials(os.Getenv("TW_AT"), os.Getenv("TW_ATS"), os.Getenv("TW_CK"), os.Getenv("TW_CS"))

	return svc, api, nil
}

func getNewestTweet(api *anaconda.TwitterApi) (anaconda.Tweet, error) {
	searchResult, err := api.GetSearch("今週の週代わり定食 from:shokujinjp", nil)
	if err != nil {
		return anaconda.Tweet{}, err
	}

	// get newest tweet
	tweet := searchResult.Statuses[0]
	return tweet, nil
}

func doVisionRequest(svc *vision.Service, imageURL string) (*vision.BatchAnnotateImagesResponse, error) {
	imgSource := &vision.ImageSource{
		ImageUri: imageURL,
	}
	img := &vision.Image{Source: imgSource}
	feature := &vision.Feature{
		Type:       "DOCUMENT_TEXT_DETECTION",
		MaxResults: 10,
	}
	req := &vision.AnnotateImageRequest{
		Image:    img,
		Features: []*vision.Feature{feature},
	}
	batch := &vision.BatchAnnotateImagesRequest{
		Requests: []*vision.AnnotateImageRequest{req},
	}
	res, err := svc.Images.Annotate(batch).Do()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func main() {
	visionSvc, api, err := initialize()
	if err != nil {
		log.Fatal(err)
	}

	tweet, err := getNewestTweet(api)
	if err != nil {
		log.Fatal(err)
	}
	t, err := time.Parse(createAtFormat, tweet.CreatedAt)
	if err != nil {
		log.Fatal(err)
	}
	csvFileName := "../" + t.Format(idFormat) + ".csv"
	if Exists(csvFileName) {
		log.Printf("already done: %s\n", csvFileName)
		os.Exit(0)
	}

	res, err := doVisionRequest(visionSvc, tweet.Entities.Media[0].Media_url_https)
	if err != nil {
		log.Fatal(err)
	}
	rawText := res.Responses[0].FullTextAnnotation.Text

	var oneline string
	for _, s := range rawText {
		o := string([]rune{s})
		oneline += strings.TrimSpace(o)
	}
	slice915 := re.FindAllStringSubmatch(oneline, -1)

	menu9 := Record{
		Id:          idFormat + "09",
		Name:        slice915[0][2],
		Price:       slice915[0][3],
		Category:    "定食",
		Description: "週代わり定食9番",
		DayStart:    t.Format(dayFormat),
		DayEnd:      t.AddDate(0, 0, 6).Format(dayFormat),
	}
	menu15 := Record{
		Id:          idFormat + "15",
		Name:        slice915[1][2],
		Price:       slice915[1][3],
		Category:    "定食",
		Description: "週代わり定食15番",
		DayStart:    t.Format(dayFormat),
		DayEnd:      t.AddDate(0, 0, 6).Format(dayFormat),
	}
	var menu []Record
	menu = append(menu, menu9)
	menu = append(menu, menu15)

	file, _ := os.OpenFile(csvFileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	defer file.Close()
	gocsv.MarshalFile(&menu, file)
}

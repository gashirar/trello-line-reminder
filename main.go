package main

import (
	"fmt"
	"github.com/adlio/trello"
	"github.com/domnikl/ifttt-webhook"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sort"
	"time"
)

var trelloApiKey string
var trelloToken string
var trelloListId string
var iftttWebhookKey string

func main()  {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	trelloApiKey = os.Getenv("TRELLO_API_KEY")
	trelloToken = os.Getenv("TRELLO_TOKEN")
	trelloListId = os.Getenv("TRELLO_LIST_ID")
	iftttWebhookKey = os.Getenv("IFTTT_WEBHOOK_KEY")


	client := trello.NewClient(trelloApiKey, trelloToken)

	list, err := client.GetList(trelloListId, trello.Defaults())
	if err != nil {
		log.Fatal(err)
	}

	cards, err := list.GetCards(trello.Defaults())
	if err != nil {
		log.Fatal(err)
	}

	cards = filter(cards, func(card *trello.Card) bool { return card.Due != nil})
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].Due != nil {
			return cards[i].Due.Before(*cards[j].Due)
		}
		return false
	})

	tomorrow := utc2jst(time.Now()).AddDate(0, 0, 1).Day()
	for _, card := range cards {
		jstDue := utc2jst(*card.Due)
		log.Printf("%02d月%02d日%02d時%02d分 : %s\n", jstDue.Month(), jstDue.Day(), jstDue.Hour(), jstDue.Minute(), card.Name)
		if tomorrow == jstDue.Day() {
			i := iftttWebhook.New(iftttWebhookKey)
			title := card.Name
			startTime := fmt.Sprintf("%02d時%02d分", jstDue.Hour(), jstDue.Minute())
			err := i.Emit("reminder", title, startTime, "")
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func filter(cards []*trello.Card, f func(*trello.Card) bool) []*trello.Card {
	filteredCards := make([]*trello.Card, 0)
	for _, card := range cards {
		if f(card) {
			filteredCards = append(filteredCards, card)
		}
	}
	return filteredCards
}

func utc2jst(t time.Time) time.Time {
	nowUTC := t.UTC()
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)

	return nowUTC.In(jst)
}

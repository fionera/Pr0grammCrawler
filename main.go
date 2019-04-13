package main

import (
	"github.com/cenkalti/backoff"
	"github.com/fionera/go-pr0gramm"
	"github.com/sirupsen/logrus"
	"net/http"
)

var session *pr0gramm.Session
var itemChan chan pr0gramm.Item

func main() {
	session = pr0gramm.NewSession(&http.Client{})
	itemChan = make(chan pr0gramm.Item, 10)

	go StartRequestLoop()

	for item := range itemChan {
		// Here you can work with all posts
		logrus.Println(item.Id)
	}
}

func StartRequestLoop() {
	err := backoff.Retry(func() error {
		var smallestId = pr0gramm.Id(0)
		var lastCrawl pr0gramm.Id
		var err error

		for {
			lastCrawl, err = RequestItems(smallestId)

			if lastCrawl == 1 {
				break
			} else {
				smallestId = lastCrawl
			}
		}

		return err
	}, backoff.NewExponentialBackOff())

	if err != nil {
		logrus.Panic(err)
	}

	close(itemChan)
}

func RequestItems(older pr0gramm.Id) (smallestId pr0gramm.Id, err error) {
	items, err := session.GetItems(&pr0gramm.ItemsRequest{
		Older:        pr0gramm.Id(older),
		ContentTypes: pr0gramm.ContentTypes{pr0gramm.SFW},
	})
	if err != nil {
		return 0, err
	}

	for _, item := range items.Items {
		if item.Id < smallestId || smallestId == 0 {
			smallestId = item.Id
		}

		itemChan <- item
	}

	return smallestId, nil
}

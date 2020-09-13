package worker

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/BarTar213/bartlomiej-tarczynski/models"
	"github.com/BarTar213/bartlomiej-tarczynski/storage"
	"github.com/robfig/cron/v3"
)

type Worker struct {
	c           *cron.Cron
	storage     storage.Storage
	client      *http.Client
	historyPool *sync.Pool
}

func New(storage storage.Storage, historyPool *sync.Pool) *Worker {
	c := cron.New(cron.WithSeconds())
	c.Start()
	return &Worker{
		c:           c,
		storage:     storage,
		historyPool: historyPool,
		client:      http.DefaultClient,
	}
}

func (w *Worker) RegisterJob(fetcher *models.Fetcher) (int, error) {
	entryID, err := w.c.AddFunc(fmt.Sprintf("@every %ds", fetcher.Interval), func() {
		w.doJob(fetcher.Url)
	})
	if err != nil {
		return 0, err
	}

	return int(entryID), nil
}

func (w *Worker) DeregisterJob(id int) {
	w.c.Remove(cron.EntryID(id))
}

func (w *Worker) doJob(url string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	history := w.historyPool.Get().(*models.History)
	defer w.ReturnHistoryItem(history)

	t := time.Now()
	response, err := w.client.Do(req)
	duration := time.Since(t).Seconds()
	if err == nil {
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err == nil {
			history.Response = pointer(string(body))
		}
	}

	history.CreatedAt = t.Unix()
	history.Duration = duration

	err = w.storage.AddHistory(history)
	if err != nil {
		return
	}
}

func (w *Worker) ReturnHistoryItem(h *models.History) {
	h.Reset()
	w.historyPool.Put(h)
}

func pointer(s string) *string {
	return &s
}
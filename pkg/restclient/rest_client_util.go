package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"

	"github.com/sarumaj/gh-pr/pkg/config"
)

var linkRegex = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)

func getPaged[T any](c RESTClient, ep apiEndpoint, ctx context.Context) (result []T, err error) {
	resp, err := c.doRequest(
		ctx,
		http.MethodGet,
		newRequestPath(ep).
			Add("per_page", "100").
			Add("page", "1").
			String(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	paged, err := unmarshalListHead[T](resp)
	if err != nil {
		return nil, err
	}

	result = append(result, paged...)

	var wg sync.WaitGroup
	var mutex sync.Mutex
	for page, last := 2, getLastPage(resp.Header); page <= last; page++ {
		wg.Add(1)
		go func(result *[]T, err *error, page int) {
			defer wg.Done()

			var paged []T
			*err = c.DoWithContext(
				ctx,
				http.MethodGet,
				newRequestPath(ep).
					Add("per_page", "100").
					Add("page", fmt.Sprintf("%d", page)).
					String(),
				nil,
				&paged,
			)
			if *err != nil || len(paged) == 0 {
				return
			}

			for !mutex.TryLock() {
			}
			*result = append(*result, paged...)
			mutex.Unlock()
		}(&result, &err, page)
		c.Debugf("Dispatched request for page %d", page)

		if (page-1)%config.Concurrency == 0 || page == last {
			c.Debug("Waiting for coroutines to finish")
			wg.Wait()
		}
	}

	return
}

func getLastPage(responseHeader http.Header) (limit int) {
	for _, m := range linkRegex.FindAllStringSubmatch(responseHeader.Get("Link"), -1) {
		if len(m) <= 2 || m[2] != "last" {
			continue
		}

		parsed, err := url.Parse(m[1])
		if err != nil {
			return
		}

		page, err := strconv.Atoi(parsed.Query().Get("page"))
		if err != nil {
			return
		}

		return page
	}
	return
}

func unmarshalListHead[T any](response *http.Response) ([]T, error) {
	if response.Body == nil {
		return nil, nil
	}
	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var head []T
	err = json.Unmarshal(b, &head)
	if err != nil {
		return nil, err
	}

	return head, nil
}

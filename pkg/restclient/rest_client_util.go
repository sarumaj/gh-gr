package restclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/sarumaj/gh-gr/pkg/util"
	"gopkg.in/go-playground/pool.v3"
)

var linkRegex = regexp.MustCompile(`<(?P<Link>[^>]+)>;\s*rel="(?P<Type>[^"]+)"`)

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

	p := pool.NewLimited(c.Concurrency)
	defer p.Close()

	batch := p.Batch()

	logger := util.Logger()
	go func() {
		for page, last := 2, getLastPage(resp.Header); page <= last; page++ {
			logger.Debugf("Dispatching request for page %d", page)
			batch.Queue(getPagedWorkUnit[T](c, ep, ctx, page))
		}

		batch.QueueComplete()
	}()

	for items := range batch.Results() {
		if err := items.Error(); err != nil {
			return nil, err
		}
		result = append(result, items.Value().([]T)...)
	}

	return
}

func getPagedWorkUnit[T any](c RESTClient, ep apiEndpoint, ctx context.Context, page int) pool.WorkFunc {
	logger := util.Logger()
	return func(wu pool.WorkUnit) (any, error) {
		var paged []T
		err := c.DoWithContext(
			ctx,
			http.MethodGet,
			newRequestPath(ep).
				Add("per_page", "100").
				Add("page", fmt.Sprintf("%d", page)).
				String(),
			nil,
			&paged,
		)

		if wu.IsCancelled() {
			logger.Warn("work unit has been prematurely canceled")
			return nil, nil
		}

		if err != nil || len(paged) == 0 {
			logger.Error(err)
			return nil, err
		}

		return paged, nil
	}
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

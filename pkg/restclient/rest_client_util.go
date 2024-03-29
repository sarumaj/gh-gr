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

	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	pool "gopkg.in/go-playground/pool.v3"
)

// Regular expression used to extract last page information from response header.
var lastPageLinkRegex = regexp.MustCompile(`<(?P<Link>[^>]+)>;\s*rel="last"`)

// Send HTTP request to fetch first page and retrieve number of pages.
func getLastPage(responseHeader http.Header) (limit int) {
	match := lastPageLinkRegex.FindStringSubmatch(responseHeader.Get("Link"))
	if len(match) <= 1 {
		return
	}

	parsed, err := url.Parse(match[1])
	if err != nil {
		return
	}

	page, err := strconv.Atoi(parsed.Query().Get("page"))
	if err != nil {
		return
	}

	return page
}

// Retrieve all elements through paginated requests.
func getPaged[T any](c RESTClient, ep apiEndpoint, ctx context.Context) (result []T, err error) {
	resp, err := c.RequestWithContext(
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

	go func() {
		last := getLastPage(resp.Header)
		_ = c.ChangeMax(last)
		for page := 2; page <= last; page++ {
			util.Logger.Debugf("Dispatching request for page %d", page)
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

// Worker to send paginated requests.
func getPagedWorkUnit[T any](c RESTClient, ep apiEndpoint, ctx context.Context, page int) pool.WorkFunc {
	return func(wu pool.WorkUnit) (any, error) {
		defer c.Inc()

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
			util.Logger.Warn("work unit has been prematurely canceled")
			return nil, nil
		}

		if err != nil {
			util.Logger.Error(err)
			return nil, err
		}

		return paged, nil
	}
}

// Unmarshal first page of paginated response.
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

package restclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/sarumaj/gh-gr/v2/pkg/restclient/resources"
	util "github.com/sarumaj/gh-gr/v2/pkg/util"
	pool "gopkg.in/go-playground/pool.v3"
)

// Regular expression used to extract last page information from response header.
var lastPageLinkRegex = regexp.MustCompile(`<(?P<Link>[^>]+)>;\s*rel="last"`)

// Consolidate two paged results.
func consolidate[T any, R interface {
	[]T | resources.SearchResult[T]
}](target, source R) R {
	{
		s, sOK := any(source).(*resources.SearchResult[T])
		t, tOK := any(target).(*resources.SearchResult[T])
		if sOK && tOK {
			t.Items = append(t.Items, s.Items...)
			target = any(t).(R)
			return target
		}
	}

	{
		s, sOK := any(source).([]T)
		t, tOK := any(target).([]T)
		if sOK && tOK {
			result := append(t, s...)
			target = any(result).(R)
			return target
		}
	}

	return target
}

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
func getPaged[T any, R interface {
	[]T | resources.SearchResult[T]
}](c *RESTClient, ep apiEndpoint, ctx context.Context, options ...func(*requestPath)) (result R, err error) {
	params := newRequestPath(ep).
		Add("per_page", "100").
		Add("page", "1")
	for _, option := range options {
		option(params)
	}

	err = params.Validate()
	if err != nil {
		return
	}

	var resp *http.Response
	resp, err = c.RequestWithContext(
		ctx,
		http.MethodGet,
		params.String(),
		nil,
	)
	if err != nil {
		return
	}

	var paged R
	paged, err = unmarshalListHead[T, R](resp)
	if err != nil {
		return
	}

	result = consolidate[T](result, paged)

	p := pool.NewLimited(c.Concurrency)
	defer p.Close()

	batch := p.Batch()

	go func() {
		last := getLastPage(resp.Header)
		_ = c.ChangeMax(last)
		for page := 2; page <= last; page++ {
			util.Logger.Debugf("Dispatching request for page %d", page)
			batch.Queue(getPagedWorkUnit[T, R](c, ep, ctx, page))
		}

		batch.QueueComplete()
	}()

	for items := range batch.Results() {
		err = items.Error()
		if err != nil {
			return
		}

		result = consolidate[T](result, items.Value().(R))
	}

	return
}

// Worker to send paginated requests.
func getPagedWorkUnit[T any, R interface {
	[]T | resources.SearchResult[T]
}](c *RESTClient, ep apiEndpoint, ctx context.Context, page int) pool.WorkFunc {
	return func(wu pool.WorkUnit) (any, error) {
		defer c.Inc()

		var paged R
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
func unmarshalListHead[T any, R interface {
	[]T | resources.SearchResult[T]
}](response *http.Response) (head R, err error) {
	if response.Body == nil {
		return
	}

	err = json.NewDecoder(response.Body).Decode(&head)
	return
}

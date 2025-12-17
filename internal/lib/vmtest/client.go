package vmtest

import (
	"context"
	"fmt"

	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/video-manager-api/go/vmapi"
)

type Response[RawT any, HappyT any] interface {
	Raw() (*RawT, error)
	Get() (*HappyT, error)
	Must(exam.E, deep.Env) *HappyT
}

type VoidResponse[RawT any] interface {
	Raw() (*RawT, error)
	Error() error
	Must(exam.E, deep.Env)
}

type PagedResponse[RawT any, PageT any, HappyT any] interface {
	Raw() ([]*RawT, error)
	GetPages() ([]*PageT, error)
	MustPages(exam.E, deep.Env) []*PageT
	Get() ([]*HappyT, error)
	Must(exam.E, deep.Env) []*HappyT
}

// Response Implementation
// -----------------------
type getHappyFunc[RawT any, HappyT any] func(*RawT, error) (*HappyT, error)

type responseImpl[RawT any, HappyT any] struct {
	raw      *RawT
	err      error
	getHappy getHappyFunc[RawT, HappyT]
}

func (r *responseImpl[RawT, HappyT]) Raw() (*RawT, error) {
	return r.raw, r.err
}

func (r *responseImpl[RawT, HappyT]) Get() (*HappyT, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.getHappy(r.raw, r.err)
}

func (r *responseImpl[RawT, HappyT]) Must(e exam.E, env deep.Env) *HappyT {
	e.Helper()
	happy, err := r.Get()
	exam.Nil(e, env, err).Must()
	return happy
}

func newResponse[RawT any, HappyT any](raw *RawT, err error, getHappy getHappyFunc[RawT, HappyT]) Response[RawT, HappyT] {
	return &responseImpl[RawT, HappyT]{
		raw:      raw,
		err:      err,
		getHappy: getHappy,
	}
}

// Void Response Implementation
// ----------------------------
type checkAppErrorFunc[RawT any] func(*RawT, error) error

type voidResponseImpl[RawT any] struct {
	raw           *RawT
	err           error
	checkAppError checkAppErrorFunc[RawT]
}

func (r *voidResponseImpl[RawT]) Raw() (*RawT, error) {
	return r.raw, r.err
}

func (r *voidResponseImpl[RawT]) Error() error {
	return r.checkAppError(r.raw, r.err)
}

func (r *voidResponseImpl[RawT]) Must(e exam.E, env deep.Env) {
	e.Helper()
	exam.Nil(e, env, r.Error()).Must()
}

func newVoidResponse[RawT any](raw *RawT, err error, checkAppError checkAppErrorFunc[RawT]) VoidResponse[RawT] {
	return &voidResponseImpl[RawT]{
		raw:           raw,
		err:           err,
		checkAppError: checkAppError,
	}
}

type fetchPageFunc[RawT any] func(token *string, pageSize *uint32) (*RawT, error)
type readPageFunc[RawT any, PageT any, HappyT any] func(raw *RawT, err error) (*PageT, []HappyT, string, error)

// Paged Response Implementation
// -----------------------------
type pageReader[RawT any, PageT any, HappyT any] struct {
	fetchPage fetchPageFunc[RawT]
	readPage  readPageFunc[RawT, PageT, HappyT]
}

func newPageReader[RawT any, PageT any, HappyT any](fetch fetchPageFunc[RawT], read readPageFunc[RawT, PageT, HappyT]) *pageReader[RawT, PageT, HappyT] {
	return &pageReader[RawT, PageT, HappyT]{
		fetchPage: fetch,
		readPage:  read,
	}
}

func (pr *pageReader[RawT, PageT, HappyT]) ReadAll(firstToken *string, pageSize *uint32) PagedResponse[RawT, PageT, HappyT] {
	var token string
	if firstToken != nil {
		token = *firstToken
	}
	var allRaw []*RawT
	var allPages []*PageT
	var allHappy []*HappyT
	for {
		raw, err := pr.fetchPage(&token, pageSize)
		page, happy, nextToken, err := pr.readPage(raw, err)
		if err != nil {
			return &pagedResponseImpl[RawT, PageT, HappyT]{
				raw:   allRaw,
				pages: allPages,
				happy: allHappy,
				err:   err,
			}
		}
		allRaw = append(allRaw, raw)
		allPages = append(allPages, page)
		for _, h := range happy {
			allHappy = append(allHappy, &h)
		}
		if nextToken == "" {
			break
		}
		token = nextToken
	}
	return &pagedResponseImpl[RawT, PageT, HappyT]{
		raw:   allRaw,
		pages: allPages,
		happy: allHappy,
		err:   nil,
	}
}

type pagedResponseImpl[RawT any, PageT any, HappyT any] struct {
	raw   []*RawT
	err   error
	pages []*PageT
	happy []*HappyT
}

func (r *pagedResponseImpl[RawT, PageT, HappyT]) Raw() ([]*RawT, error) {
	return r.raw, r.err
}

func (r *pagedResponseImpl[RawT, PageT, HappyT]) GetPages() ([]*PageT, error) {
	return r.pages, r.err
}

func (r *pagedResponseImpl[RawT, PageT, HappyT]) MustPages(e exam.E, env deep.Env) []*PageT {
	e.Helper()
	exam.Nil(e, env, r.err).Must()
	return r.pages
}

func (r *pagedResponseImpl[RawT, PageT, HappyT]) Get() ([]*HappyT, error) {
	return r.happy, r.err
}

func (r *pagedResponseImpl[RawT, PageT, HappyT]) Must(e exam.E, env deep.Env) []*HappyT {
	e.Helper()
	exam.Nil(e, env, r.err).Must()
	return r.happy
}

// Page Size
// ---------
type PageSizeOption interface {
	ListCardsOption
}

type pageSizeOptionImpl uint32

func (p pageSizeOptionImpl) applyListCards(params *vmapi.ListCardsParams) {
	size := uint32(p)
	params.PageSize = &size
}

func WithPageSize(size int32) PageSizeOption {
	return pageSizeOptionImpl(size)
}

// Page Token
// ----------
type PageTokenOption interface {
	ListCardsOption
}

type pageTokenOptionImpl string

func (p pageTokenOptionImpl) applyListCards(params *vmapi.ListCardsParams) {
	token := string(p)
	params.PageToken = &token
}

func WithPageToken(token string) PageTokenOption {
	return pageTokenOptionImpl(token)
}

// Name
// ----
type NameOption interface {
	PostCardOption
	PatchCardOption
}

type nameOptionImpl string

func (n nameOptionImpl) applyPostCard(body *vmapi.PostCardJSONRequestBody) {
	name := string(n)
	body.Name = name
}

func (n nameOptionImpl) applyPatchCard(body *vmapi.CardPatch) {
	name := string(n)
	body.Name = &name
}

func WithName(name string) NameOption {
	return nameOptionImpl(name)
}

// Options for each method that can accept them
// --------------------------------------------

type ListCardsOption interface {
	applyListCards(*vmapi.ListCardsParams)
}

type PostCardOption interface {
	applyPostCard(*vmapi.PostCardJSONRequestBody)
}

type PatchCardOption interface {
	applyPatchCard(*vmapi.CardPatch)
}

// Client Interface
// ----------------
type Client interface {
	// Cards
	ListCards(...ListCardsOption) PagedResponse[vmapi.ListCardsResponse, vmapi.CardPage, vmapi.Card]
	PostCard(...PostCardOption) Response[vmapi.PostCardResponse, vmapi.Card]
	DeleteCard(uint32) VoidResponse[vmapi.DeleteCardResponse]
	GetCard(uint32) Response[vmapi.GetCardResponse, vmapi.Card]
	PatchCard(uint32, ...PatchCardOption) Response[vmapi.PatchCardResponse, vmapi.Card]
}

// Client Implementation
// ---------------------
type clientImpl struct {
	ctx    context.Context
	client vmapi.ClientWithResponsesInterface
}

func handleResponseErrors(err error, resp *vmapi.ErrorResponse) error {
	if err != nil {
		return err
	}
	if resp == nil {
		return nil
	}
	return fmt.Errorf("application error: %v", resp)
}

func (c *clientImpl) ListCards(opts ...ListCardsOption) PagedResponse[vmapi.ListCardsResponse, vmapi.CardPage, vmapi.Card] {
	params := &vmapi.ListCardsParams{}
	for _, opt := range opts {
		opt.applyListCards(params)
	}
	fetch := func(token *string, pageSize *uint32) (*vmapi.ListCardsResponse, error) {
		params.PageToken = token
		params.PageSize = pageSize
		return c.client.ListCardsWithResponse(c.ctx, params)
	}
	read := func(raw *vmapi.ListCardsResponse, err error) (*vmapi.CardPage, []vmapi.Card, string, error) {
		if err := handleResponseErrors(err, raw.ApplicationproblemJSONDefault); err != nil {
			return nil, nil, "", err
		}
		nextToken := ""
		if raw.JSON200.NextPageToken != nil {
			nextToken = *raw.JSON200.NextPageToken
		}
		return raw.JSON200, raw.JSON200.Cards, nextToken, nil
	}
	pr := newPageReader(fetch, read)
	return pr.ReadAll(params.PageToken, params.PageSize)
}

func (c *clientImpl) PostCard(opts ...PostCardOption) Response[vmapi.PostCardResponse, vmapi.Card] {
	body := vmapi.PostCardJSONRequestBody{}
	for _, opt := range opts {
		opt.applyPostCard(&body)
	}
	rawResp, err := c.client.PostCardWithResponse(c.ctx, body)
	getHappy := func(raw *vmapi.PostCardResponse, err error) (*vmapi.Card, error) {
		if err := handleResponseErrors(err, raw.ApplicationproblemJSONDefault); err != nil {
			return nil, err
		}
		return raw.JSON201, nil
	}
	return newResponse(rawResp, err, getHappy)
}

func (c *clientImpl) DeleteCard(cardID uint32) VoidResponse[vmapi.DeleteCardResponse] {
	rawResp, err := c.client.DeleteCardWithResponse(c.ctx, cardID)
	check := func(raw *vmapi.DeleteCardResponse, err error) error {
		return handleResponseErrors(err, raw.ApplicationproblemJSONDefault)
	}
	return newVoidResponse(rawResp, err, check)
}

func (c *clientImpl) GetCard(cardID uint32) Response[vmapi.GetCardResponse, vmapi.Card] {
	rawResp, err := c.client.GetCardWithResponse(c.ctx, cardID)
	getHappy := func(raw *vmapi.GetCardResponse, err error) (*vmapi.Card, error) {
		if err := handleResponseErrors(err, raw.ApplicationproblemJSONDefault); err != nil {
			return nil, err
		}
		return raw.JSON200, nil
	}
	return newResponse(rawResp, err, getHappy)
}

func (c *clientImpl) PatchCard(cardID uint32, opts ...PatchCardOption) Response[vmapi.PatchCardResponse, vmapi.Card] {
	patches := make([]vmapi.CardPatch, 0, len(opts))
	for _, opt := range opts {
		patch := vmapi.CardPatch{}
		opt.applyPatchCard(&patch)
		patches = append(patches, patch)
	}
	rawResp, err := c.client.PatchCardWithResponse(c.ctx, cardID, patches)
	getHappy := func(raw *vmapi.PatchCardResponse, err error) (*vmapi.Card, error) {
		if err := handleResponseErrors(err, raw.ApplicationproblemJSONDefault); err != nil {
			return nil, err
		}
		return raw.JSON200, nil
	}
	return newResponse(rawResp, err, getHappy)
}

// Client Constructor
func NewClient(ctx context.Context, client vmapi.ClientWithResponsesInterface) Client {
	return &clientImpl{
		ctx:    ctx,
		client: client,
	}
}

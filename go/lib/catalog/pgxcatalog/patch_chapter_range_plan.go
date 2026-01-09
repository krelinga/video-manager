package pgxcatalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PatchChapterRangePlan(ctx context.Context, planUUID uuid.UUID) catalog.ChapterRangePlanPatcher {
	return &chapterRangePlanPatcher{
		Ctx:      ctx,
		Pool:     c.Pool,
		PlanUUID: planUUID,
	}
}

type chapterRangePlanPatcher struct {
	Ctx      context.Context
	Pool     *pgxpool.Pool
	PlanUUID uuid.UUID

	fileSourceUUID patchReqField[uuid.UUID]
	workUUID       patchReqField[uuid.UUID]
	startChapter   patchOptField[int]
	endChapter     patchOptField[int]
}

func (crpp *chapterRangePlanPatcher) SetFileSourceUUID(uuid uuid.UUID) catalog.ChapterRangePlanPatcher {
	crpp.fileSourceUUID.Set(uuid)
	return crpp
}

func (crpp *chapterRangePlanPatcher) SetWorkUUID(uuid uuid.UUID) catalog.ChapterRangePlanPatcher {
	crpp.workUUID.Set(uuid)
	return crpp
}

func (crpp *chapterRangePlanPatcher) SetStartChapter(chapter int) catalog.ChapterRangePlanPatcher {
	crpp.startChapter.Set(chapter)
	return crpp
}

func (crpp *chapterRangePlanPatcher) ClearStartChapter() catalog.ChapterRangePlanPatcher {
	crpp.startChapter.Clear()
	return crpp
}

func (crpp *chapterRangePlanPatcher) SetEndChapter(chapter int) catalog.ChapterRangePlanPatcher {
	crpp.endChapter.Set(chapter)
	return crpp
}

func (crpp *chapterRangePlanPatcher) ClearEndChapter() catalog.ChapterRangePlanPatcher {
	crpp.endChapter.Clear()
	return crpp
}

func (crpp *chapterRangePlanPatcher) SaveGet() (*catalog.Plan, error) {
	txn, err := crpp.Pool.Begin(crpp.Ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to begin transaction for patching plan %s: %w", catalog.ErrInternal, crpp.PlanUUID, err)
	}
	defer txn.Rollback(crpp.Ctx)

	row := txn.QueryRow(crpp.Ctx, `
		SELECT
			kind,
			body
		FROM cat.plans
		WHERE uuid = $1
	`, crpp.PlanUUID)
	var kind planKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query plan %s for patching: %w", catalog.ErrInternal, crpp.PlanUUID, err)
	} else if kind != planKindChapterRange {
		return nil, fmt.Errorf("%w: plan %s is not a chapter range plan", catalog.ErrKind, crpp.PlanUUID)
	}

	body := &chapterRangePlanJSON{}
	if err := body.UnmarshalJSON(rawBody); err != nil {
		return nil, err
	}

	if crpp.fileSourceUUID.Changed() {
		body.FileSourceUUID = crpp.fileSourceUUID.Get()
	}
	if crpp.workUUID.Changed() {
		body.WorkUUID = crpp.workUUID.Get()
	}
	if crpp.startChapter.Changed() {
		body.StartChapter = crpp.startChapter.Get()
	}
	if crpp.endChapter.Changed() {
		body.EndChapter = crpp.endChapter.Get()
	}

	if err := update(
		crpp.Ctx,
		txn,
		"cat.plans",
		kind,
		crpp.PlanUUID,
		body,
	); err != nil {
		return nil, err
	}

	if crpp.fileSourceUUID.Changed() {
		if err := updatePlanSources(
			crpp.Ctx,
			txn,
			crpp.PlanUUID,
			body,
		); err != nil {
			return nil, err
		}
	}

	if crpp.workUUID.Changed() {
		if err := updatePlanWorks(
			crpp.Ctx,
			txn,
			crpp.PlanUUID,
			body,
		); err != nil {
			return nil, err
		}
	}

	if err := txn.Commit(crpp.Ctx); err != nil {
		return nil, fmt.Errorf("%w: failed to commit transaction for patching plan %s: %w", catalog.ErrInternal, crpp.PlanUUID, err)
	}

	return &catalog.Plan{
		UUID:             crpp.PlanUUID,
		ChapterRangePlan: body.ToPublic(),
	}, nil
}

func (crpp *chapterRangePlanPatcher) Save() error {
	_, err := crpp.SaveGet()
	return err
}

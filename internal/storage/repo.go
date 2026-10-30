package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinat1313/source_html/internal/apitypes"
	"github.com/rinat1313/source_html/internal/logging"
	"github.com/rinat1313/source_html/internal/storage/upsertlogic"
)

type Profile struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Run struct {
	ID             string          `json:"id"`
	ProfileID      *string         `json:"profile_id,omitempty"`
	ConfigSnapshot json.RawMessage `json:"config_snapshot"`
	Status         string          `json:"status"`
	ErrorMessage   *string         `json:"error_message,omitempty"`
	Law            string          `json:"law"`
	LimitN         int             `json:"limit_n"`
	MatchedSaved   int             `json:"matched_saved"`
	CreatedAt      time.Time       `json:"created_at"`
	FinishedAt     *time.Time      `json:"finished_at,omitempty"`
}

type Purchase struct {
	ID            string          `json:"id"`
	Law           string          `json:"law"`
	EntityKind    string          `json:"entity_kind"`
	RegNumber     *string         `json:"reg_number,omitempty"`
	GUID          *string         `json:"guid,omitempty"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	CustomerName  string          `json:"customer_name"`
	CustomerINN   *string         `json:"customer_inn,omitempty"`
	CustomerExtra json.RawMessage `json:"customer_extra"`
	FileLinks     json.RawMessage `json:"file_links"`
	EffectiveDate time.Time       `json:"effective_date"`
	RawMeta       json.RawMessage `json:"raw_meta"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) CreateProfile(ctx context.Context, name string, config json.RawMessage) (Profile, error) {
	var err error
	defer logging.Timed("CreateProfile", "name", name)(&err)
	sanitized, err := apitypes.SanitizeConfigJSON(config)
	if err != nil {
		return Profile{}, err
	}
	var p Profile
	err = r.pool.QueryRow(ctx, `
		INSERT INTO search_profiles (name, config) VALUES ($1, $2)
		RETURNING id, name, config, created_at, updated_at
	`, name, sanitized).Scan(&p.ID, &p.Name, &p.Config, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repo) ListProfiles(ctx context.Context) ([]Profile, error) {
	var err error
	defer logging.Timed("ListProfiles")(&err)
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, config, created_at, updated_at FROM search_profiles ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Profile
	for rows.Next() {
		var p Profile
		if err = rows.Scan(&p.ID, &p.Name, &p.Config, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repo) GetProfile(ctx context.Context, id string) (Profile, error) {
	var err error
	defer logging.Timed("GetProfile", "id", id)(&err)
	var p Profile
	err = r.pool.QueryRow(ctx, `
		SELECT id, name, config, created_at, updated_at FROM search_profiles WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Config, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repo) UpdateProfile(ctx context.Context, id, name string, config json.RawMessage) (Profile, error) {
	var err error
	defer logging.Timed("UpdateProfile", "id", id)(&err)
	sanitized, err := apitypes.SanitizeConfigJSON(config)
	if err != nil {
		return Profile{}, err
	}
	var p Profile
	err = r.pool.QueryRow(ctx, `
		UPDATE search_profiles SET name = $2, config = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, name, config, created_at, updated_at
	`, id, name, sanitized).Scan(&p.ID, &p.Name, &p.Config, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repo) DeleteProfile(ctx context.Context, id string) error {
	var err error
	defer logging.Timed("DeleteProfile", "id", id)(&err)
	tag, err := r.pool.Exec(ctx, `DELETE FROM search_profiles WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repo) CreateRun(ctx context.Context, profileID *string, snapshot json.RawMessage, law string, limitN int) (Run, error) {
	var err error
	defer logging.Timed("CreateRun")(&err)
	var run Run
	err = r.pool.QueryRow(ctx, `
		INSERT INTO search_runs (profile_id, config_snapshot, status, law, limit_n)
		VALUES ($1, $2, 'pending', $3, $4)
		RETURNING id, profile_id, config_snapshot, status, error_message, law, limit_n, matched_saved, created_at, finished_at
	`, profileID, snapshot, law, limitN).Scan(
		&run.ID, &run.ProfileID, &run.ConfigSnapshot, &run.Status, &run.ErrorMessage,
		&run.Law, &run.LimitN, &run.MatchedSaved, &run.CreatedAt, &run.FinishedAt,
	)
	return run, err
}

func (r *Repo) GetRun(ctx context.Context, id string) (Run, error) {
	var err error
	defer logging.Timed("GetRun", "id", id)(&err)
	var run Run
	err = r.pool.QueryRow(ctx, `
		SELECT id, profile_id, config_snapshot, status, error_message, law, limit_n, matched_saved, created_at, finished_at
		FROM search_runs WHERE id = $1
	`, id).Scan(
		&run.ID, &run.ProfileID, &run.ConfigSnapshot, &run.Status, &run.ErrorMessage,
		&run.Law, &run.LimitN, &run.MatchedSaved, &run.CreatedAt, &run.FinishedAt,
	)
	return run, err
}

type RunPatch struct {
	Status       *string
	ErrorMessage *string
	MatchedSaved *int
	FinishedAt   *time.Time
}

func (r *Repo) PatchRun(ctx context.Context, id string, patch RunPatch) (Run, error) {
	var err error
	defer logging.Timed("PatchRun", "id", id)(&err)
	run, err := r.GetRun(ctx, id)
	if err != nil {
		return Run{}, err
	}
	if patch.Status != nil {
		run.Status = *patch.Status
	}
	if patch.ErrorMessage != nil {
		run.ErrorMessage = patch.ErrorMessage
	}
	if patch.MatchedSaved != nil {
		run.MatchedSaved = *patch.MatchedSaved
	}
	if patch.FinishedAt != nil {
		run.FinishedAt = patch.FinishedAt
	}
	err = r.pool.QueryRow(ctx, `
		UPDATE search_runs SET status = $2, error_message = $3, matched_saved = $4, finished_at = $5
		WHERE id = $1
		RETURNING id, profile_id, config_snapshot, status, error_message, law, limit_n, matched_saved, created_at, finished_at
	`, id, run.Status, run.ErrorMessage, run.MatchedSaved, run.FinishedAt).Scan(
		&run.ID, &run.ProfileID, &run.ConfigSnapshot, &run.Status, &run.ErrorMessage,
		&run.Law, &run.LimitN, &run.MatchedSaved, &run.CreatedAt, &run.FinishedAt,
	)
	return run, err
}

func (r *Repo) UpsertPurchaseForRun(ctx context.Context, runID string, card apitypes.PurchaseCard) (apitypes.UpsertResult, error) {
	var err error
	defer logging.Timed("UpsertPurchaseForRun", "run_id", runID)(&err)
	if card.Law != "44" && card.Law != "223" {
		err = errors.New("invalid law")
		return apitypes.UpsertResult{}, err
	}
	if card.EntityKind != "notice" && card.EntityKind != "contract" {
		err = errors.New("invalid entity_kind")
		return apitypes.UpsertResult{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return apitypes.UpsertResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var existingID string
	var existingDate time.Time
	var findErr error
	if card.Law == "44" {
		if card.RegNumber == "" {
			err = errors.New("reg_number required for law 44")
			return apitypes.UpsertResult{}, err
		}
		findErr = tx.QueryRow(ctx, `
			SELECT id, effective_date FROM purchases
			WHERE law = '44' AND entity_kind = $1 AND reg_number = $2
		`, card.EntityKind, card.RegNumber).Scan(&existingID, &existingDate)
	} else {
		keyGUID := card.GUID
		if keyGUID == "" {
			if card.RegNumber == "" {
				err = errors.New("guid or reg_number required for law 223")
				return apitypes.UpsertResult{}, err
			}
			findErr = tx.QueryRow(ctx, `
				SELECT id, effective_date FROM purchases
				WHERE law = '223' AND entity_kind = $1 AND guid IS NULL AND reg_number = $2
			`, card.EntityKind, card.RegNumber).Scan(&existingID, &existingDate)
		} else {
			findErr = tx.QueryRow(ctx, `
				SELECT id, effective_date FROM purchases
				WHERE law = '223' AND entity_kind = $1 AND guid = $2
			`, card.EntityKind, keyGUID).Scan(&existingID, &existingDate)
		}
	}

	var existingPtr *time.Time
	action := upsertlogic.ActionCreated
	if findErr == nil {
		existingPtr = &existingDate
		action = upsertlogic.DecideAction(existingPtr, card.EffectiveDate)
	} else if !errors.Is(findErr, pgx.ErrNoRows) {
		err = findErr
		return apitypes.UpsertResult{}, err
	}

	custExtra, _ := json.Marshal(card.CustomerExtra)
	if custExtra == nil {
		custExtra = []byte("{}")
	}
	fileLinks, _ := json.Marshal(card.FileLinks)
	if fileLinks == nil {
		fileLinks = []byte("[]")
	}
	rawMeta, _ := json.Marshal(card.RawMeta)
	if rawMeta == nil {
		rawMeta = []byte("{}")
	}

	purchaseID := existingID
	if action == upsertlogic.ActionCreated {
		var regNum, guid, inn *string
		if card.RegNumber != "" {
			regNum = &card.RegNumber
		}
		if card.GUID != "" {
			guid = &card.GUID
		}
		if card.CustomerINN != "" {
			inn = &card.CustomerINN
		}
		err = tx.QueryRow(ctx, `
			INSERT INTO purchases (
				law, entity_kind, reg_number, guid, title, description,
				customer_name, customer_inn, customer_extra, file_links, effective_date, raw_meta
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			RETURNING id
		`, card.Law, card.EntityKind, regNum, guid, card.Title, card.Description,
			card.CustomerName, inn, custExtra, fileLinks, card.EffectiveDate, rawMeta).Scan(&purchaseID)
		if err != nil {
			return apitypes.UpsertResult{}, err
		}
	} else if action == upsertlogic.ActionUpdated {
		var inn *string
		if card.CustomerINN != "" {
			inn = &card.CustomerINN
		}
		_, err = tx.Exec(ctx, `
			UPDATE purchases SET title = $2, description = $3, customer_name = $4, customer_inn = $5,
				customer_extra = $6, file_links = $7, effective_date = $8, raw_meta = $9, updated_at = now()
			WHERE id = $1
		`, purchaseID, card.Title, card.Description, card.CustomerName, inn, custExtra, fileLinks, card.EffectiveDate, rawMeta)
		if err != nil {
			return apitypes.UpsertResult{}, err
		}
	}

	var alreadyAttached bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM search_run_purchases WHERE run_id = $1 AND purchase_id = $2)
	`, runID, purchaseID).Scan(&alreadyAttached)
	if err != nil {
		return apitypes.UpsertResult{}, err
	}

	if !alreadyAttached {
		_, err = tx.Exec(ctx, `
			INSERT INTO search_run_purchases (run_id, purchase_id) VALUES ($1, $2)
		`, runID, purchaseID)
		if err != nil {
			return apitypes.UpsertResult{}, err
		}
	}

	newSlot := upsertlogic.CountsTowardLimit(action, alreadyAttached)
	if newSlot {
		_, err = tx.Exec(ctx, `
			UPDATE search_runs SET matched_saved = matched_saved + 1 WHERE id = $1
		`, runID)
		if err != nil {
			return apitypes.UpsertResult{}, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return apitypes.UpsertResult{}, err
	}
	return apitypes.UpsertResult{
		PurchaseID: purchaseID,
		Action:     action,
		NewSlot:    newSlot,
	}, nil
}

func (r *Repo) ListPurchases(ctx context.Context, law, q string, offset, limit int) ([]Purchase, int, error) {
	var err error
	defer logging.Timed("ListPurchases")(&err)
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	var args []any
	where := "WHERE 1=1"
	n := 1
	if law != "" {
		where += fmt.Sprintf(" AND law = $%d", n)
		args = append(args, law)
		n++
	}
	if strings.TrimSpace(q) != "" {
		where += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", n, n)
		args = append(args, "%"+strings.TrimSpace(q)+"%")
		n++
	}
	var total int
	countQ := "SELECT COUNT(*) FROM purchases " + where
	if err = r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listQ := fmt.Sprintf(`
		SELECT id, law, entity_kind, reg_number, guid, title, description, customer_name, customer_inn,
			customer_extra, file_links, effective_date, raw_meta, created_at, updated_at
		FROM purchases %s ORDER BY effective_date DESC LIMIT $%d OFFSET $%d
	`, where, n, n+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []Purchase
	for rows.Next() {
		var p Purchase
		if err = rows.Scan(
			&p.ID, &p.Law, &p.EntityKind, &p.RegNumber, &p.GUID, &p.Title, &p.Description,
			&p.CustomerName, &p.CustomerINN, &p.CustomerExtra, &p.FileLinks, &p.EffectiveDate,
			&p.RawMeta, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

func (r *Repo) GetPurchase(ctx context.Context, id string) (Purchase, error) {
	var err error
	defer logging.Timed("GetPurchase", "id", id)(&err)
	var p Purchase
	err = r.pool.QueryRow(ctx, `
		SELECT id, law, entity_kind, reg_number, guid, title, description, customer_name, customer_inn,
			customer_extra, file_links, effective_date, raw_meta, created_at, updated_at
		FROM purchases WHERE id = $1
	`, id).Scan(
		&p.ID, &p.Law, &p.EntityKind, &p.RegNumber, &p.GUID, &p.Title, &p.Description,
		&p.CustomerName, &p.CustomerINN, &p.CustomerExtra, &p.FileLinks, &p.EffectiveDate,
		&p.RawMeta, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

type RunResultItem struct {
	AttachedAt time.Time `json:"attached_at"`
	Purchase   Purchase  `json:"purchase"`
}

func (r *Repo) ListRunResults(ctx context.Context, runID string, offset, limit int) ([]RunResultItem, int, error) {
	var err error
	defer logging.Timed("ListRunResults", "run_id", runID)(&err)
	if limit <= 0 {
		limit = 10
	}
	var total int
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM search_run_purchases WHERE run_id = $1
	`, runID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT srp.attached_at,
			p.id, p.law, p.entity_kind, p.reg_number, p.guid, p.title, p.description, p.customer_name, p.customer_inn,
			p.customer_extra, p.file_links, p.effective_date, p.raw_meta, p.created_at, p.updated_at
		FROM search_run_purchases srp
		JOIN purchases p ON p.id = srp.purchase_id
		WHERE srp.run_id = $1
		ORDER BY srp.attached_at ASC
		LIMIT $2 OFFSET $3
	`, runID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []RunResultItem
	for rows.Next() {
		var item RunResultItem
		if err = rows.Scan(
			&item.AttachedAt,
			&item.Purchase.ID, &item.Purchase.Law, &item.Purchase.EntityKind, &item.Purchase.RegNumber, &item.Purchase.GUID,
			&item.Purchase.Title, &item.Purchase.Description, &item.Purchase.CustomerName, &item.Purchase.CustomerINN,
			&item.Purchase.CustomerExtra, &item.Purchase.FileLinks, &item.Purchase.EffectiveDate,
			&item.Purchase.RawMeta, &item.Purchase.CreatedAt, &item.Purchase.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	return out, total, rows.Err()
}

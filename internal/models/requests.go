package models

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"secondarymetabolites.org/mibig-api/internal/data"
)

type RequestModel interface {
	AddAccessionRequest(request *data.AccessionRequest) error
}

type LiveRequestModel struct {
	DB *sql.DB
}

func (r *LiveRequestModel) AddAccessionRequest(request *data.AccessionRequest) error {
	query := `INSERT INTO mibig_submissions.accession_requests (user_id, compounds)
		VALUES ($1, $2) RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_DB_TIMEOUT)
	defer cancel()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var requestId int

	err = tx.QueryRowContext(ctx, query, request.UserId, pq.Array(request.Compounds)).Scan(&requestId)
	if err != nil {
		tx.Rollback()
		return err
	}

	query = `INSERT INTO mibig_submissions.accession_request_loci (accession, start, end, request)
		VALUES ($1, $2, $3, $4)`

	for _, locus := range request.Loci {

		args := []interface{}{
			locus.GenBankAccession,
			locus.Start,
			locus.End,
			request.UserId,
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func NewRequestModel(db *sql.DB) *LiveRequestModel {
	return &LiveRequestModel{DB: db}
}

type MockRequestModel struct {
}

func NewMockRequestModel(db *sql.DB) *MockRequestModel {
	return &MockRequestModel{}
}

func (r *MockRequestModel) AddAccessionRequest(request *data.AccessionRequest) error {
	return data.ErrNotImplemented
}

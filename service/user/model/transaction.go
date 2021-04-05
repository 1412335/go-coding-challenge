package model

import (
	"time"

	pb "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/user"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type Transaction struct {
	ID              int64     `json:"id"`
	AccountID       int64     `json:"account_id" validate:"nonzero"`
	Amount          float64   `json:"amount" validate:"min=1"`
	TransactionType string    `json:"transaction_type"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (t *Transaction) Transform2GRPC() *pb.Transaction {
	trans := &pb.Transaction{
		Id:        t.ID,
		AccountId: t.AccountID,
		Amount:    t.Amount,
		CreatedAt: timestamppb.New(t.CreatedAt),
	}
	//
	if t, ok := pb.TransactionType_value[t.TransactionType]; ok {
		trans.TransactionType = pb.TransactionType(t)
	}
	return trans
}

// func (t *Transaction) updateFromGRPC(trans *pb.Transaction) {
// 	t.AccountID = trans.GetAccountId()
// 	t.Amount = trans.GetAmount()
// 	t.TransactionType = trans.GetTransactionType().String()
// }

func (t *Transaction) cache() error {
	return nil
}

func (t *Transaction) rmCache() error {
	return nil
}

func (t *Transaction) sanitize() {
}

func (t *Transaction) Validate() error {
	// sanitize fileds
	t.sanitize()
	// validate
	if e := validator.Validate(t); e != nil {
		errs, ok := e.(validator.ErrorMap)
		if !ok {
			return errors.BadRequest("validate failed", map[string]string{"error": errs.Error()})
		}
		fields := make(map[string]string, len(errs))
		for field, err := range errs {
			fields[field] = err[0].Error()
		}
		return errors.BadRequest("validate failed", fields)
	}
	return nil
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	return nil
}

func (t *Transaction) AfterCreate(tx *gorm.DB) error {
	// cache user
	if err := t.cache(); err != nil {
		return err
	}
	return nil
}

func (t *Transaction) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// Updating data in same transaction
func (t *Transaction) AfterUpdate(tx *gorm.DB) error {
	// cache user
	if err := t.cache(); err != nil {
		return err
	}
	return nil
}

func (t *Transaction) BeforeDelete(tx *gorm.DB) error {
	return nil
}

func (t *Transaction) AfterDelete(tx *gorm.DB) error {
	// rm cache user
	if err := t.rmCache(); err != nil {
		return err
	}
	return nil
}

package model

import (
	"fmt"
	"time"

	pb "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/user"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/errors"
	"github.com/microcosm-cc/bluemonday"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type Account struct {
	ID           int64         `json:"id"`
	UserID       int64         `json:"user_id" validate:"nonzero"`
	Name         string        `json:"name" validate:"max=100"`
	Bank         string        `json:"bank" validate:"nonzero"`
	Balance      float64       `json:"balance"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Transactions []Transaction `json:"transactions"`
}

func (a *Account) Transform2GRPC() *pb.Account {
	acc := &pb.Account{
		Id:        a.ID,
		Name:      a.Name,
		UserId:    a.UserID,
		Balance:   a.Balance,
		CreatedAt: timestamppb.New(a.CreatedAt),
	}
	//
	if bank, ok := pb.Bank_value[a.Bank]; ok {
		acc.Bank = pb.Bank(bank)
	}
	return acc
}

// func (a *Account) updateFromGRPC(acc *pb.Account) {
// 	a.UserID = acc.GetUserId()
// 	a.Name = acc.GetName()
// 	a.Bank = acc.GetBank().String()
// }

func (a *Account) cache() error {
	return nil
}

func (a *Account) rmCache() error {
	return nil
}

func (a *Account) sanitize() {
	p := bluemonday.UGCPolicy()
	a.Name = p.Sanitize(a.Name)
}

func (a *Account) Validate() error {
	// sanitize fileds
	a.sanitize()
	// validate
	e := validator.Validate(a)
	if e != nil {
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

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	if a.Name == "" {
		a.Name = fmt.Sprintf("%d.%s", a.UserID, a.Bank)
	}
	return nil
}

func (a *Account) AfterCreate(tx *gorm.DB) error {
	// cache user
	if err := a.cache(); err != nil {
		return err
	}
	return nil
}

func (a *Account) BeforeUpdate(tx *gorm.DB) error {
	if a.Name == "" {
		a.Name = fmt.Sprintf("%d.%s", a.UserID, a.Bank)
	}
	return nil
}

// Updating data in same transaction
func (a *Account) AfterUpdate(tx *gorm.DB) error {
	// cache user
	if err := a.cache(); err != nil {
		return err
	}
	return nil
}

func (a *Account) BeforeDelete(tx *gorm.DB) error {
	return nil
}

func (a *Account) AfterDelete(tx *gorm.DB) error {
	// rm cache user
	if err := a.rmCache(); err != nil {
		return err
	}
	return nil
}

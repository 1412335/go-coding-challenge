package user

import (
	"strings"
	"time"

	pb "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/user"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/errors"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/utils"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `gorm:"uniqueIndex" validate:"nonzero"`
	Password  string    `json:"-" validate:"min=8"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) transform2GRPC() *pb.User {
	user := &pb.User{
		Id:    u.ID,
		Email: u.Email,
	}
	return user
}

func (u *User) updateFromGRPC(user *pb.User) {
	u.Email = user.GetEmail()
	u.Password = user.GetPassword()
}

func (u *User) cache() error {
	return nil
}

func (u *User) rmCache() error {
	return nil
}

func (u *User) hashPassword() error {
	// hash password
	hashedPassword, err := utils.GenHash(u.Password)
	if err != nil {
		// 	u.logger.For(ctx).Error("Hash password failed", zap.Error(err))
		return ErrHashPassword
	}
	u.Password = hashedPassword
	return nil
}

func (u *User) sanitize() {
}

func (u *User) validate() error {
	// sanitize fileds
	u.sanitize()
	// validate
	errs, ok := validator.Validate(u).(validator.ErrorMap)
	if !ok {
		return errors.BadRequest("validate failed", map[string]string{"errors": errs.Error()})
	}
	fields := make(map[string]string, len(errs))
	for field, err := range errs {
		fields[field] = err[0].Error()
	}
	return errors.BadRequest("validate failed", fields)
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if err := u.hashPassword(); err != nil {
		return err
	}
	u.Email = strings.ToLower(u.Email)
	return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
	// cache user
	if err := u.cache(); err != nil {
		return err
	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if err := u.hashPassword(); err != nil {
		return err
	}
	u.Email = strings.ToLower(u.Email)
	return nil
}

// Updating data in same transaction
func (u *User) AfterUpdate(tx *gorm.DB) error {
	// cache user
	if err := u.cache(); err != nil {
		return err
	}
	return nil
}

func (u *User) BeforeDelete(tx *gorm.DB) error {
	return nil
}

func (u *User) AfterDelete(tx *gorm.DB) error {
	// rm cache user
	if err := u.rmCache(); err != nil {
		return err
	}
	return nil
}

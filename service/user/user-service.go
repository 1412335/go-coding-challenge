package user

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/structs"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	pb "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/user"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/dal/postgres"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/errors"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/utils"
	errorSrv "github.com/1412335/moneyforward-go-coding-challenge/service/user/error"
	"github.com/1412335/moneyforward-go-coding-challenge/service/user/model"
)

type userServiceImpl struct {
	dal      *postgres.DataAccessLayer
	logger   log.Factory
	tokenSrv *TokenService
}

var _ pb.UserServiceServer = (*userServiceImpl)(nil)

func NewUserService(dal *postgres.DataAccessLayer, tokenSrv *TokenService) pb.UserServiceServer {
	return &userServiceImpl{
		dal:      dal,
		logger:   log.With(zap.String("srv", "user")),
		tokenSrv: tokenSrv,
	}
}

// get user by id from redis & db
func (u *userServiceImpl) getUserByID(ctx context.Context, id int64) (*model.User, error) {
	user := &model.User{}
	err := u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// find user by id
		if e := tx.Where(&model.User{ID: id}).First(user).Error; e == gorm.ErrRecordNotFound {
			return errorSrv.ErrUserNotFound
		} else if e != nil {
			u.logger.For(ctx).Error("Find user", zap.Error(e))
			return errorSrv.ErrConnectDB
		}
		// // cache
		// if e := user.cache(); e != nil {
		// 	u.logger.For(ctx).Error("Cache user", zap.Error(e))
		// }
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, err
}

// create user & token
func (u *userServiceImpl) Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// validate request
	if !isValidEmail(req.GetEmail()) {
		return nil, errorSrv.ErrInvalidEmail
	}
	if !isValidPassword(req.GetPassword()) {
		return nil, errorSrv.ErrInvalidPassword
	}

	user := &model.User{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	if err := user.Validate(); err != nil {
		u.logger.For(ctx).Error("Error validate user", zap.Error(err))
		return nil, err
	}

	// init response
	rsp := &pb.CreateUserResponse{}

	// create
	err := u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil && strings.Contains(err.Error(), "idx_users_email") {
			return errorSrv.ErrDuplicateEmail
		} else if err != nil {
			u.logger.For(ctx).Error("Error connecting from db", zap.Error(err))
			return errorSrv.ErrConnectDB
		}

		// create token
		token, err := u.tokenSrv.Generate(user)
		if err != nil {
			u.logger.For(ctx).Error("Error generate token", zap.Error(err))
			return errorSrv.ErrTokenGenerated
		}

		rsp.User = user.Transform2GRPC()
		rsp.Token = token
		return nil
	})
	if err != nil {
		return nil, err
	}
	// set header in your handler
	md := metadata.Pairs("X-Http-Code", "201")
	grpc.SetHeader(ctx, md)
	return rsp, nil
}

// delete user by id
func (u *userServiceImpl) Delete(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if req.GetId() == 0 {
		return nil, errorSrv.ErrMissingUserID
	}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(req.GetId()).Delete(&model.User{}).Error; err == gorm.ErrRecordNotFound {
			return errorSrv.ErrUserNotFound
		} else if err != nil {
			u.logger.For(ctx).Error("Error connecting from db", zap.Error(err))
			return errorSrv.ErrConnectDB
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &pb.DeleteUserResponse{
		Id: req.GetId(),
	}, nil
}

// update user by id
func (u *userServiceImpl) Update(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if req.GetUser().GetId() == 0 {
		return nil, errorSrv.ErrMissingUserID
	}
	rsp := &pb.UpdateUserResponse{}
	err := u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// find user by id
		user, e := u.getUserByID(tx.Statement.Context, req.GetUser().GetId())
		if e != nil {
			u.logger.For(ctx).Error("Get user by ID", zap.Error(e))
			return errors.InternalServerError("Get user failed", "Lookup user by ID w redis/db failed")
		}

		u.logger.For(ctx).Info("mask", zap.Strings("path", req.GetUpdateMask().GetPaths()))
		// If there is no update mask do a regular update
		if req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
			user.UpdateFromGRPC(req.GetUser())
		} else {
			st := structs.New(*user)
			in := structs.New(req.GetUser())
			for _, path := range req.GetUpdateMask().GetPaths() {
				if path == "id" {
					return errors.BadRequest("cannot update id", map[string]string{"update_mask": "cannot update id field"})
				}
				// This doesn't translate properly if a CustomName setting is used,
				// but none of the fields except ID has that set, so NO WORRIES.
				fname := path
				field, ok := st.FieldOk(fname)
				if !ok {
					return errors.BadRequest("invalid field specified", map[string]string{
						"update_mask": fmt.Sprintf("The user message type does not have a field called %q", path),
					})
				}
				// set update value
				if e := field.Set(in.Field(fname).Value()); e != nil {
					return e
				}
			}
		}
		// check fields valid
		if !isValidEmail(user.Email) {
			return errorSrv.ErrInvalidEmail
		}
		if !isValidPassword(user.Password) {
			return errorSrv.ErrInvalidPassword
		}
		if err := user.Validate(); err != nil {
			u.logger.For(ctx).Error("Error validate user", zap.Error(err))
			return err
		}
		// update user in db
		if e := tx.Save(user).Error; e != nil && strings.Contains(e.Error(), "idx_users_email") {
			return errorSrv.ErrDuplicateEmail
		} else if e != nil {
			return errorSrv.ErrConnectDB
		}
		// response
		rsp.User = user.Transform2GRPC()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}

// build query statement & get list users
func (u *userServiceImpl) getUsers(ctx context.Context, req *pb.ListUsersRequest) ([]*pb.User, error) {
	var users []model.User
	// build sql statement
	psql := u.dal.GetDatabase().WithContext(ctx)
	if req.GetId() != nil {
		psql = psql.Where("id = ?", req.GetId())
	}
	if req.GetEmail() != nil {
		psql = psql.Where("email LIKE '%?%'", req.GetEmail().Value)
	}
	// exec
	if err := psql.Order("created_at desc").Find(&users).Error; err != nil {
		u.logger.For(ctx).Error("Error find users", zap.Error(err))
		return nil, errorSrv.ErrConnectDB
	}
	// check empty from db
	if len(users) == 0 {
		return nil, errorSrv.ErrUserNotFound
	}
	// filter
	rsp := make([]*pb.User, len(users))
	for i, user := range users {
		rsp[i] = user.Transform2GRPC()
	}
	return rsp, nil
}

// list users w unary response
func (u *userServiceImpl) List(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := u.getUsers(ctx, req)
	if err != nil {
		return nil, err
	}
	// response
	rsp := &pb.ListUsersResponse{
		Users: users,
	}
	return rsp, nil
}

// list users w stream response
func (u *userServiceImpl) ListStream(req *pb.ListUsersRequest, srv pb.UserService_ListStreamServer) error {
	users, err := u.getUsers(srv.Context(), req)
	if err != nil {
		return err
	}
	for _, user := range users {
		if err := srv.Send(user); err != nil {
			return err
		}
	}
	return nil
}

// login w email + pwd & gen token
func (u *userServiceImpl) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// validate request
	if len(req.GetEmail()) == 0 {
		return nil, errorSrv.ErrMissingEmail
	}
	if !isValidEmail(req.GetEmail()) {
		return nil, errorSrv.ErrInvalidEmail
	}
	if len(req.GetPassword()) == 0 {
		return nil, errorSrv.ErrInvalidPassword
	}
	// response
	rsp := &pb.LoginResponse{}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		var user model.User
		// find user by email
		if e := tx.Where(&model.User{Email: strings.ToLower(req.GetEmail())}).First(&user).Error; e == gorm.ErrRecordNotFound {
			return errorSrv.ErrUserNotFound
		} else if e != nil {
			u.logger.For(ctx).Error("Error find user", zap.Error(e))
			return errorSrv.ErrConnectDB
		}
		// verify password
		if e := utils.CompareHash(user.Password, req.GetPassword()); e != nil {
			return errorSrv.ErrIncorrectPassword
		}
		// gen new token
		token, e := u.tokenSrv.Generate(&user)
		if e != nil {
			u.logger.For(ctx).Error("Error gen token", zap.Error(e))
			return errorSrv.ErrTokenGenerated
		}
		// // cache user
		// if e := user.cache(); e != nil {
		// 	u.logger.For(ctx).Error("Cache user", zap.Error(e))
		// }
		//
		rsp.User = user.Transform2GRPC()
		rsp.Token = token
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}

// logout: clear redis cache
func (u *userServiceImpl) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if req.GetId() == 0 {
		return nil, errorSrv.ErrMissingUserID
	}
	return nil, nil
}

// validate token: update isActive=true & return user
func (u *userServiceImpl) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	if len(req.GetToken()) == 0 {
		return nil, errorSrv.ErrMissingToken
	}
	rsp := &pb.ValidateResponse{}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		// verrify token
		claims, e := u.tokenSrv.Verify(req.Token)
		if e != nil {
			u.logger.For(ctx).Error("verify token failed", zap.Error(e))
			return errorSrv.ErrTokenInvalid
		}
		// get cache user
		user, e := u.getUserByID(ctx, claims.ID)
		if e != nil {
			u.logger.For(ctx).Error("Get user by ID", zap.Error(e))
			return errors.InternalServerError("Get user failed", "Lookup user by id failed")
		}
		rsp.User = user.Transform2GRPC()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}

// CreateAccount
func (u *userServiceImpl) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	// validate request
	if req.GetUserId() == 0 {
		return nil, errorSrv.ErrMissingUserID
	}
	if req.GetBalance() < 0 {
		return nil, errorSrv.ErrInvalidAccountBalance
	}

	// response
	rsp := &pb.CreateAccountResponse{}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		var user model.User
		// find user by id
		if e := tx.Where(&model.User{ID: req.GetUserId()}).First(&user).Error; e == gorm.ErrRecordNotFound {
			return errorSrv.ErrUserNotFound
		} else if e != nil {
			u.logger.For(ctx).Error("Error find user by id", zap.Error(e))
			return errorSrv.ErrConnectDB
		}

		// create account
		acc := &model.Account{
			UserID:  user.ID,
			Name:    req.GetName(),
			Bank:    req.GetBank().String(),
			Balance: req.GetBalance(),
		}
		if err := acc.Validate(); err != nil {
			u.logger.For(ctx).Error("Error validate account", zap.Error(err))
			return err
		}
		if err := tx.Create(acc).Error; err != nil {
			u.logger.For(ctx).Error("Error create account", zap.Error(err))
			return errorSrv.ErrConnectDB
		}
		//
		rsp.Account = acc.Transform2GRPC()
		return nil
	})
	if err != nil {
		return nil, err
	}
	// set header in your handler
	md := metadata.Pairs("X-Http-Code", "201")
	grpc.SetHeader(ctx, md)
	return rsp, nil
}

// ListAccounts
func (u *userServiceImpl) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	// validate request
	if req.GetUserId() == nil {
		return nil, errorSrv.ErrMissingUserID
	}

	var user model.User
	// lookup user by id
	if e := u.dal.GetDatabase().Where(&model.User{ID: req.GetUserId().Value}).Preload("Accounts").First(&user).Error; e == gorm.ErrRecordNotFound {
		return nil, errorSrv.ErrUserNotFound
	} else if e != nil {
		u.logger.For(ctx).Error("Error find user by id", zap.Error(e))
		return nil, errorSrv.ErrConnectDB
	}
	rsp := &pb.ListAccountsResponse{}
	// fetch accounts belong to the user
	rsp.Accounts = make([]*pb.Account, len(user.Accounts))
	for i, acc := range user.Accounts {
		rsp.Accounts[i] = acc.Transform2GRPC()
	}
	return rsp, nil
}

// CreateTransaction
func (u *userServiceImpl) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	// validate request
	if req.GetUserId() == 0 {
		return nil, errorSrv.ErrMissingUserID
	}
	if req.GetAccountId() == 0 {
		return nil, errorSrv.ErrMissingAccountID
	}
	if req.GetAmount() <= 0 {
		return nil, errorSrv.ErrInvalidTransactionAmountGT0
	}

	// response
	rsp := &pb.CreateTransactionResponse{}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		var acc model.Account
		// find account by userId + accId
		if e := tx.Where(&model.Account{ID: req.GetAccountId(), UserID: req.GetUserId()}).First(&acc).Error; e == gorm.ErrRecordNotFound {
			return errorSrv.ErrAccountNotFound
		} else if e != nil {
			u.logger.For(ctx).Error("Error find user by id", zap.Error(e))
			return errorSrv.ErrConnectDB
		}

		// // check account balance w withdraw transaction
		// if req.GetTransactionType() == pb.TransactionType_WITHDRAW && acc.Balance < req.GetAmount() {
		// 	return ErrInvalidTransactionAmount
		// }
		// check account balance
		switch req.GetTransactionType() {
		case pb.TransactionType_WITHDRAW:
			acc.Balance -= req.GetAmount()
			if acc.Balance < 0 {
				return errorSrv.ErrInvalidWithdrawTransactionAmount
			}
		case pb.TransactionType_DEPOSIT:
			acc.Balance += req.GetAmount()
		}

		// create transaction
		trans := &model.Transaction{
			AccountID:       acc.ID,
			Amount:          req.GetAmount(),
			TransactionType: req.GetTransactionType().String(),
			CreatedAt:       time.Now().Round(time.Millisecond),
		}
		if err := trans.Validate(); err != nil {
			u.logger.For(ctx).Error("Error validate trans", zap.Error(err))
			return err
		}
		if err := tx.Create(trans).Error; err != nil {
			u.logger.For(ctx).Error("Error create transaction", zap.Error(err))
			return errorSrv.ErrConnectDB
		}

		// update account
		if e := tx.Save(&acc).Error; e != nil {
			u.logger.For(ctx).Error("Error update account balance", zap.Error(e))
			return errorSrv.ErrConnectDB
		}

		// response
		rsp.Transaction = trans.Transform2GRPC()
		return nil
	})
	if err != nil {
		return nil, err
	}

	// set header in your handler
	md := metadata.Pairs("X-Http-Code", "201")
	grpc.SetHeader(ctx, md)
	return rsp, nil
}

// ListTransactions
func (u *userServiceImpl) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	// validate request
	if req.GetUserId() == 0 {
		return nil, errorSrv.ErrMissingUserID
	}

	// build query
	q := u.dal.GetDatabase().Where(&model.Account{UserID: req.GetUserId()})
	if req.GetAccountId() != 0 {
		q = q.Where("id = ?", req.GetAccountId())
	}

	// lookup acc & its transactions
	var accs []model.Account
	if e := q.Preload("Transactions").Find(&accs).Error; e != nil {
		u.logger.For(ctx).Error("Error find user by id", zap.Error(e))
		return nil, errorSrv.ErrConnectDB
	}

	if len(accs) == 0 {
		return nil, errorSrv.ErrTransactionNotFound
	}

	// get all transactions
	rsp := &pb.ListTransactionsResponse{}
	rsp.Transactions = []*pb.ListTransactionsResponse_Result{}
	for _, acc := range accs {
		for _, trans := range acc.Transactions {
			pbTrans := &pb.ListTransactionsResponse_Result{
				Id:        trans.ID,
				AccountId: trans.AccountID,
				Amount:    trans.Amount,
				CreatedAt: timestamppb.New(trans.CreatedAt),
			}
			if b, ok := pb.Bank_value[acc.Bank]; ok {
				pbTrans.Bank = pb.Bank(b)
			}
			if t, ok := pb.TransactionType_value[trans.TransactionType]; ok {
				pbTrans.TransactionType = pb.TransactionType(t)
			}
			rsp.Transactions = append(rsp.Transactions, pbTrans)
		}
	}
	return rsp, nil
}

// DeleteTransaction
func (u *userServiceImpl) DeleteTransaction(ctx context.Context, req *pb.DeleteTransactionRequest) (*pb.DeleteTransactionResponse, error) {
	// validate request
	if req.GetUserId() == 0 {
		return nil, errorSrv.ErrMissingUserID
	}

	// build query
	q := u.dal.GetDatabase().Where(&model.Account{UserID: req.GetUserId()})
	if req.GetAccountId() != nil {
		q = q.Where("id = ?", req.GetAccountId().Value)
	}

	// lookup account
	var accs []model.Account
	if e := q.Preload("Transactions", func(db *gorm.DB) *gorm.DB {
		if req.GetId() != nil {
			db = db.Where("id = ?", req.GetId().Value)
		}
		return db
	}).Find(&accs).Error; e != nil {
		u.logger.For(ctx).Error("Error find user by id", zap.Error(e))
		return nil, errorSrv.ErrConnectDB
	}
	if len(accs) == 0 {
		return nil, errorSrv.ErrTransactionNotFound
	}

	// lookup transaction
	var ids []int64
	for _, acc := range accs {
		for _, trans := range acc.Transactions {
			ids = append(ids, trans.ID)
		}
	}
	if e := u.dal.GetDatabase().Where("id IN ?", ids).Delete(&model.Transaction{}).Error; e != nil {
		u.logger.For(ctx).Error("Error delete transaction", zap.Error(e))
		return nil, errorSrv.ErrConnectDB
	}
	return &pb.DeleteTransactionResponse{Ids: ids}, nil
}

// UpdateTransaction
func (u *userServiceImpl) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.UpdateTransactionResponse, error) {
	if req.GetUserId() == 0 {
		return nil, errorSrv.ErrMissingUserID
	}
	if req.GetAccountId() == 0 {
		return nil, errorSrv.ErrMissingAccountID
	}
	if req.GetTransaction().GetId() == 0 {
		return nil, errorSrv.ErrMissingTransactionID
	}

	rsp := &pb.UpdateTransactionResponse{}
	err := u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// find acc
		var acc model.Account
		if e := tx.Where(&model.Account{ID: req.GetAccountId(), UserID: req.GetUserId()}).Preload("Transactions", func(db *gorm.DB) *gorm.DB {
			db = db.Where("id = ?", req.GetTransaction().GetId())
			return db
		}).First(&acc).Error; e == gorm.ErrRecordNotFound {
			return errorSrv.ErrAccountNotFound
		} else if e != nil {
			u.logger.For(ctx).Error("Error find account", zap.Error(e))
			return errorSrv.ErrConnectDB
		}

		if len(acc.Transactions) == 0 {
			return errorSrv.ErrTransactionNotFound
		}
		trans := acc.Transactions[0]

		// If there is no update mask do a regular update
		if req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
			switch trans.TransactionType {
			case pb.TransactionType_WITHDRAW.String():
				acc.Balance += trans.Amount - req.GetTransaction().GetAmount()
				if acc.Balance < 0 {
					return errorSrv.ErrInvalidWithdrawTransactionAmount
				}
			case pb.TransactionType_DEPOSIT.String():
				acc.Balance = acc.Balance - trans.Amount + req.GetTransaction().GetAmount()
				if acc.Balance < 0 {
					return errorSrv.ErrInvalidTransactionAmountGT0
				}
			}
			trans.Amount = req.GetTransaction().GetAmount()
		} else {
			paths := req.GetUpdateMask().GetPaths()
			sort.Strings(paths)
			for _, path := range paths {
				if path == "id" {
					return errors.BadRequest("cannot update id", map[string]string{"update_mask": "cannot update id field"})
				}
				if path == "transaction_type" {
					return errors.BadRequest("cannot update transaction type", map[string]string{"update_mask": "cannot update transaction_type"})
				}
				if path == "amount" {
					switch trans.TransactionType {
					case pb.TransactionType_WITHDRAW.String():
						acc.Balance += trans.Amount - req.GetTransaction().GetAmount()
						if acc.Balance < 0 {
							return errorSrv.ErrInvalidWithdrawTransactionAmount
						}
					case pb.TransactionType_DEPOSIT.String():
						acc.Balance = acc.Balance - trans.Amount + req.GetTransaction().GetAmount()
						if acc.Balance < 0 {
							return errorSrv.ErrInvalidTransactionAmountGT0
						}
					}
					trans.Amount = req.GetTransaction().GetAmount()
				}
			}
		}
		// update trans
		if e := tx.Save(trans).Error; e != nil {
			u.logger.For(ctx).Error("Error update trans", zap.Error(e))
			return errorSrv.ErrConnectDB
		}
		// update acc balance
		if e := tx.Save(&acc).Error; e != nil {
			u.logger.For(ctx).Error("Error update account", zap.Error(e))
			return errorSrv.ErrConnectDB
		}
		// response
		rsp.Transaction = trans.Transform2GRPC()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}

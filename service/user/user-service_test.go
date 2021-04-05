package user

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/user"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/configs"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/dal/postgres"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	errorSrv "github.com/1412335/moneyforward-go-coding-challenge/service/user/error"
	"github.com/1412335/moneyforward-go-coding-challenge/service/user/model"
	"go.uber.org/zap"
)

func newUserService(t *testing.T) pb.UserServiceServer {
	// service configs
	config := configs.ServiceConfig{
		Database: &configs.Database{
			Host:           "postgres",
			Port:           "5432",
			User:           "root",
			Password:       "root",
			Scheme:         "users",
			MaxIdleConns:   10,
			MaxOpenConns:   100,
			ConnectTimeout: 1 * time.Hour,
		},
		JWT: &configs.JWT{
			SecretKey: "lu",
			Duration:  10 * time.Minute,
			Issuer:    "lu",
		},
	}

	// init postgres
	dal, err := postgres.NewDataAccessLayer(context.Background(), config.Database)
	require.NoError(t, err)
	require.NotNil(t, dal.GetDatabase())

	// truncate table
	err = dal.GetDatabase().Exec("TRUNCATE TABLE users CASCADE").Error
	require.NoError(t, err)

	// migrate db
	if err := dal.GetDatabase().AutoMigrate(
		&model.User{},
		&model.Account{},
		&model.Transaction{},
	); err != nil {
		log.Fatal("migrate db failed", zap.Error(err))
		return nil
	}

	// token service
	tokenSrv := NewTokenService(config.JWT)

	// create server
	return NewUserService(dal, tokenSrv)
}

func TestNewUserService(t *testing.T) {
	h := newUserService(t)
	require.NotNil(t, h)
}

func Test_userServiceImpl_getUserByID(t *testing.T) {
	type fields struct {
		dal      *postgres.DataAccessLayer
		logger   log.Factory
		tokenSrv *TokenService
	}
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &userServiceImpl{
				dal:      tt.fields.dal,
				logger:   tt.fields.logger,
				tokenSrv: tt.fields.tokenSrv,
			}
			got, err := u.getUserByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("userServiceImpl.getUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userServiceImpl.getUserByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userServiceImpl_Create(t *testing.T) {
	s := newUserService(t)
	require.NotNil(t, s)

	tests := []struct {
		name string
		req  *pb.CreateUserRequest
		err  error
	}{
		{
			name: "MissingUserEmail",
			req: &pb.CreateUserRequest{
				Email: "",
			},
			err: errorSrv.ErrInvalidEmail,
		},
		{
			name: "MissingUserPassword",
			req: &pb.CreateUserRequest{
				Email: "abc@gmail.com",
			},
			err: errorSrv.ErrInvalidPassword,
		},
		{
			name: "Success",
			req: &pb.CreateUserRequest{
				Email:    "abc@gmail.com",
				Password: "abc123",
			},
		},
		{
			name: "DuplicateUserEmail",
			req: &pb.CreateUserRequest{
				Email:    "abc@gmail.com",
				Password: "abc123",
			},
			err: errorSrv.ErrDuplicateEmail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := s.Create(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp.User)
				require.Empty(t, rsp.Token)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp.User)
				require.Equal(t, tt.req.Email, rsp.User.Email)
				require.NotEmpty(t, rsp.Token)
			}
		})
	}
}

func Test_userServiceImpl_Delete(t *testing.T) {
	type fields struct {
		dal      *postgres.DataAccessLayer
		logger   log.Factory
		tokenSrv *TokenService
	}
	type args struct {
		ctx context.Context
		req *pb.DeleteUserRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.DeleteUserResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &userServiceImpl{
				dal:      tt.fields.dal,
				logger:   tt.fields.logger,
				tokenSrv: tt.fields.tokenSrv,
			}
			got, err := u.Delete(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("userServiceImpl.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userServiceImpl.Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userServiceImpl_Update(t *testing.T) {
	type fields struct {
		dal      *postgres.DataAccessLayer
		logger   log.Factory
		tokenSrv *TokenService
	}
	type args struct {
		ctx context.Context
		req *pb.UpdateUserRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.UpdateUserResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &userServiceImpl{
				dal:      tt.fields.dal,
				logger:   tt.fields.logger,
				tokenSrv: tt.fields.tokenSrv,
			}
			got, err := u.Update(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("userServiceImpl.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userServiceImpl.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userServiceImpl_getUsers(t *testing.T) {
	type fields struct {
		dal      *postgres.DataAccessLayer
		logger   log.Factory
		tokenSrv *TokenService
	}
	type args struct {
		ctx context.Context
		req *pb.ListUsersRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*pb.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &userServiceImpl{
				dal:      tt.fields.dal,
				logger:   tt.fields.logger,
				tokenSrv: tt.fields.tokenSrv,
			}
			got, err := u.getUsers(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("userServiceImpl.getUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userServiceImpl.getUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userServiceImpl_List(t *testing.T) {
	type fields struct {
		dal      *postgres.DataAccessLayer
		logger   log.Factory
		tokenSrv *TokenService
	}
	type args struct {
		ctx context.Context
		req *pb.ListUsersRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.ListUsersResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &userServiceImpl{
				dal:      tt.fields.dal,
				logger:   tt.fields.logger,
				tokenSrv: tt.fields.tokenSrv,
			}
			got, err := u.List(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("userServiceImpl.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userServiceImpl.List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userServiceImpl_ListStream(t *testing.T) {
	type fields struct {
		dal      *postgres.DataAccessLayer
		logger   log.Factory
		tokenSrv *TokenService
	}
	type args struct {
		req *pb.ListUsersRequest
		srv pb.UserService_ListStreamServer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &userServiceImpl{
				dal:      tt.fields.dal,
				logger:   tt.fields.logger,
				tokenSrv: tt.fields.tokenSrv,
			}
			if err := u.ListStream(tt.args.req, tt.args.srv); (err != nil) != tt.wantErr {
				t.Errorf("userServiceImpl.ListStream() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_userServiceImpl_Login(t *testing.T) {
	s := newUserService(t)
	require.NotNil(t, s)

	// mockup
	req := &pb.CreateUserRequest{
		Email:    "abc@gmail.com",
		Password: "abc123",
	}
	rspCreated, err := s.Create(context.TODO(), req)
	require.NoError(t, err)
	require.NotNil(t, rspCreated.User)
	require.Equal(t, req.Email, rspCreated.User.Email)
	require.NotEmpty(t, rspCreated.Token)

	tests := []struct {
		name string
		req  *pb.LoginRequest
		err  error
	}{
		{
			name: "MissingUserEmail",
			req: &pb.LoginRequest{
				Email: "",
			},
			err: errorSrv.ErrMissingEmail,
		},
		{
			name: "InvalidUserEmail",
			req: &pb.LoginRequest{
				Email: "abc",
			},
			err: errorSrv.ErrInvalidEmail,
		},
		{
			name: "MissingUserPassword",
			req: &pb.LoginRequest{
				Email: "abc@gmail.com",
			},
			err: errorSrv.ErrInvalidPassword,
		},
		{
			name: "ErrUserNotFound",
			req: &pb.LoginRequest{
				Email:    "a@gmail.com",
				Password: "abc123",
			},
			err: errorSrv.ErrUserNotFound,
		},
		{
			name: "ErrIncorrectPassword",
			req: &pb.LoginRequest{
				Email:    "abc@gmail.com",
				Password: "a",
			},
			err: errorSrv.ErrIncorrectPassword,
		},
		{
			name: "Success",
			req: &pb.LoginRequest{
				Email:    "abc@gmail.com",
				Password: "abc123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := s.Login(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp.User)
				require.Empty(t, rsp.Token)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp.User)
				require.Equal(t, rspCreated.User.Id, rsp.User.Id)
				require.Equal(t, rspCreated.User.Email, rsp.User.Email)
				require.NotEmpty(t, rsp.Token)
			}
		})
	}
}

func Test_userServiceImpl_Logout(t *testing.T) {
	type fields struct {
		dal      *postgres.DataAccessLayer
		logger   log.Factory
		tokenSrv *TokenService
	}
	type args struct {
		ctx context.Context
		req *pb.LogoutRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.LogoutResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &userServiceImpl{
				dal:      tt.fields.dal,
				logger:   tt.fields.logger,
				tokenSrv: tt.fields.tokenSrv,
			}
			got, err := u.Logout(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("userServiceImpl.Logout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userServiceImpl.Logout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userServiceImpl_Validate(t *testing.T) {
	type fields struct {
		dal      *postgres.DataAccessLayer
		logger   log.Factory
		tokenSrv *TokenService
	}
	type args struct {
		ctx context.Context
		req *pb.ValidateRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.ValidateResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &userServiceImpl{
				dal:      tt.fields.dal,
				logger:   tt.fields.logger,
				tokenSrv: tt.fields.tokenSrv,
			}
			got, err := u.Validate(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("userServiceImpl.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userServiceImpl.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userServiceImpl_CreateAccount(t *testing.T) {
	s := newUserService(t)
	require.NotNil(t, s)

	// create user
	req := &pb.CreateUserRequest{
		Email:    "abc@gmail.com",
		Password: "abc123",
	}
	rspUserCreated, err := s.Create(context.TODO(), req)
	require.NoError(t, err)
	require.NotNil(t, rspUserCreated.User)
	require.Equal(t, req.Email, rspUserCreated.User.Email)
	require.NotEmpty(t, rspUserCreated.Token)

	tests := []struct {
		name string
		req  *pb.CreateAccountRequest
		err  error
	}{
		{
			name: "ErrMissingUserID",
			req:  &pb.CreateAccountRequest{},
			err:  errorSrv.ErrMissingUserID,
		},
		{
			name: "ErrUserNotFound",
			req: &pb.CreateAccountRequest{
				UserId: 1,
			},
			err: errorSrv.ErrUserNotFound,
		},
		{
			name: "Success",
			req: &pb.CreateAccountRequest{
				UserId:  rspUserCreated.User.Id,
				Name:    rspUserCreated.User.Email,
				Bank:    pb.Bank_ACB,
				Balance: 100000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := s.CreateAccount(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp.Account)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp.Account)
				require.Equal(t, tt.req.UserId, rsp.Account.UserId)
				require.Equal(t, tt.req.Name, rsp.Account.Name)
				require.Equal(t, tt.req.Bank, rsp.Account.Bank)
				require.Equal(t, tt.req.Balance, rsp.Account.Balance)
			}
		})
	}
}

func Test_userServiceImpl_ListAccounts(t *testing.T) {
	type fields struct {
		dal      *postgres.DataAccessLayer
		logger   log.Factory
		tokenSrv *TokenService
	}
	type args struct {
		ctx context.Context
		req *pb.ListAccountsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.ListAccountsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &userServiceImpl{
				dal:      tt.fields.dal,
				logger:   tt.fields.logger,
				tokenSrv: tt.fields.tokenSrv,
			}
			got, err := u.ListAccounts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("userServiceImpl.ListAccounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userServiceImpl.ListAccounts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userServiceImpl_CreateTransaction(t *testing.T) {
	s := newUserService(t)
	require.NotNil(t, s)

	// create user
	req := &pb.CreateUserRequest{
		Email:    "abc@gmail.com",
		Password: "abc123",
	}
	rspUserCreated, err := s.Create(context.TODO(), req)
	require.NoError(t, err)
	require.NotNil(t, rspUserCreated.User)
	require.Equal(t, req.Email, rspUserCreated.User.Email)
	require.NotEmpty(t, rspUserCreated.Token)

	// create account
	reqAcc := &pb.CreateAccountRequest{
		UserId:  rspUserCreated.User.Id,
		Name:    rspUserCreated.User.Email,
		Bank:    pb.Bank_ACB,
		Balance: 10000,
	}
	rspAccCreated, err := s.CreateAccount(context.TODO(), reqAcc)
	require.NoError(t, err)
	require.NotNil(t, rspAccCreated.Account)
	require.Equal(t, reqAcc.UserId, rspAccCreated.Account.UserId)
	require.Equal(t, reqAcc.Name, rspAccCreated.Account.Name)
	require.Equal(t, reqAcc.Bank, rspAccCreated.Account.Bank)
	require.Equal(t, reqAcc.Balance, rspAccCreated.Account.Balance)

	tests := []struct {
		name string
		req  *pb.CreateTransactionRequest
		err  error
	}{
		{
			name: "ErrMissingUserID",
			req:  &pb.CreateTransactionRequest{},
			err:  errorSrv.ErrMissingUserID,
		},
		{
			name: "ErrMissingAccountID",
			req: &pb.CreateTransactionRequest{
				UserId: 1,
			},
			err: errorSrv.ErrMissingAccountID,
		},
		{
			name: "ErrInvalidTransactionAmountGT0",
			req: &pb.CreateTransactionRequest{
				UserId:    1,
				AccountId: 1,
			},
			err: errorSrv.ErrInvalidTransactionAmountGT0,
		},
		{
			name: "ErrAccountNotFound",
			req: &pb.CreateTransactionRequest{
				UserId:    1,
				AccountId: 1,
				Amount:    10000,
			},
			err: errorSrv.ErrAccountNotFound,
		},
		{
			name: "ErrInvalidWithdrawTransactionAmount",
			req: &pb.CreateTransactionRequest{
				UserId:          rspUserCreated.User.Id,
				AccountId:       rspAccCreated.Account.Id,
				Amount:          20000,
				TransactionType: pb.TransactionType_WITHDRAW,
			},
			err: errorSrv.ErrInvalidWithdrawTransactionAmount,
		},
		{
			name: "WithdrawSuccess",
			req: &pb.CreateTransactionRequest{
				UserId:          rspUserCreated.User.Id,
				AccountId:       rspAccCreated.Account.Id,
				Amount:          10000,
				TransactionType: pb.TransactionType_WITHDRAW,
			},
		},
		{
			name: "DepositSuccess",
			req: &pb.CreateTransactionRequest{
				UserId:          rspUserCreated.User.Id,
				AccountId:       rspAccCreated.Account.Id,
				Amount:          50000,
				TransactionType: pb.TransactionType_DEPOSIT,
			},
		},
	}
	accBalance := rspAccCreated.Account.Balance
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := s.CreateTransaction(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp.Transaction)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp.Transaction)
				require.Equal(t, tt.req.AccountId, rsp.Transaction.AccountId)
				require.Equal(t, tt.req.Amount, rsp.Transaction.Amount)
				require.Equal(t, tt.req.TransactionType, rsp.Transaction.TransactionType)
				require.NotEmpty(t, rsp.Transaction.UpdatedAt)
				switch tt.req.TransactionType {
				case pb.TransactionType_WITHDRAW:
					accBalance -= rsp.Transaction.Amount
				case pb.TransactionType_DEPOSIT:
					accBalance += rsp.Transaction.Amount
				}
				// get account
				rspAccs, err := s.ListAccounts(context.TODO(), &pb.ListAccountsRequest{
					UserId: wrapperspb.Int64(tt.req.UserId),
				})
				require.NoError(t, err)
				require.NotNil(t, rspAccs.Account)
				require.Len(t, 1, len(rspAccs.Account))
				for _, acc := range rspAccs.Account {
					if acc.Id == tt.req.AccountId {
						require.Equal(t, accBalance, acc.Balance)
					}
				}
			}
		})
	}
}

func Test_userServiceImpl_ListTransactions(t *testing.T) {
	s := newUserService(t)
	require.NotNil(t, s)

	// create user
	req := &pb.CreateUserRequest{
		Email:    "abc@gmail.com",
		Password: "abc123",
	}
	rspUserCreated, err := s.Create(context.TODO(), req)
	require.NoError(t, err)
	require.NotNil(t, rspUserCreated.User)
	require.Equal(t, req.Email, rspUserCreated.User.Email)
	require.NotEmpty(t, rspUserCreated.Token)

	// create account
	reqAcc := &pb.CreateAccountRequest{
		UserId:  rspUserCreated.User.Id,
		Name:    rspUserCreated.User.Email,
		Bank:    pb.Bank_ACB,
		Balance: 10000,
	}
	rspAccCreated, err := s.CreateAccount(context.TODO(), reqAcc)
	require.NoError(t, err)
	require.NotNil(t, rspAccCreated.Account)
	require.Equal(t, reqAcc.UserId, rspAccCreated.Account.UserId)
	require.Equal(t, reqAcc.Name, rspAccCreated.Account.Name)
	require.Equal(t, reqAcc.Bank, rspAccCreated.Account.Bank)
	require.Equal(t, reqAcc.Balance, rspAccCreated.Account.Balance)

	accBalance := rspAccCreated.Account.Balance

	// create transactions
	reqTrans := []*pb.CreateTransactionRequest{
		{
			UserId:          rspUserCreated.User.Id,
			AccountId:       rspAccCreated.Account.Id,
			Amount:          50000,
			TransactionType: pb.TransactionType_DEPOSIT,
		},
	}
	rspTransCreated := make([]*pb.CreateTransactionResponse, len(reqTrans))
	for _, trans := range reqTrans {
		rsp, err := s.CreateTransaction(context.TODO(), trans)
		require.NoError(t, err)
		require.NotNil(t, rsp.Transaction)
		require.Equal(t, trans.AccountId, rsp.Transaction.AccountId)
		require.Equal(t, trans.Amount, rsp.Transaction.Amount)
		require.Equal(t, trans.TransactionType, rsp.Transaction.TransactionType)
		require.NotEmpty(t, rsp.Transaction.UpdatedAt)
		switch trans.TransactionType {
		case pb.TransactionType_WITHDRAW:
			accBalance -= trans.Amount
		case pb.TransactionType_DEPOSIT:
			accBalance += trans.Amount
		}
		rspTransCreated = append(rspTransCreated, rsp)
	}

	// get account
	rspAccs, err := s.ListAccounts(context.TODO(), &pb.ListAccountsRequest{
		UserId: wrapperspb.Int64(rspUserCreated.User.Id),
	})
	require.NoError(t, err)
	require.NotNil(t, rspAccs.Account)
	require.Len(t, 1, len(rspAccs.Account))
	for _, acc := range rspAccs.Account {
		if acc.Id == rspAccCreated.Account.Id {
			require.Equal(t, accBalance, acc.Balance)
		}
	}

	tests := []struct {
		name  string
		req   *pb.ListTransactionsRequest
		err   error
		len   int
		trans []*pb.CreateTransactionResponse
	}{
		{
			name: "ErrMissingUserID",
			req:  &pb.ListTransactionsRequest{},
			err:  errorSrv.ErrMissingUserID,
		},
		{
			name: "ErrTransactionNotFound",
			req: &pb.ListTransactionsRequest{
				UserId:    1,
				AccountId: 1,
			},
			err: errorSrv.ErrTransactionNotFound,
		},
		{
			name: "SuccessGetAllTransactionsOfUser",
			req: &pb.ListTransactionsRequest{
				UserId: rspUserCreated.User.Id,
			},
			len:   2,
			trans: rspTransCreated,
		},
		{
			name: "SuccessGetTransactionsOfUserAccount",
			req: &pb.ListTransactionsRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated.Account.Id,
			},
			len:   2,
			trans: rspTransCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := s.ListTransactions(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp.Transactions)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp.Transactions)
				require.Equal(t, tt.req.AccountId, rsp.Transactions)
				require.Len(t, tt.len, len(rsp.Transactions))
				for _, trans := range rsp.Transactions {
					for _, tr := range tt.trans {
						if tr.Transaction.Id == trans.Id {
							require.Equal(t, tr.Transaction.AccountId, trans.AccountId)
							require.Equal(t, tr.Transaction.Amount, trans.Amount)
							require.Equal(t, tr.Transaction.TransactionType, trans.TransactionType)
							require.Equal(t, tr.Transaction.CreatedAt, trans.CreatedAt)
						}
					}
				}
			}
		})
	}
}

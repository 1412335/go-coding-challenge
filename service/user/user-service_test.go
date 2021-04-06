package user

import (
	"context"
	"reflect"
	"testing"
	"time"

	pb "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/user"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/configs"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/dal/postgres"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/errors"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	errorSrv "github.com/1412335/moneyforward-go-coding-challenge/service/user/error"
	"github.com/1412335/moneyforward-go-coding-challenge/service/user/model"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func newUserServiceError(t *testing.T) {
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
	}
	// init postgres
	dal, err := postgres.NewDataAccessLayer(context.Background(), config.Database)
	require.Error(t, err)
	require.Nil(t, dal)
}

func newUserService(t *testing.T) pb.UserServiceServer {
	// service configs
	config := configs.ServiceConfig{
		Database: &configs.Database{
			Host:           "localhost",
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
	require.NotNil(t, dal)
	require.NotNil(t, dal.GetDatabase())

	// truncate table
	err = dal.GetDatabase().Exec("TRUNCATE TABLE users, accounts, transactions CASCADE").Error
	require.NoError(t, err)

	// migrate db
	err = dal.GetDatabase().AutoMigrate(
		&model.User{},
		&model.Account{},
		&model.Transaction{},
	)
	require.NoError(t, err)

	// token service
	tokenSrv := NewTokenService(config.JWT)
	require.NotNil(t, tokenSrv)

	// create server
	return NewUserService(dal, tokenSrv)
}

func TestNewUserService_Error(t *testing.T) {
	newUserServiceError(t)
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
			name: "ErrInvalidPasswordLength",
			req: &pb.CreateUserRequest{
				Email:    "abc@gmail.com",
				Password: "abc",
			},
			err: errorSrv.ErrInvalidPassword,
		},
		{
			name: "Success",
			req: &pb.CreateUserRequest{
				Email:    "abc@gmail.com",
				Password: "abc123456",
			},
		},
		{
			name: "DuplicateUserEmail",
			req: &pb.CreateUserRequest{
				Email:    "abc@gmail.com",
				Password: "abc123456",
			},
			err: errorSrv.ErrDuplicateEmail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := s.Create(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp.User)
				require.Equal(t, tt.req.Email, rsp.User.Email)
				require.NotEmpty(t, rsp.Token)
				// fetch response header
				// var header metadata.MD
				// grpc.Header(&header)
				// xrespid := header.Get("X-Http-Code")
				// log.Info("Got response", zap.Strings("X-Http-Code", xrespid))
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
		Password: "abc123456",
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
				Password: "abc123456",
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
				Password: "abc123456",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := s.Login(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp)
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
		Password: "abc123456",
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
			name: "ErrInvalidAccountBalance",
			req: &pb.CreateAccountRequest{
				UserId:  rspUserCreated.User.Id,
				Bank:    pb.Bank_ACB,
				Balance: -100000,
			},
			err: errorSrv.ErrInvalidAccountBalance,
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
				require.Nil(t, rsp)
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
	s := newUserService(t)
	require.NotNil(t, s)

	// create user
	req := &pb.CreateUserRequest{
		Email:    "abc@gmail.com",
		Password: "abc123456",
	}
	rspUserCreated, err := s.Create(context.TODO(), req)
	require.NoError(t, err)
	require.NotNil(t, rspUserCreated.User)
	require.Equal(t, req.Email, rspUserCreated.User.Email)
	require.NotEmpty(t, rspUserCreated.Token)

	// create accounts
	reqAccs := []*pb.CreateAccountRequest{
		{
			UserId:  rspUserCreated.User.Id,
			Name:    rspUserCreated.User.Email,
			Bank:    pb.Bank_ACB,
			Balance: 10000,
		},
		{
			UserId:  rspUserCreated.User.Id,
			Name:    rspUserCreated.User.Email,
			Bank:    pb.Bank_VCB,
			Balance: 20000,
		},
	}
	rspAccCreated := make([]*pb.CreateAccountResponse, len(reqAccs))
	for i, acc := range reqAccs {
		rsp, err := s.CreateAccount(context.TODO(), acc)
		require.NoError(t, err)
		require.NotNil(t, rsp.Account)
		require.Equal(t, acc.UserId, rsp.Account.UserId)
		require.Equal(t, acc.Name, rsp.Account.Name)
		require.Equal(t, acc.Bank, rsp.Account.Bank)
		require.Equal(t, acc.Balance, rsp.Account.Balance)
		rspAccCreated[i] = rsp
	}

	tests := []struct {
		name string
		req  *pb.ListAccountsRequest
		err  error
	}{
		{
			name: "ErrMissingUserID",
			req:  &pb.ListAccountsRequest{},
			err:  errorSrv.ErrMissingUserID,
		},
		{
			name: "ErrUserNotFound",
			req: &pb.ListAccountsRequest{
				UserId: wrapperspb.Int64(1),
			},
			err: errorSrv.ErrUserNotFound,
		},
		{
			name: "Success",
			req: &pb.ListAccountsRequest{
				UserId: wrapperspb.Int64(rspUserCreated.User.Id),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := s.ListAccounts(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp.Accounts)
				require.Len(t, rsp.Accounts, len(rspAccCreated))
				for _, acc := range rsp.Accounts {
					switch acc.Id {
					case rspAccCreated[0].Account.Id:
						require.Equal(t, rspAccCreated[0].Account.Name, acc.Name)
						require.Equal(t, rspAccCreated[0].Account.Bank, acc.Bank)
						require.Equal(t, rspAccCreated[0].Account.Balance, acc.Balance)
					case rspAccCreated[1].Account.Id:
						require.Equal(t, rspAccCreated[1].Account.Name, acc.Name)
						require.Equal(t, rspAccCreated[1].Account.Bank, acc.Bank)
						require.Equal(t, rspAccCreated[1].Account.Balance, acc.Balance)
					}
				}
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
		Password: "abc123456",
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
				require.Nil(t, rsp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp.Transaction)
				require.Equal(t, tt.req.AccountId, rsp.Transaction.AccountId)
				require.Equal(t, tt.req.Amount, rsp.Transaction.Amount)
				require.Equal(t, tt.req.TransactionType, rsp.Transaction.TransactionType)
				require.NotEmpty(t, rsp.Transaction.CreatedAt)
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
				require.NotNil(t, rspAccs.Accounts)
				require.Len(t, rspAccs.Accounts, 1)
				for _, acc := range rspAccs.Accounts {
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
		Password: "abc123456",
	}
	rspUserCreated, err := s.Create(context.TODO(), req)
	require.NoError(t, err)
	require.NotNil(t, rspUserCreated.User)
	require.Equal(t, req.Email, rspUserCreated.User.Email)
	require.NotEmpty(t, rspUserCreated.Token)

	// create accounts
	reqAccs := []*pb.CreateAccountRequest{
		{
			UserId:  rspUserCreated.User.Id,
			Name:    rspUserCreated.User.Email,
			Bank:    pb.Bank_ACB,
			Balance: 10000,
		},
		{
			UserId:  rspUserCreated.User.Id,
			Name:    rspUserCreated.User.Email,
			Bank:    pb.Bank_VCB,
			Balance: 20000,
		},
	}
	rspAccCreated := make([]*pb.CreateAccountResponse, len(reqAccs))
	for i, acc := range reqAccs {
		rsp, err := s.CreateAccount(context.TODO(), acc)
		require.NoError(t, err)
		require.NotNil(t, rsp.Account)
		require.Equal(t, acc.UserId, rsp.Account.UserId)
		require.Equal(t, acc.Name, rsp.Account.Name)
		require.Equal(t, acc.Bank, rsp.Account.Bank)
		require.Equal(t, acc.Balance, rsp.Account.Balance)
		rspAccCreated[i] = rsp
	}

	// get accounts balance
	accBalance := make(map[int64]float64, len(rspAccCreated))
	accBank := make(map[int64]string, len(rspAccCreated))
	for _, rsp := range rspAccCreated {
		accBalance[rsp.Account.Id] = rsp.Account.Balance
		accBank[rsp.Account.Id] = rsp.Account.Bank.String()
	}

	// create transactions
	reqTrans := []*pb.CreateTransactionRequest{
		{
			UserId:          rspUserCreated.User.Id,
			AccountId:       rspAccCreated[0].Account.Id,
			Amount:          50000,
			TransactionType: pb.TransactionType_DEPOSIT,
		},
		{
			UserId:          rspUserCreated.User.Id,
			AccountId:       rspAccCreated[1].Account.Id,
			Amount:          1000,
			TransactionType: pb.TransactionType_WITHDRAW,
		},
	}
	rspTransCreated := make([]*pb.CreateTransactionResponse, len(reqTrans))
	for i, trans := range reqTrans {
		rsp, err := s.CreateTransaction(context.TODO(), trans)
		require.NoError(t, err)
		require.NotNil(t, rsp.Transaction)
		require.Equal(t, trans.AccountId, rsp.Transaction.AccountId)
		require.Equal(t, trans.Amount, rsp.Transaction.Amount)
		require.Equal(t, trans.TransactionType, rsp.Transaction.TransactionType)
		require.NotEmpty(t, rsp.Transaction.CreatedAt)
		_, ok := accBalance[rsp.Transaction.AccountId]
		require.True(t, ok)
		switch trans.TransactionType {
		case pb.TransactionType_WITHDRAW:
			accBalance[rsp.Transaction.AccountId] -= trans.Amount
		case pb.TransactionType_DEPOSIT:
			accBalance[rsp.Transaction.AccountId] += trans.Amount
		}
		rspTransCreated[i] = rsp
	}
	require.Len(t, rspTransCreated, len(reqAccs))

	// get account
	rspAccs, err := s.ListAccounts(context.TODO(), &pb.ListAccountsRequest{
		UserId: wrapperspb.Int64(rspUserCreated.User.Id),
	})
	require.NoError(t, err)
	require.NotNil(t, rspAccs.Accounts)
	require.Len(t, rspAccs.Accounts, len(rspAccCreated))
	for _, acc := range rspAccs.Accounts {
		switch acc.Id {
		case rspAccCreated[0].Account.Id:
			require.Equal(t, accBalance[acc.Id], acc.Balance)
		case rspAccCreated[1].Account.Id:
			require.Equal(t, accBalance[acc.Id], acc.Balance)
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
				AccountId: 2,
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
				AccountId: rspAccCreated[0].Account.Id,
			},
			len:   1,
			trans: rspTransCreated[:1],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := s.ListTransactions(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp.Transactions)
				require.Len(t, rsp.Transactions, tt.len)
				for _, trans := range rsp.Transactions {
					for _, tr := range tt.trans {
						if tr.Transaction.Id == trans.Id {
							require.Equal(t, tr.Transaction.AccountId, trans.AccountId)
							require.Equal(t, tr.Transaction.Amount, trans.Amount)
							require.Equal(t, accBank[tr.Transaction.AccountId], trans.Bank.String())
							require.Equal(t, tr.Transaction.TransactionType, trans.TransactionType)
							require.True(t, trans.CreatedAt.AsTime().Equal(tr.Transaction.CreatedAt.AsTime()))
							break
						}
					}
				}
			}
		})
	}
}

func Test_userServiceImpl_DeleteTransaction(t *testing.T) {
	s := newUserService(t)
	require.NotNil(t, s)

	// create user
	req := &pb.CreateUserRequest{
		Email:    "abc@gmail.com",
		Password: "abc123456",
	}
	rspUserCreated, err := s.Create(context.TODO(), req)
	require.NoError(t, err)
	require.NotNil(t, rspUserCreated.User)
	require.Equal(t, req.Email, rspUserCreated.User.Email)
	require.NotEmpty(t, rspUserCreated.Token)

	// create accounts
	reqAccs := []*pb.CreateAccountRequest{
		{
			UserId:  rspUserCreated.User.Id,
			Name:    rspUserCreated.User.Email,
			Bank:    pb.Bank_ACB,
			Balance: 10000,
		},
	}
	rspAccCreated := make([]*pb.CreateAccountResponse, len(reqAccs))
	for i, acc := range reqAccs {
		rsp, err := s.CreateAccount(context.TODO(), acc)
		require.NoError(t, err)
		require.NotNil(t, rsp.Account)
		require.Equal(t, acc.UserId, rsp.Account.UserId)
		require.Equal(t, acc.Name, rsp.Account.Name)
		require.Equal(t, acc.Bank, rsp.Account.Bank)
		require.Equal(t, acc.Balance, rsp.Account.Balance)
		rspAccCreated[i] = rsp
	}

	// create transactions
	reqTrans := []*pb.CreateTransactionRequest{
		{
			UserId:          rspUserCreated.User.Id,
			AccountId:       rspAccCreated[0].Account.Id,
			Amount:          50000,
			TransactionType: pb.TransactionType_DEPOSIT,
		},
		{
			UserId:          rspUserCreated.User.Id,
			AccountId:       rspAccCreated[0].Account.Id,
			Amount:          10000,
			TransactionType: pb.TransactionType_WITHDRAW,
		},
	}
	rspTransCreated := make([]*pb.CreateTransactionResponse, len(reqTrans))
	for i, trans := range reqTrans {
		rsp, err := s.CreateTransaction(context.TODO(), trans)
		require.NoError(t, err)
		require.NotNil(t, rsp.Transaction)
		require.Equal(t, trans.AccountId, rsp.Transaction.AccountId)
		require.Equal(t, trans.Amount, rsp.Transaction.Amount)
		require.Equal(t, trans.TransactionType, rsp.Transaction.TransactionType)
		require.NotEmpty(t, rsp.Transaction.CreatedAt)
		rspTransCreated[i] = rsp
	}
	require.Len(t, rspTransCreated, len(reqTrans))

	tests := []struct {
		name string
		req  *pb.DeleteTransactionRequest
		err  error
		len  int
		ids  []int64
	}{
		{
			name: "ErrMissingUserID",
			req:  &pb.DeleteTransactionRequest{},
			err:  errorSrv.ErrMissingUserID,
		},
		{
			name: "ErrTransactionNotFound",
			req: &pb.DeleteTransactionRequest{
				UserId:    1,
				AccountId: wrapperspb.Int64(1),
			},
			err: errorSrv.ErrTransactionNotFound,
		},
		{
			name: "DeleteSingleTransSuccess",
			req: &pb.DeleteTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: wrapperspb.Int64(rspAccCreated[0].Account.Id),
				Id:        wrapperspb.Int64(rspTransCreated[0].Transaction.Id),
			},
			len: 1,
			ids: []int64{rspTransCreated[0].Transaction.Id},
		},
		{
			name: "DeleteAllTransOfAccountSuccess",
			req: &pb.DeleteTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: wrapperspb.Int64(rspAccCreated[0].Account.Id),
			},
			len: 1,
			ids: []int64{rspTransCreated[1].Transaction.Id},
		},
		{
			name: "DeleteAllTransOfUserSuccess",
			req: &pb.DeleteTransactionRequest{
				UserId: rspUserCreated.User.Id,
			},
			len: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := s.DeleteTransaction(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp)
				require.Len(t, rsp.Ids, tt.len)
				if tt.len > 0 {
					require.Equal(t, tt.ids, rsp.Ids)
				}
			}
		})
	}
}

func Test_userServiceImpl_UpdateTransaction(t *testing.T) {
	s := newUserService(t)
	require.NotNil(t, s)

	// create user
	req := &pb.CreateUserRequest{
		Email:    "abc@gmail.com",
		Password: "abc123456",
	}
	rspUserCreated, err := s.Create(context.TODO(), req)
	require.NoError(t, err)
	require.NotNil(t, rspUserCreated.User)
	require.Equal(t, req.Email, rspUserCreated.User.Email)
	require.NotEmpty(t, rspUserCreated.Token)

	// create accounts
	reqAccs := []*pb.CreateAccountRequest{
		{
			UserId:  rspUserCreated.User.Id,
			Name:    rspUserCreated.User.Email,
			Bank:    pb.Bank_ACB,
			Balance: 10000,
		},
	}
	rspAccCreated := make([]*pb.CreateAccountResponse, len(reqAccs))
	for i, acc := range reqAccs {
		rsp, err := s.CreateAccount(context.TODO(), acc)
		require.NoError(t, err)
		require.NotNil(t, rsp.Account)
		require.Equal(t, acc.UserId, rsp.Account.UserId)
		require.Equal(t, acc.Name, rsp.Account.Name)
		require.Equal(t, acc.Bank, rsp.Account.Bank)
		require.Equal(t, acc.Balance, rsp.Account.Balance)
		rspAccCreated[i] = rsp
	}

	// create transactions
	reqTrans := []*pb.CreateTransactionRequest{
		{
			UserId:          rspUserCreated.User.Id,
			AccountId:       rspAccCreated[0].Account.Id,
			Amount:          10000,
			TransactionType: pb.TransactionType_WITHDRAW,
		},
		{
			UserId:          rspUserCreated.User.Id,
			AccountId:       rspAccCreated[0].Account.Id,
			Amount:          10000,
			TransactionType: pb.TransactionType_DEPOSIT,
		},
		{
			UserId:          rspUserCreated.User.Id,
			AccountId:       rspAccCreated[0].Account.Id,
			Amount:          10000,
			TransactionType: pb.TransactionType_WITHDRAW,
		},
	}
	rspTransCreated := make([]*pb.CreateTransactionResponse, len(reqTrans))
	for i, trans := range reqTrans {
		rsp, err := s.CreateTransaction(context.TODO(), trans)
		require.NoError(t, err)
		require.NotNil(t, rsp.Transaction)
		require.Equal(t, trans.AccountId, rsp.Transaction.AccountId)
		require.Equal(t, trans.Amount, rsp.Transaction.Amount)
		require.Equal(t, trans.TransactionType, rsp.Transaction.TransactionType)
		require.NotEmpty(t, rsp.Transaction.CreatedAt)
		rspTransCreated[i] = rsp
	}
	require.Len(t, rspTransCreated, len(reqTrans))

	tests := []struct {
		name       string
		req        *pb.UpdateTransactionRequest
		err        error
		len        int
		ids        []int64
		accBalance float64
	}{
		{
			name: "ErrMissingUserID",
			req:  &pb.UpdateTransactionRequest{},
			err:  errorSrv.ErrMissingUserID,
		},
		{
			name: "ErrMissingAccountID",
			req: &pb.UpdateTransactionRequest{
				UserId: 1,
			},
			err: errorSrv.ErrMissingAccountID,
		},
		{
			name: "ErrMissingTransactionID",
			req: &pb.UpdateTransactionRequest{
				UserId:    1,
				AccountId: 1,
			},
			err: errorSrv.ErrMissingTransactionID,
		},
		{
			name: "ErrAccountNotFound",
			req: &pb.UpdateTransactionRequest{
				UserId:    1,
				AccountId: 1,
				Transaction: &pb.Transaction{
					Id: 1,
				},
			},
			err: errorSrv.ErrAccountNotFound,
		},
		{
			name: "ErrTransactionNotFound",
			req: &pb.UpdateTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated[0].Account.Id,
				Transaction: &pb.Transaction{
					Id: 1,
				},
			},
			err: errorSrv.ErrTransactionNotFound,
		},
		{
			name: "ErrInvalidWithdrawTransactionAmount",
			req: &pb.UpdateTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated[0].Account.Id,
				Transaction: &pb.Transaction{
					Id:     rspTransCreated[0].Transaction.Id,
					Amount: 100000000,
				},
			},
			err: errorSrv.ErrInvalidWithdrawTransactionAmount,
		},
		{
			name: "ErrInvalidWithdrawTransactionAmountWithMask",
			req: &pb.UpdateTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated[0].Account.Id,
				Transaction: &pb.Transaction{
					Id:     rspTransCreated[0].Transaction.Id,
					Amount: 100000000,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"amount"},
				},
			},
			err: errorSrv.ErrInvalidWithdrawTransactionAmount,
		},
		{
			name: "ErrInvalidTransactionAmountGT0",
			req: &pb.UpdateTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated[0].Account.Id,
				Transaction: &pb.Transaction{
					Id:     rspTransCreated[1].Transaction.Id,
					Amount: 1000,
				},
			},
			err: errorSrv.ErrInvalidTransactionAmountGT0,
		},
		{
			name: "ErrInvalidTransactionAmountGT0WithMask",
			req: &pb.UpdateTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated[0].Account.Id,
				Transaction: &pb.Transaction{
					Id:     rspTransCreated[1].Transaction.Id,
					Amount: 1000,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"amount"},
				},
			},
			err: errorSrv.ErrInvalidTransactionAmountGT0,
		},
		{
			name: "ErrInvalidUpdateTransactionID",
			req: &pb.UpdateTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated[0].Account.Id,
				Transaction: &pb.Transaction{
					Id:     rspTransCreated[1].Transaction.Id,
					Amount: 10000,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"id"},
				},
			},
			err: errors.BadRequest("cannot update id", map[string]string{"update_mask": "cannot update id field"}),
		},
		{
			name: "ErrInvalidUpdateTransactionType",
			req: &pb.UpdateTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated[0].Account.Id,
				Transaction: &pb.Transaction{
					Id:     rspTransCreated[1].Transaction.Id,
					Amount: 10000,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"transaction_type"},
				},
			},
			err: errors.BadRequest("cannot update transaction type", map[string]string{"update_mask": "cannot update transaction_type"}),
		},
		{
			name: "UpdateWithdrawTransSuccess",
			req: &pb.UpdateTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated[0].Account.Id,
				Transaction: &pb.Transaction{
					Id:     rspTransCreated[0].Transaction.Id,
					Amount: 5000,
				},
			},
			accBalance: 5000,
		},
		{
			name: "UpdateDepositTransSuccess",
			req: &pb.UpdateTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated[0].Account.Id,
				Transaction: &pb.Transaction{
					Id:     rspTransCreated[1].Transaction.Id,
					Amount: 10000,
				},
			},
			accBalance: 5000,
		},
		{
			name: "UpdateDepositTransWithMaskSuccess",
			req: &pb.UpdateTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated[0].Account.Id,
				Transaction: &pb.Transaction{
					Id:     rspTransCreated[1].Transaction.Id,
					Amount: 50000,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"amount"},
				},
			},
			accBalance: 45000,
		},
		{
			name: "UpdateWithdrawTransWithMaskSuccess",
			req: &pb.UpdateTransactionRequest{
				UserId:    rspUserCreated.User.Id,
				AccountId: rspAccCreated[0].Account.Id,
				Transaction: &pb.Transaction{
					Id:     rspTransCreated[2].Transaction.Id,
					Amount: 50000,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"amount"},
				},
			},
			accBalance: 5000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := s.UpdateTransaction(context.TODO(), tt.req)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				require.Nil(t, rsp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rsp)
				require.Equal(t, tt.req.Transaction.Id, rsp.Transaction.Id)
				require.Equal(t, tt.req.AccountId, rsp.Transaction.AccountId)
				require.Equal(t, tt.req.Transaction.Amount, rsp.Transaction.Amount)
				//
				accs, err := s.ListAccounts(context.TODO(), &pb.ListAccountsRequest{
					UserId: wrapperspb.Int64(tt.req.UserId),
					Id:     wrapperspb.Int64(rsp.Transaction.AccountId),
				})
				require.NoError(t, err)
				require.NotNil(t, accs)
				require.NotNil(t, accs.Accounts[0])
				require.Equal(t, tt.accBalance, accs.Accounts[0].Balance)
			}
		})
	}
}

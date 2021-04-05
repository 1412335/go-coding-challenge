package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/structs"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	// pbAccount "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/account"
	// pbTransaction "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/transaction"
	pb "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/user"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/dal/postgres"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/errors"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/utils"
)

var (
	ErrMissingEmail      = errors.BadRequest("MISSING_EMAIL", map[string]string{"email": "Missing email"})
	ErrDuplicateEmail    = errors.BadRequest("DUPLICATE_EMAIL", map[string]string{"email": "A user with this email address already exists"})
	ErrInvalidEmail      = errors.BadRequest("INVALID_EMAIL", map[string]string{"email": "The email provided is invalid"})
	ErrInvalidPassword   = errors.BadRequest("INVALID_PASSWORD", map[string]string{"password": "Password must be at least 8 characters long"})
	ErrIncorrectPassword = errors.Unauthenticated("INCORRECT_PASSWORD", "password", "Email or password is incorrect")
	ErrMissingID         = errors.BadRequest("MISSING_ID", map[string]string{"id": "Missing user id"})
	ErrMissingToken      = errors.BadRequest("MISSING_TOKEN", map[string]string{"token": "Missing token"})

	ErrHashPassword = errors.InternalServerError("HASH_PASSWORD", "hash password failed")

	ErrConnectDB = errors.InternalServerError("CONNECT_DB", "Connecting to database failed")
	ErrNotFound  = errors.NotFound("NOT_FOUND", map[string]string{"user": "User not found"})

	ErrTokenGenerated = errors.InternalServerError("TOKEN_GEN_FAILED", "Generate token failed")
	ErrTokenInvalid   = errors.Unauthenticated("TOKEN_INVALID", "token", "Token invalid")
	// ErrTokenNotFound  = errors.BadRequest("TOKEN_NOT_FOUND", "Token not found")
	// ErrTokenExpired   = errors.Unauthorized("TOKEN_EXPIRE", "Token expired")
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
func (u *userServiceImpl) getUserByID(ctx context.Context, id int64) (*User, error) {
	user := &User{}
	err := u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// find user by id
		if e := tx.Where(&User{ID: id}).First(user).Error; e == gorm.ErrRecordNotFound {
			return ErrNotFound
		} else if e != nil {
			u.logger.For(ctx).Error("Find user", zap.Error(e))
			return ErrConnectDB
		}
		// cache
		if e := user.cache(); e != nil {
			u.logger.For(ctx).Error("Cache user", zap.Error(e))
		}
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
		return nil, ErrInvalidEmail
	}
	if !isValidPassword(req.GetPassword()) {
		return nil, ErrInvalidPassword
	}

	user := &User{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	if err := user.validate(); err != nil {
		u.logger.For(ctx).Error("Error validate user", zap.Error(err))
		return nil, err
	}

	// init response
	rsp := &pb.CreateUserResponse{}

	// create
	return rsp, u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil && strings.Contains(err.Error(), "idx_users_email") {
			return ErrDuplicateEmail
		} else if err != nil {
			u.logger.For(ctx).Error("Error connecting from db", zap.Error(err))
			return ErrConnectDB
		}

		// create token
		token, err := u.tokenSrv.Generate(user)
		if err != nil {
			u.logger.For(ctx).Error("Error generate token", zap.Error(err))
			return ErrTokenGenerated
		}

		rsp.User = user.transform2GRPC()
		rsp.Token = token
		return nil
	})
}

// delete user by id
func (u *userServiceImpl) Delete(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if req.GetId() == 0 {
		return nil, ErrMissingID
	}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(req.GetId()).Delete(&User{}).Error; err == gorm.ErrRecordNotFound {
			return ErrNotFound
		} else if err != nil {
			u.logger.For(ctx).Error("Error connecting from db", zap.Error(err))
			return ErrConnectDB
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
		return nil, ErrMissingID
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
			user.updateFromGRPC(req.GetUser())
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
			return ErrInvalidEmail
		}
		if !isValidPassword(user.Password) {
			return ErrInvalidPassword
		}
		if err := user.validate(); err != nil {
			u.logger.For(ctx).Error("Error validate user", zap.Error(err))
			return err
		}
		// update user in db
		if e := tx.Save(user).Error; e != nil && strings.Contains(e.Error(), "idx_users_email") {
			return ErrDuplicateEmail
		} else if e != nil {
			return ErrConnectDB
		}
		// response
		rsp.User = user.transform2GRPC()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}

// build query statement & get list users
func (u *userServiceImpl) getUsers(ctx context.Context, req *pb.ListUsersRequest) ([]*pb.User, error) {
	var users []User
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
		return nil, ErrConnectDB
	}
	// check empty from db
	if len(users) == 0 {
		st := status.New(codes.NotFound, "not found users")
		des, err := st.WithDetails(&errdetails.PreconditionFailure{
			Violations: []*errdetails.PreconditionFailure_Violation{
				{
					Type:        "USER",
					Subject:     "no users",
					Description: "no users have been found",
				},
			},
		})
		if err != nil {
			return nil, des.Err()
		}
		return nil, st.Err()
	}
	// filter
	rsp := make([]*pb.User, len(users))
	for i, user := range users {
		rsp[i] = user.transform2GRPC()
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
		return nil, ErrMissingEmail
	}
	if !isValidEmail(req.GetEmail()) {
		return nil, ErrInvalidEmail
	}
	if len(req.GetPassword()) == 0 {
		return nil, ErrInvalidPassword
	}
	// response
	rsp := &pb.LoginResponse{}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		var user User
		// find user by email
		if e := tx.Where(&User{Email: strings.ToLower(req.GetEmail())}).First(&user).Error; e == gorm.ErrRecordNotFound {
			return ErrNotFound
		} else if e != nil {
			u.logger.For(ctx).Error("Error find user", zap.Error(e))
			return ErrConnectDB
		}
		// verify password
		if e := utils.CompareHash(user.Password, req.GetPassword()); e != nil {
			return ErrIncorrectPassword
		}
		// gen new token
		token, e := u.tokenSrv.Generate(&user)
		if e != nil {
			u.logger.For(ctx).Error("Error gen token", zap.Error(e))
			return ErrTokenGenerated
		}
		// cache user
		if e := user.cache(); e != nil {
			u.logger.For(ctx).Error("Cache user", zap.Error(e))
		}
		//
		rsp.User = user.transform2GRPC()
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
		return nil, ErrMissingID
	}
	return nil, nil
}

// validate token: update isActive=true & return user
func (u *userServiceImpl) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	if len(req.GetToken()) == 0 {
		return nil, ErrMissingToken
	}
	rsp := &pb.ValidateResponse{}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		// verrify token
		claims, e := u.tokenSrv.Verify(req.Token)
		if e != nil {
			u.logger.For(ctx).Error("verify token failed", zap.Error(e))
			return ErrTokenInvalid
		}
		// update active
		if e = tx.Model(&User{ID: claims.ID}).Update("active", true).Error; e == gorm.ErrRecordNotFound {
			return ErrNotFound
		} else if e != nil {
			u.logger.For(ctx).Error("Error update user", zap.Error(e))
			return ErrConnectDB
		}
		// get cache user
		user, e := u.getUserByID(ctx, claims.ID)
		if e != nil {
			u.logger.For(ctx).Error("Get user by ID", zap.Error(e))
			return errors.InternalServerError("Get user failed", "Lookup user by ID w redis/db failed")
		}
		rsp.User = user.transform2GRPC()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}

// accounts
func (u *userServiceImpl) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAccount not implemented")
}
func (u *userServiceImpl) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAccounts not implemented")
}

// transactions
func (u *userServiceImpl) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTransaction not implemented")
}
func (u *userServiceImpl) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTransactions not implemented")
}

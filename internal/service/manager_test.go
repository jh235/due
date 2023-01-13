package service_test

import (
	"context"
	"github.com/dobyte/due/errors"
	"github.com/dobyte/due/internal/service"
	"testing"
)

type User struct {
	UID      int64
	Age      int
	Nickname string
}

type UserService struct {
	users map[int64]*User
}

type AddUserArgs struct {
	UID      int64
	Age      int
	Nickname string
}

type AddUserReply struct {
}

type GetUserArgs struct {
	UID int64
}

type GetUserReply struct {
	User *User
}

type ModifyNicknameArgs struct {
	UID      int64
	Nickname string
}

type ModifyNicknameReply struct {
	User *User
}

func NewUserService() *UserService {
	return &UserService{
		users: make(map[int64]*User),
	}
}

// ServiceName 更改服务名称
func (s *UserService) ServiceName() string {
	return "user"
}

// 添加用户
func (s *UserService) AddUser(ctx context.Context, args *AddUserArgs) (*AddUserReply, error) {
	s.users[args.UID] = &User{
		UID:      args.UID,
		Age:      args.Age,
		Nickname: args.Nickname,
	}

	return &AddUserReply{}, nil
}

// GetUser 获取用户
func (s *UserService) GetUser(ctx context.Context, args *GetUserArgs) (*GetUserReply, error) {
	user, ok := s.users[args.UID]
	if !ok {
		return nil, errors.New("not found user")
	}

	return &GetUserReply{
		User: user,
	}, nil
}

// ModifyNickname 修改用户昵称
func (s *UserService) ModifyNickname(ctx context.Context, args *ModifyNicknameArgs) (*ModifyNicknameReply, error) {
	user, ok := s.users[args.UID]
	if !ok {
		return nil, errors.New("not found user")
	}

	user.Nickname = args.Nickname

	return &ModifyNicknameReply{
		User: user,
	}, nil
}

func TestManager_Call(t *testing.T) {
	m := service.NewManager()
	m.Register(NewUserService())

	ctx := context.Background()

	_, err := m.Call(ctx, "user", "AddUser", &AddUserArgs{
		UID:      1,
		Age:      31,
		Nickname: "fuxiao",
	})
	if err != nil {
		t.Fatal(err)
	}

	reply1, err := m.Call(ctx, "user", "GetUser", &GetUserArgs{
		UID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	if r, ok := reply1.(*GetUserReply); ok {
		t.Logf("%+v", r.User)
	}

	reply2, err := m.Call(ctx, "user", "ModifyNickname", &ModifyNicknameArgs{
		UID:      1,
		Nickname: "yuebanfuxiao",
	})
	if err != nil {
		t.Fatal(err)
	}

	if r, ok := reply2.(*ModifyNicknameReply); ok {
		t.Logf("%+v", r.User)
	}

	reply3, err := m.Call(ctx, "user", "GetUser", &GetUserArgs{
		UID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	if r, ok := reply3.(*GetUserReply); ok {
		t.Logf("%+v", r.User)
	}
}

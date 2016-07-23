package auth

import (
	"github.com/Dataman-Cloud/rolex/model"
)

type Authenticator interface {
	AccountPermissions(account *Account) (*[]string, error)
	AccountGroups(account *Account) (*[]Group, error)

	Login(account *Account) (token string, err error)
	EncryptPassword(password string) string

	DeleteGroup(groupId uint64) error
	Groups(listOptions model.ListOptions) (*[]Group, error)
	Group(id uint64) (*Group, error)
	CreateGroup(role *Group) error
	UpdateGroup(role *Group) error

	Accounts(listOptions model.ListOptions) (*[]Account, error)
	Account(id interface{}) (*Account, error)
	CreateAccount(a *Account) error
	UpdateAccount(a *Account) error

	JoinGroup(accountId, groupId uint64) error
	LeaveGroup(accountId, groupId uint64) error

	ModificationAllowed() bool
	GroupOperationAllowed(accountId, groupId uint64) bool
}

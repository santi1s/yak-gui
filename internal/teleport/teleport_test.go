package teleport

import (
	"reflect"
	"testing"
)

func TestTeleportRolesToRequest(t *testing.T) {
	var tConfig = Config{
		Version: "14.3.6",
		Cluster: "teleport.doctolib.net",
		Accounts: map[string]Account{
			"0000": {
				Name: "foo",
				Roles: []Role{
					{
						Name:       "db-superuser-foo",
						Type:       "db",
						Permission: "superadministrator",
					},
					{
						Name:       "aws-foo-superadministrator",
						Type:       "aws",
						Permission: "superadministrator",
					},
				},
			},
			"1111": {
				Name: "bar",
				Roles: []Role{
					{
						Name:       "aws-bar-superadministrator",
						Type:       "aws",
						Permission: "superadministrator",
					},
				},
			},
			"2222": {
				Name: "baz",
				Roles: []Role{
					{
						Name:       "aws-baz-superadministrator",
						Type:       "aws",
						Permission: "superadministrator",
					},
				},
			},
		},
	}

	type args struct {
		tConfig           Config
		permission        string
		tshRoles          []string
		requestedAccounts []string
		roleType          string
	}
	tests := []struct {
		name        string
		args        args
		wantToReq   []Role
		wantAssumed []Role
	}{
		{
			name: "valid test",
			args: args{
				permission:        "superadministrator",
				tshRoles:          []string{"aws-foo-superadministrator"},
				requestedAccounts: []string{"baz", "foo"},
				roleType:          AWSConfigType,
				tConfig:           tConfig,
			},
			wantToReq: []Role{
				{
					Name:       "aws-baz-superadministrator",
					Type:       "aws",
					Permission: "superadministrator",
					AccountNo:  "2222",
				},
			},
			wantAssumed: []Role{
				{
					Name:       "aws-foo-superadministrator",
					Type:       "aws",
					Permission: "superadministrator",
					AccountNo:  "0000",
				},
			},
		},
		{
			name: "valid test #2",
			args: args{
				permission:        "superadministrator",
				tshRoles:          []string{"aws-foo-superadministrator", "aws-bar-superadministrator"},
				requestedAccounts: []string{"bar", "foo"},
				roleType:          AWSConfigType,
				tConfig:           tConfig,
			},
			wantToReq: []Role{},
			wantAssumed: []Role{
				{
					Name:       "aws-bar-superadministrator",
					Type:       "aws",
					Permission: "superadministrator",
					AccountNo:  "1111",
				},
				{
					Name:       "aws-foo-superadministrator",
					Type:       "aws",
					Permission: "superadministrator",
					AccountNo:  "0000",
				},
			},
		},
		{
			name: "valid test #3",
			args: args{
				permission:        "superadministrator",
				tshRoles:          []string{},
				requestedAccounts: []string{"bar", "baz"},
				roleType:          AWSConfigType,
				tConfig:           tConfig,
			},
			wantToReq: []Role{
				{
					Name:       "aws-bar-superadministrator",
					Type:       "aws",
					Permission: "superadministrator",
					AccountNo:  "1111",
				},
				{
					Name:       "aws-baz-superadministrator",
					Type:       "aws",
					Permission: "superadministrator",
					AccountNo:  "2222",
				},
			},
			wantAssumed: []Role{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToReq, gotAssumed := TeleportRolesToRequest(tt.args.tConfig, tt.args.permission, tt.args.tshRoles, tt.args.requestedAccounts, tt.args.roleType)
			if !reflect.DeepEqual(gotToReq, tt.wantToReq) {
				t.Errorf("TeleportRolesToRequest() gotToReq = %v, wantToReq %v", gotToReq, tt.wantToReq)
			}
			if !reflect.DeepEqual(gotAssumed, tt.wantAssumed) {
				t.Errorf("TeleportRolesToRequest() gotAssumed = %v, wantAssumed %v", gotAssumed, tt.wantAssumed)
			}
		})
	}
}

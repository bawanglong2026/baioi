package adminauthzhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	admindomain "github.com/dujiao-next/internal/modules/identity/admin/domain"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

type accountHandlerAdminDirectoryStub struct {
	admins      map[uint]*admindomain.Admin
	nextID      uint
	createCalls int
	updateCalls int
}

func newAccountHandlerAdminDirectoryStub(admins ...*admindomain.Admin) *accountHandlerAdminDirectoryStub {
	stub := &accountHandlerAdminDirectoryStub{
		admins: make(map[uint]*admindomain.Admin),
		nextID: 1,
	}
	for _, admin := range admins {
		copy := *admin
		if copy.ID == 0 {
			copy.ID = stub.nextID
		}
		if copy.ID >= stub.nextID {
			stub.nextID = copy.ID + 1
		}
		stub.admins[copy.ID] = &copy
	}
	return stub
}

func (s *accountHandlerAdminDirectoryStub) List() ([]admindomain.Admin, error) {
	admins := make([]admindomain.Admin, 0, len(s.admins))
	for _, admin := range s.admins {
		admins = append(admins, *admin)
	}
	return admins, nil
}

func (s *accountHandlerAdminDirectoryStub) GetByID(id uint) (*admindomain.Admin, error) {
	admin := s.admins[id]
	if admin == nil {
		return nil, nil
	}
	copy := *admin
	return &copy, nil
}

func (s *accountHandlerAdminDirectoryStub) GetByUsername(username string) (*admindomain.Admin, error) {
	for _, admin := range s.admins {
		if admin.Username == username {
			copy := *admin
			return &copy, nil
		}
	}
	return nil, nil
}

func (s *accountHandlerAdminDirectoryStub) Create(admin *admindomain.Admin) error {
	s.createCalls++
	if admin.ID == 0 {
		admin.ID = s.nextID
		s.nextID++
	}
	copy := *admin
	s.admins[copy.ID] = &copy
	return nil
}

func (s *accountHandlerAdminDirectoryStub) Update(admin *admindomain.Admin) error {
	s.updateCalls++
	copy := *admin
	s.admins[copy.ID] = &copy
	return nil
}

func (s *accountHandlerAdminDirectoryStub) Delete(uint) error { return nil }

func (s *accountHandlerAdminDirectoryStub) Count() (int64, error) {
	return int64(len(s.admins)), nil
}

type accountHandlerPasswordStub struct{}

func (accountHandlerPasswordStub) ValidatePassword(string) error { return nil }

func (accountHandlerPasswordStub) HashPassword(string) (string, error) { return "hashed", nil }

type accountHandlerAuthStateStub struct{}

func (accountHandlerAuthStateStub) SetAdminAuthState(context.Context, *admindomain.Admin) error {
	return nil
}

func (accountHandlerAuthStateStub) DelAdminAuthState(context.Context, uint) error { return nil }

type accountHandlerRolePolicyStub struct{}

func (accountHandlerRolePolicyStub) GetAdminRoles(uint) ([]string, error) { return nil, nil }

func (accountHandlerRolePolicyStub) GetAdminPolicies(uint) ([]Policy, error) { return nil, nil }

func (accountHandlerRolePolicyStub) ListRoles() ([]string, error) { return nil, nil }

func (accountHandlerRolePolicyStub) EnsureRole(role string) (string, error) { return role, nil }

func (accountHandlerRolePolicyStub) DeleteRole(string) error { return nil }

func (accountHandlerRolePolicyStub) GetRolePolicies(string) ([]Policy, error) { return nil, nil }

func (accountHandlerRolePolicyStub) GrantRolePolicy(string, string, string) error { return nil }

func (accountHandlerRolePolicyStub) RevokeRolePolicy(string, string, string) error { return nil }

func (accountHandlerRolePolicyStub) SetAdminRoles(uint, []string) error { return nil }

func newAccountHandlerTestHandler(directory AdminDirectory) *AdminHandler {
	return NewAdminHandler(
		accountHandlerRolePolicyStub{},
		directory,
		accountHandlerPasswordStub{},
		accountHandlerAuthStateStub{},
		nil,
	)
}

func newAccountHandlerTestContext(t *testing.T, method, path, body string, operator *admindomain.Admin) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if method == http.MethodPut {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		c.Params = gin.Params{{Key: "id", Value: parts[len(parts)-1]}}
	}
	c.Set("admin_id", operator.ID)
	c.Set("username", operator.Username)
	c.Set("admin_is_super", operator.IsSuper)
	return c, recorder
}

func decodeAccountHandlerResponse(t *testing.T, recorder *httptest.ResponseRecorder) response.Response {
	t.Helper()
	var payload response.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, recorder.Body.String())
	}
	return payload
}

func TestCreateAuthzAdminRequiresSuperAdminForSuperAccount(t *testing.T) {
	operator := &admindomain.Admin{ID: 1, Username: "operator", IsSuper: false}

	t.Run("regular admin can create regular admin", func(t *testing.T) {
		directory := newAccountHandlerAdminDirectoryStub(operator)
		handler := newAccountHandlerTestHandler(directory)
		c, recorder := newAccountHandlerTestContext(t, http.MethodPost, "/admin/authz/admins", `{"username":"normal-user","password":"strong-password"}`, operator)

		handler.CreateAuthzAdmin(c)

		payload := decodeAccountHandlerResponse(t, recorder)
		if payload.StatusCode != response.CodeOK {
			t.Fatalf("status_code = %d, want %d", payload.StatusCode, response.CodeOK)
		}
		if directory.createCalls != 1 {
			t.Fatalf("create calls = %d, want 1", directory.createCalls)
		}
		created, _ := directory.GetByUsername("normal-user")
		if created == nil || created.IsSuper {
			t.Fatalf("created admin = %#v, want regular admin", created)
		}
	})

	t.Run("regular admin cannot create super admin", func(t *testing.T) {
		directory := newAccountHandlerAdminDirectoryStub(operator)
		handler := newAccountHandlerTestHandler(directory)
		c, recorder := newAccountHandlerTestContext(t, http.MethodPost, "/admin/authz/admins", `{"username":"super-user","password":"strong-password","is_super":true}`, operator)

		handler.CreateAuthzAdmin(c)

		payload := decodeAccountHandlerResponse(t, recorder)
		if payload.StatusCode != response.CodeForbidden {
			t.Fatalf("status_code = %d, want %d", payload.StatusCode, response.CodeForbidden)
		}
		if directory.createCalls != 0 {
			t.Fatalf("create calls = %d, want 0", directory.createCalls)
		}
	})

	t.Run("regular admin cannot create protected super admin by username", func(t *testing.T) {
		directory := newAccountHandlerAdminDirectoryStub(operator)
		handler := newAccountHandlerTestHandler(directory)
		c, recorder := newAccountHandlerTestContext(t, http.MethodPost, "/admin/authz/admins", `{"username":"ADMIN","password":"strong-password"}`, operator)

		handler.CreateAuthzAdmin(c)

		payload := decodeAccountHandlerResponse(t, recorder)
		if payload.StatusCode != response.CodeForbidden {
			t.Fatalf("status_code = %d, want %d", payload.StatusCode, response.CodeForbidden)
		}
		if directory.createCalls != 0 {
			t.Fatalf("create calls = %d, want 0", directory.createCalls)
		}
	})

	t.Run("super admin can create super admin", func(t *testing.T) {
		superOperator := &admindomain.Admin{ID: 1, Username: "admin", IsSuper: true}
		directory := newAccountHandlerAdminDirectoryStub(superOperator)
		handler := newAccountHandlerTestHandler(directory)
		c, recorder := newAccountHandlerTestContext(t, http.MethodPost, "/admin/authz/admins", `{"username":"super-user","password":"strong-password","is_super":true}`, superOperator)

		handler.CreateAuthzAdmin(c)

		payload := decodeAccountHandlerResponse(t, recorder)
		if payload.StatusCode != response.CodeOK {
			t.Fatalf("status_code = %d, want %d", payload.StatusCode, response.CodeOK)
		}
		created, _ := directory.GetByUsername("super-user")
		if created == nil || !created.IsSuper {
			t.Fatalf("created admin = %#v, want super admin", created)
		}
	})
}

func TestUpdateAuthzAdminRequiresSuperAdminForSuperFlagChanges(t *testing.T) {
	regularOperator := &admindomain.Admin{ID: 1, Username: "operator", IsSuper: false}
	normalTarget := &admindomain.Admin{ID: 2, Username: "normal-user", PasswordHash: "old-hash"}
	superTarget := &admindomain.Admin{ID: 3, Username: "super-user", IsSuper: true, PasswordHash: "old-hash"}

	t.Run("regular admin can update regular admin without changing super flag", func(t *testing.T) {
		directory := newAccountHandlerAdminDirectoryStub(regularOperator, normalTarget)
		handler := newAccountHandlerTestHandler(directory)
		c, recorder := newAccountHandlerTestContext(t, http.MethodPut, "/admin/authz/admins/2", `{"username":"normal-renamed","is_super":false}`, regularOperator)

		handler.UpdateAuthzAdmin(c)

		payload := decodeAccountHandlerResponse(t, recorder)
		if payload.StatusCode != response.CodeOK {
			t.Fatalf("status_code = %d, want %d", payload.StatusCode, response.CodeOK)
		}
		if directory.updateCalls != 1 {
			t.Fatalf("update calls = %d, want 1", directory.updateCalls)
		}
		updated, _ := directory.GetByID(2)
		if updated == nil || updated.Username != "normal-renamed" || updated.IsSuper {
			t.Fatalf("updated admin = %#v, want renamed regular admin", updated)
		}
	})

	t.Run("regular admin cannot elevate regular admin", func(t *testing.T) {
		directory := newAccountHandlerAdminDirectoryStub(regularOperator, normalTarget)
		handler := newAccountHandlerTestHandler(directory)
		c, recorder := newAccountHandlerTestContext(t, http.MethodPut, "/admin/authz/admins/2", `{"is_super":true}`, regularOperator)

		handler.UpdateAuthzAdmin(c)

		payload := decodeAccountHandlerResponse(t, recorder)
		if payload.StatusCode != response.CodeForbidden {
			t.Fatalf("status_code = %d, want %d", payload.StatusCode, response.CodeForbidden)
		}
		if directory.updateCalls != 0 {
			t.Fatalf("update calls = %d, want 0", directory.updateCalls)
		}
	})

	t.Run("regular admin cannot demote super admin", func(t *testing.T) {
		directory := newAccountHandlerAdminDirectoryStub(regularOperator, superTarget)
		handler := newAccountHandlerTestHandler(directory)
		c, recorder := newAccountHandlerTestContext(t, http.MethodPut, "/admin/authz/admins/3", `{"is_super":false}`, regularOperator)

		handler.UpdateAuthzAdmin(c)

		payload := decodeAccountHandlerResponse(t, recorder)
		if payload.StatusCode != response.CodeForbidden {
			t.Fatalf("status_code = %d, want %d", payload.StatusCode, response.CodeForbidden)
		}
		if directory.updateCalls != 0 {
			t.Fatalf("update calls = %d, want 0", directory.updateCalls)
		}
	})

	t.Run("super admin can elevate regular admin", func(t *testing.T) {
		superOperator := &admindomain.Admin{ID: 1, Username: "admin", IsSuper: true}
		directory := newAccountHandlerAdminDirectoryStub(superOperator, normalTarget)
		handler := newAccountHandlerTestHandler(directory)
		c, recorder := newAccountHandlerTestContext(t, http.MethodPut, "/admin/authz/admins/2", `{"is_super":true}`, superOperator)

		handler.UpdateAuthzAdmin(c)

		payload := decodeAccountHandlerResponse(t, recorder)
		if payload.StatusCode != response.CodeOK {
			t.Fatalf("status_code = %d, want %d", payload.StatusCode, response.CodeOK)
		}
		updated, _ := directory.GetByID(2)
		if updated == nil || !updated.IsSuper {
			t.Fatalf("updated admin = %#v, want super admin", updated)
		}
	})

	t.Run("super admin can demote non-protected super admin", func(t *testing.T) {
		superOperator := &admindomain.Admin{ID: 1, Username: "admin", IsSuper: true}
		directory := newAccountHandlerAdminDirectoryStub(superOperator, superTarget)
		handler := newAccountHandlerTestHandler(directory)
		c, recorder := newAccountHandlerTestContext(t, http.MethodPut, "/admin/authz/admins/3", `{"is_super":false}`, superOperator)

		handler.UpdateAuthzAdmin(c)

		payload := decodeAccountHandlerResponse(t, recorder)
		if payload.StatusCode != response.CodeOK {
			t.Fatalf("status_code = %d, want %d", payload.StatusCode, response.CodeOK)
		}
		updated, _ := directory.GetByID(3)
		if updated == nil || updated.IsSuper {
			t.Fatalf("updated admin = %#v, want regular admin", updated)
		}
	})
}

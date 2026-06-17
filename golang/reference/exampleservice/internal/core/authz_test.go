package core_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db"
)

func ctxWith(p core.Principal) context.Context {
	return core.WithPrincipal(context.Background(), p)
}

// TestServiceRequiresPrincipal proves every operation fails closed when the
// context carries no authenticated principal: ErrUnauthenticated, never an
// empty-tenant action.
func TestServiceRequiresPrincipal(t *testing.T) {
	svc := newService()
	ctx := context.Background() // no principal

	if _, err := svc.CreateWidget(ctx, "w1", "n"); !errors.Is(err, core.ErrUnauthenticated) {
		t.Errorf("CreateWidget = %v, want ErrUnauthenticated", err)
	}
	if _, err := svc.GetWidget(ctx, "w1"); !errors.Is(err, core.ErrUnauthenticated) {
		t.Errorf("GetWidget = %v, want ErrUnauthenticated", err)
	}
	if _, err := svc.ListWidgetsPage(ctx, core.Cursor{}, 10); !errors.Is(err, core.ErrUnauthenticated) {
		t.Errorf("ListWidgetsPage = %v, want ErrUnauthenticated", err)
	}
}

// TestServiceRequiresRole proves the boundary RBAC check: a reader-only
// principal cannot create, and a principal with no roles cannot read.
func TestServiceRequiresRole(t *testing.T) {
	svc := newService()

	readerOnly := ctxWith(core.Principal{Subject: "r", TenantID: "t1", Roles: []core.Role{core.RoleReader}})
	if _, err := svc.CreateWidget(readerOnly, "w1", "n"); !errors.Is(err, core.ErrForbidden) {
		t.Errorf("CreateWidget as reader = %v, want ErrForbidden", err)
	}

	noRoles := ctxWith(core.Principal{Subject: "n", TenantID: "t1"})
	if _, err := svc.GetWidget(noRoles, "w1"); !errors.Is(err, core.ErrForbidden) {
		t.Errorf("GetWidget with no roles = %v, want ErrForbidden", err)
	}
	if _, err := svc.ListWidgetsPage(noRoles, core.Cursor{}, 10); !errors.Is(err, core.ErrForbidden) {
		t.Errorf("ListWidgetsPage with no roles = %v, want ErrForbidden", err)
	}
}

// TestServiceCrossTenantRead proves the per-resource ownership check: a widget
// created by tenant t1 is not readable by tenant t2 — the cross-tenant read is
// a not-found, never another tenant's row.
func TestServiceCrossTenantRead(t *testing.T) {
	base := time.Unix(1700000000, 0).UTC()
	store := db.NewMemory()
	svc := core.NewService(store, fixedClock{t: base})

	t1 := ctxWith(core.Principal{Subject: "a", TenantID: "t1", Roles: []core.Role{core.RoleReader, core.RoleWriter}})
	t2 := ctxWith(core.Principal{Subject: "b", TenantID: "t2", Roles: []core.Role{core.RoleReader, core.RoleWriter}})

	if _, err := svc.CreateWidget(t1, "w1", "t1 widget"); err != nil {
		t.Fatalf("CreateWidget(t1): %v", err)
	}

	// t1 reads its own widget; the create stamps the principal's tenant, so the
	// body cannot smuggle a different tenant.
	if _, err := svc.GetWidget(t1, "w1"); err != nil {
		t.Fatalf("GetWidget(t1): %v", err)
	}

	// t2 cannot see t1's widget: not found.
	if _, err := svc.GetWidget(t2, "w1"); !errors.Is(err, core.ErrNotFound) {
		t.Errorf("GetWidget(t2) = %v, want ErrNotFound (cross-tenant isolation)", err)
	}

	// t2's list is empty.
	page, err := svc.ListWidgetsPage(t2, core.Cursor{}, 10)
	if err != nil {
		t.Fatalf("ListWidgetsPage(t2): %v", err)
	}
	if len(page.Widgets) != 0 {
		t.Errorf("t2 list len = %d, want 0", len(page.Widgets))
	}
}

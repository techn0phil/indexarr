package repository

import (
	"errors"
	"testing"
)

func TestUserRepository_CreateGetList(t *testing.T) {
	db := setupTestDBWithMigrations(t)
	repo := NewUserRepository(db)

	created, err := repo.Create("alice", "hash-alice", "admin")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created user to have non-zero id")
	}
	if !created.Enabled {
		t.Fatalf("expected created user enabled=true")
	}

	byID, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if byID.Username != "alice" || byID.Role != "admin" {
		t.Fatalf("unexpected user from GetByID: username=%q role=%q", byID.Username, byID.Role)
	}

	byUsername, err := repo.GetByUsername("alice")
	if err != nil {
		t.Fatalf("GetByUsername returned error: %v", err)
	}
	if byUsername.ID != created.ID {
		t.Fatalf("expected matching ids, got %d vs %d", byUsername.ID, created.ID)
	}

	if _, err := repo.Create("bob", "hash-bob", "guest"); err != nil {
		t.Fatalf("failed to create second user: %v", err)
	}

	users, err := repo.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Username != "alice" || users[1].Username != "bob" {
		t.Fatalf("expected users sorted by username, got %q then %q", users[0].Username, users[1].Username)
	}
}

func TestUserRepository_CreateDuplicateUsername(t *testing.T) {
	db := setupTestDBWithMigrations(t)
	repo := NewUserRepository(db)

	if _, err := repo.Create("alice", "hash", "admin"); err != nil {
		t.Fatalf("failed to create initial user: %v", err)
	}
	if _, err := repo.Create("alice", "other-hash", "guest"); !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestUserRepository_UpdateAndConflict(t *testing.T) {
	db := setupTestDBWithMigrations(t)
	repo := NewUserRepository(db)

	first, err := repo.Create("alice", "hash-a", "admin")
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}
	second, err := repo.Create("bob", "hash-b", "guest")
	if err != nil {
		t.Fatalf("failed to create second user: %v", err)
	}

	disabled := false
	updated, err := repo.Update(first.ID, "alice-renamed", "guest", &disabled)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Username != "alice-renamed" || updated.Role != "guest" || updated.Enabled {
		t.Fatalf("unexpected updated user fields: username=%q role=%q enabled=%v", updated.Username, updated.Role, updated.Enabled)
	}

	if _, err := repo.Update(second.ID, "alice-renamed", "", nil); !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists for username conflict, got %v", err)
	}
}

func TestUserRepository_UpdatePasswordDeleteAndCountAdmins(t *testing.T) {
	db := setupTestDBWithMigrations(t)
	repo := NewUserRepository(db)

	admin, err := repo.Create("admin1", "hash-admin", "admin")
	if err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}
	guest, err := repo.Create("guest1", "hash-guest", "guest")
	if err != nil {
		t.Fatalf("failed to create guest: %v", err)
	}

	disabled := false
	if _, err := repo.Update(admin.ID, "", "", &disabled); err != nil {
		t.Fatalf("failed to disable admin: %v", err)
	}

	admin2, err := repo.Create("admin2", "hash-admin2", "admin")
	if err != nil {
		t.Fatalf("failed to create second admin: %v", err)
	}

	adminCount, err := repo.CountAdmins()
	if err != nil {
		t.Fatalf("CountAdmins returned error: %v", err)
	}
	if adminCount != 1 {
		t.Fatalf("expected 1 enabled admin, got %d", adminCount)
	}

	if err := repo.UpdatePassword(guest.ID, "hash-guest-updated"); err != nil {
		t.Fatalf("UpdatePassword returned error: %v", err)
	}
	guestLoaded, err := repo.GetByID(guest.ID)
	if err != nil {
		t.Fatalf("GetByID guest returned error: %v", err)
	}
	if guestLoaded.PasswordHash != "hash-guest-updated" {
		t.Fatalf("password hash was not updated")
	}

	if err := repo.Delete(admin2.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := repo.GetByID(admin2.ID); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound after delete, got %v", err)
	}
}

func TestUserRepository_NotFoundErrors(t *testing.T) {
	db := setupTestDBWithMigrations(t)
	repo := NewUserRepository(db)

	if _, err := repo.GetByID(999); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound for GetByID, got %v", err)
	}
	if _, err := repo.GetByUsername("does-not-exist"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound for GetByUsername, got %v", err)
	}
	if err := repo.UpdatePassword(999, "hash"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound for UpdatePassword, got %v", err)
	}
	if err := repo.Delete(999); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound for Delete, got %v", err)
	}
}

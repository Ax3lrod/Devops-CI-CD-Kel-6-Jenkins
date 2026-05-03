package service_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/taskflow/api/internal/model"
	"github.com/taskflow/api/internal/repository"
	"github.com/taskflow/api/internal/service"
)

func newSvc() *service.TaskService {
	return service.NewTaskService(repository.NewMemoryRepository())
}

// ── [BUG] CalculateCompletionRate ────────────────────────────────────────────
// BUG #1: Integer division — hasil selalu 0 (kecuali semua task selesai).

func TestCalculateCompletionRate(t *testing.T) {
	tests := []struct {
		name  string
		tasks []model.Task
		want  float64
		isBug bool
	}{
		{
			name:  "tidak ada task",
			tasks: []model.Task{},
			want:  0,
		},
		{
			name:  "semua done → 100%",
			tasks: []model.Task{{Status: model.StatusDone}, {Status: model.StatusDone}},
			want:  100.0,
		},
		{
			// [BUG] 1/3 dengan integer division = 0, bukan 33.33
			name: "[BUG] sepertiga selesai → 33.33%",
			tasks: []model.Task{
				{Status: model.StatusDone},
				{Status: model.StatusTodo},
				{Status: model.StatusTodo},
			},
			want:  33.33,
			isBug: true,
		},
		{
			// [BUG] 1/2 dengan integer division = 0, bukan 50.0
			name:  "[BUG] setengah selesai → 50%",
			tasks: []model.Task{{Status: model.StatusDone}, {Status: model.StatusTodo}},
			want:  50.0,
			isBug: true,
		},
		{
			name: "tidak ada yang selesai → 0%",
			tasks: []model.Task{
				{Status: model.StatusTodo},
				{Status: model.StatusInProgress},
			},
			want: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := service.CalculateCompletionRate(tc.tasks)
			// Toleransi 0.01 untuk floating point
			diff := got - tc.want
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.01 {
				if tc.isBug {
					t.Errorf("BUG TERDETEKSI — CalculateCompletionRate() = %.2f, want %.2f\n"+
						"  → Integer division: %d/%d = 0 (bukan %.2f)\n"+
						"  → Perbaiki: gunakan float64(completed)/float64(len(tasks))*100",
						got, tc.want, len(tc.tasks)/2, len(tc.tasks), tc.want)
				} else {
					t.Errorf("CalculateCompletionRate() = %.2f, want %.2f", got, tc.want)
				}
			}
		})
	}
}

func TestCalculateCompletionRate_PartialCompletion(t *testing.T) {
	tasks := []model.Task{
		{ID: "1", Status: model.StatusDone},
		{ID: "2", Status: model.StatusTodo},
		{ID: "3", Status: model.StatusTodo},
	}
	// 1 dari 3 = 33.33...
	rate := service.CalculateCompletionRate(tasks)
	if rate < 33.0 || rate > 34.0 {
		t.Errorf("expected ~33.33, got %v", rate)
	}
}

// ── Create ───────────────────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
	svc := newSvc()

	t.Run("sukses dengan default priority", func(t *testing.T) {
		task, err := svc.Create(model.CreateTaskRequest{Title: "Belajar Go"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if task.Title != "Belajar Go" {
			t.Errorf("Title = %q, want %q", task.Title, "Belajar Go")
		}
		if task.Status != model.StatusTodo {
			t.Errorf("Status = %q, want todo", task.Status)
		}
		if task.Priority != model.PriorityMedium {
			t.Errorf("Priority = %q, want medium (default)", task.Priority)
		}
		if task.ID == "" {
			t.Error("ID tidak boleh kosong")
		}
	})

	t.Run("title kosong ditolak", func(t *testing.T) {
		_, err := svc.Create(model.CreateTaskRequest{Title: ""})
		if err == nil {
			t.Error("Create() harus error jika title kosong")
		}
	})

	t.Run("title spasi saja ditolak", func(t *testing.T) {
		_, err := svc.Create(model.CreateTaskRequest{Title: "   "})
		if err == nil {
			t.Error("Create() harus error jika title hanya spasi")
		}
	})

	t.Run("priority invalid ditolak", func(t *testing.T) {
		_, err := svc.Create(model.CreateTaskRequest{Title: "T", Priority: "extreme"})
		if err == nil {
			t.Error("Create() harus error untuk priority tidak valid")
		}
	})

	t.Run("priority high sukses", func(t *testing.T) {
		task, err := svc.Create(model.CreateTaskRequest{Title: "Urgent", Priority: model.PriorityHigh})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if task.Priority != model.PriorityHigh {
			t.Errorf("Priority = %q, want high", task.Priority)
		}
	})

	t.Run("setiap task ID unik", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 50; i++ {
			task, _ := svc.Create(model.CreateTaskRequest{Title: "Task"})
			if ids[task.ID] {
				t.Errorf("ID duplikat ditemukan: %s", task.ID)
			}
			ids[task.ID] = true
		}
	})
}

// ── Update ───────────────────────────────────────────────────────────────────

func TestUpdate(t *testing.T) {
	svc := newSvc()

	t.Run("update status ke done mengisi completed_at", func(t *testing.T) {
		task, _ := svc.Create(model.CreateTaskRequest{Title: "Selesaikan"})
		statusDone := model.StatusDone
		updated, err := svc.Update(task.ID, model.UpdateTaskRequest{Status: &statusDone})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.CompletedAt == nil {
			t.Error("CompletedAt harus terisi setelah status = done")
		}
	})

	t.Run("update task tidak ada → error", func(t *testing.T) {
		statusDone := model.StatusDone
		_, err := svc.Update("id-tidak-ada", model.UpdateTaskRequest{Status: &statusDone})
		if err == nil {
			t.Error("Update() harus error untuk ID tidak ada")
		}
	})

	t.Run("update status invalid → error", func(t *testing.T) {
		task, _ := svc.Create(model.CreateTaskRequest{Title: "T"})
		s := model.Status("invalid")
		_, err := svc.Update(task.ID, model.UpdateTaskRequest{Status: &s})
		if err == nil {
			t.Error("Update() harus error untuk status tidak valid")
		}
	})
}

// ── [CICD] Full Task Lifecycle ────────────────────────────────────────────────
// [CICD] Simulasi integration test: create → get → update → delete.
// Jenis test ini dijalankan otomatis setelah deploy ke staging.

func TestTaskFullLifecycle(t *testing.T) {
	svc := newSvc()

	// 1. Create
	task, err := svc.Create(model.CreateTaskRequest{
		Title:    "Pipeline Lifecycle Test",
		Priority: model.PriorityHigh,
	})
	if err != nil {
		t.Fatalf("Create() gagal: %v", err)
	}

	// 2. Get
	got, err := svc.GetByID(task.ID)
	if err != nil || got.ID != task.ID {
		t.Fatalf("GetByID() gagal setelah create")
	}

	// 3. Update ke in_progress
	s := model.StatusInProgress
	got, err = svc.Update(task.ID, model.UpdateTaskRequest{Status: &s})
	if err != nil || got.Status != model.StatusInProgress {
		t.Fatalf("Update() ke in_progress gagal")
	}

	// 4. Update ke done
	done := model.StatusDone
	got, err = svc.Update(task.ID, model.UpdateTaskRequest{Status: &done})
	if err != nil || got.CompletedAt == nil {
		t.Fatalf("Update() ke done gagal atau CompletedAt nil")
	}

	// 5. Stats harus menunjukkan 1 done
	stats, err := svc.GetStats()
	if err != nil {
		t.Fatalf("GetStats() gagal: %v", err)
	}
	if stats.ByStatus["done"] != 1 {
		t.Errorf("Stats.ByStatus[done] = %d, want 1", stats.ByStatus["done"])
	}

	// 6. Delete
	_, err = svc.Delete(task.ID)
	if err != nil {
		t.Fatalf("Delete() gagal: %v", err)
	}

	// 7. Pastikan sudah terhapus
	if _, err = svc.GetByID(task.ID); err == nil {
		t.Error("GetByID() harus error setelah task dihapus")
	}
}

// ── [CICD] Rollback Simulation ───────────────────────────────────────────────

func TestRollbackStatusSimulation(t *testing.T) {
	svc := newSvc()
	task, _ := svc.Create(model.CreateTaskRequest{Title: "Rollback Test"})

	// Simulasi: deploy berhasil → update ke in_progress
	s := model.StatusInProgress
	svc.Update(task.ID, model.UpdateTaskRequest{Status: &s}) //nolint

	// Deployment bermasalah → rollback ke todo
	todo := model.StatusTodo
	rolled, err := svc.Update(task.ID, model.UpdateTaskRequest{Status: &todo})
	if err != nil {
		t.Fatalf("Rollback gagal: %v", err)
	}
	if rolled.Status != model.StatusTodo {
		t.Errorf("Setelah rollback, status = %q, want todo", rolled.Status)
	}
}

// ── [TODO] Tambahkan test berikut ─────────────────────────────────────────────
// TODO mahasiswa:
// - TestGetAll_WithStatusFilter (setelah bug #2 diperbaiki)
// - TestGetStats_CompletionRate (setelah bug #1 diperbaiki)
// - TestCreate_WithUnicodeTitle
// - TestDelete_AndVerifyStats

func TestGetAll(t *testing.T) {
	svc := newSvc()

	// Kosong dulu
	tasks, err := svc.GetAll("") // GetAll butuh string filter
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}

	// Tambah 2 task
	svc.Create(model.CreateTaskRequest{Title: "Task A"})
	svc.Create(model.CreateTaskRequest{Title: "Task B"})

	tasks, err = svc.GetAll("")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestGetAll_WithStatusFilter(t *testing.T) {
	svc := newSvc()

	svc.Create(model.CreateTaskRequest{Title: "Todo Task"})
	svc.Create(model.CreateTaskRequest{Title: "Done Task"})

	// Kita perlu update salah satu jadi done secara manual karena Create selalu todo
	tasks, _ := svc.GetAll("")
	svc.Update(tasks[1].ID, model.UpdateTaskRequest{Status: ptr(model.StatusDone)})

	t.Run("Filter done", func(t *testing.T) {
		results, err := svc.GetAll("done")
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 done task, got %d", len(results))
		}
	})

	t.Run("Filter tidak valid", func(t *testing.T) {
		_, err := svc.GetAll("invalid-status")
		if err == nil {
			t.Error("expected error for invalid status filter, got nil")
		}
	})
}

func TestUpdate_PartialFields(t *testing.T) {
	svc := newSvc()
	task, _ := svc.Create(model.CreateTaskRequest{Title: "Old Title", Description: "Old Desc"})

	t.Run("Update Title Only", func(t *testing.T) {
		newTitle := "New Title"
		updated, _ := svc.Update(task.ID, model.UpdateTaskRequest{Title: &newTitle})
		if updated.Title != newTitle {
			t.Errorf("title not updated: got %q", updated.Title)
		}
		if updated.Description != "Old Desc" {
			t.Error("description should not change")
		}
	})

	t.Run("Update Description Only", func(t *testing.T) {
		newDesc := "New Description"
		updated, _ := svc.Update(task.ID, model.UpdateTaskRequest{Description: &newDesc})
		if updated.Description != newDesc {
			t.Errorf("description not updated: got %q", updated.Description)
		}
	})
}

func ptr[T any](v T) *T {
	return &v
}

// ── Mock Repository untuk Error Path Testing ─────────────────────────────────

type errorRepo struct {
	base       *repository.MemoryRepository
	saveErr    error
	findErr    error
	deleteErr  error
	findAllErr error
}

func (r *errorRepo) Save(t model.Task) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	return r.base.Save(t)
}

func (r *errorRepo) FindByID(id string) (model.Task, bool, error) {
	if r.findErr != nil {
		return model.Task{}, false, r.findErr
	}
	return r.base.FindByID(id)
}

func (r *errorRepo) FindAll() ([]model.Task, error) {
	if r.findAllErr != nil {
		return nil, r.findAllErr
	}
	return r.base.FindAll()
}

func (r *errorRepo) FindByStatus(s model.Status) ([]model.Task, error) {
	if r.findAllErr != nil {
		return nil, r.findAllErr
	}
	return r.base.FindByStatus(s)
}

func (r *errorRepo) Delete(id string) (bool, error) {
	if r.deleteErr != nil {
		return false, r.deleteErr
	}
	return r.base.Delete(id)
}

func (r *errorRepo) Count() (int, error) {
	return r.base.Count()
}

func (r *errorRepo) Close() error {
	return r.base.Close()
}

// ── Error Path Tests ─────────────────────────────────────────────────────────

func TestCreate_TitleTooLong(t *testing.T) {
	svc := newSvc()
	longTitle := strings.Repeat("a", 201)
	_, err := svc.Create(model.CreateTaskRequest{Title: longTitle})
	if err == nil {
		t.Error("expected error for title > 200 chars")
	}
}

func TestCreate_InvalidPriority(t *testing.T) {
	svc := newSvc()
	_, err := svc.Create(model.CreateTaskRequest{Title: "Test", Priority: "urgent"})
	if err == nil {
		t.Error("expected error for invalid priority 'urgent'")
	}
}

func TestCreate_SaveError(t *testing.T) {
	repo := &errorRepo{
		base:    repository.NewMemoryRepository(),
		saveErr: fmt.Errorf("database down"),
	}
	svc := service.NewTaskService(repo)
	_, err := svc.Create(model.CreateTaskRequest{Title: "Test"})
	if err == nil {
		t.Error("expected error when save fails")
	}
}

func TestGetByID_DatabaseError(t *testing.T) {
	repo := &errorRepo{
		base:    repository.NewMemoryRepository(),
		findErr: fmt.Errorf("connection refused"),
	}
	svc := service.NewTaskService(repo)
	_, err := svc.GetByID("any-id")
	if err == nil {
		t.Error("expected error on database failure")
	}
}

func TestDelete_NotFoundService(t *testing.T) {
	svc := newSvc()
	_, err := svc.Delete("non-existent")
	if err == nil {
		t.Error("expected error for deleting non-existent task")
	}
}

func TestDelete_DatabaseError(t *testing.T) {
	repo := &errorRepo{
		base:    repository.NewMemoryRepository(),
		findErr: fmt.Errorf("connection refused"),
	}
	svc := service.NewTaskService(repo)
	_, err := svc.Delete("any-id")
	if err == nil {
		t.Error("expected error on database failure")
	}
}

func TestDelete_RepoDeleteError(t *testing.T) {
	repo := &errorRepo{
		base:      repository.NewMemoryRepository(),
		deleteErr: fmt.Errorf("delete failed"),
	}
	svc := service.NewTaskService(repo)
	// Simpan task dulu supaya FindByID berhasil
	repo.base.Save(model.Task{ID: "1", Title: "Test", Status: model.StatusTodo, Priority: "medium"})
	_, err := svc.Delete("1")
	if err == nil {
		t.Error("expected error when repo.Delete fails")
	}
}

func TestUpdate_DatabaseError(t *testing.T) {
	repo := &errorRepo{
		base:    repository.NewMemoryRepository(),
		findErr: fmt.Errorf("connection refused"),
	}
	svc := service.NewTaskService(repo)
	newTitle := "New"
	_, err := svc.Update("any-id", model.UpdateTaskRequest{Title: &newTitle})
	if err == nil {
		t.Error("expected error on database failure")
	}
}

func TestUpdate_SaveError(t *testing.T) {
	repo := &errorRepo{
		base:    repository.NewMemoryRepository(),
		saveErr: fmt.Errorf("save failed"),
	}
	svc := service.NewTaskService(repo)
	// Simpan langsung ke base (bypass mock saveErr)
	repo.base.Save(model.Task{ID: "1", Title: "Old", Status: model.StatusTodo, Priority: "medium"})
	newTitle := "New Title"
	_, err := svc.Update("1", model.UpdateTaskRequest{Title: &newTitle})
	if err == nil {
		t.Error("expected error when save fails during update")
	}
}

func TestUpdate_EmptyTitle(t *testing.T) {
	svc := newSvc()
	task, _ := svc.Create(model.CreateTaskRequest{Title: "Original"})
	emptyTitle := ""
	_, err := svc.Update(task.ID, model.UpdateTaskRequest{Title: &emptyTitle})
	if err == nil {
		t.Error("expected error for empty title update")
	}
}

func TestUpdate_InvalidStatus(t *testing.T) {
	svc := newSvc()
	task, _ := svc.Create(model.CreateTaskRequest{Title: "Test"})
	_, err := svc.Update(task.ID, model.UpdateTaskRequest{Status: ptr(model.Status("invalid"))})
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestGetStats_DatabaseError(t *testing.T) {
	repo := &errorRepo{
		base:       repository.NewMemoryRepository(),
		findAllErr: fmt.Errorf("connection refused"),
	}
	svc := service.NewTaskService(repo)
	_, err := svc.GetStats()
	if err == nil {
		t.Error("expected error on database failure")
	}
}

func TestDelete(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewTaskService(repo)

	// Buat task dulu
	created, _ := svc.Create(model.CreateTaskRequest{Title: "Hapus ini"})

	// Delete yang ada — harus return task + nil error
	deletedTask, err := svc.Delete(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if deletedTask.ID != created.ID {
		t.Errorf("expected deleted task ID %s, got %s", created.ID, deletedTask.ID)
	}

	// Delete yang tidak ada — harus return error
	_, err = svc.Delete("tidak-ada")
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

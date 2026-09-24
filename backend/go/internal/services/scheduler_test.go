package services

import (
	"errors"
	"testing"

	"indexarr/internal/config"
	"indexarr/internal/models"

	"github.com/rs/zerolog"
)

type fakeMovieImporter struct {
	count        int
	countErr     error
	importResult *models.ScanResult
	importErr    error
	status       *models.ScanStatus
	statusErr    error
	stopped      bool
	lastMovieID  int64
	lastCtx      *models.ProgressContext
}

func (f *fakeMovieImporter) Import(ctx *models.ProgressContext) (*models.ScanResult, error) {
	f.lastCtx = ctx
	return f.importResult, f.importErr
}
func (f *fakeMovieImporter) ImportMovie(id int64) (*models.ScanResult, error) {
	f.lastMovieID = id
	return f.importResult, f.importErr
}
func (f *fakeMovieImporter) GetPendingFileCount() (int, error) { return f.count, f.countErr }
func (f *fakeMovieImporter) GetStatus() (*models.ScanStatus, error) {
	if f.statusErr != nil {
		return nil, f.statusErr
	}
	if f.status == nil {
		return &models.ScanStatus{Status: "idle"}, nil
	}
	return f.status, nil
}
func (f *fakeMovieImporter) Stop()           { f.stopped = true }
func (f *fakeMovieImporter) IsRunning() bool { return false }

type fakeSeriesImporter struct {
	count        int
	countErr     error
	importResult *models.ScanResult
	importErr    error
	status       *models.ScanStatus
	statusErr    error
	stopped      bool
	lastSeriesID int64
	lastCtx      *models.ProgressContext
}

func (f *fakeSeriesImporter) Import(ctx *models.ProgressContext) (*models.ScanResult, error) {
	f.lastCtx = ctx
	return f.importResult, f.importErr
}
func (f *fakeSeriesImporter) ImportSeries(id int64) (*models.ScanResult, error) {
	f.lastSeriesID = id
	return f.importResult, f.importErr
}
func (f *fakeSeriesImporter) GetPendingFileCount() (int, error) { return f.count, f.countErr }
func (f *fakeSeriesImporter) GetStatus() (*models.ScanStatus, error) {
	if f.statusErr != nil {
		return nil, f.statusErr
	}
	if f.status == nil {
		return &models.ScanStatus{Status: "idle"}, nil
	}
	return f.status, nil
}
func (f *fakeSeriesImporter) Stop()           { f.stopped = true }
func (f *fakeSeriesImporter) IsRunning() bool { return false }

func TestSchedulerGetMode(t *testing.T) {
	logger := zerolog.Nop()
	config.GlobalLogger = &logger

	db := setupServicesTestDBWithMigrations(t)

	if got := NewScheduler(db, nil, nil, nil, 0).GetMode(); got != "disabled" {
		t.Fatalf("GetMode() = %q, want %q", got, "disabled")
	}
	if got := NewScheduler(db, &fakeMovieImporter{}, nil, nil, 0).GetMode(); got != "movies" {
		t.Fatalf("GetMode() = %q, want %q", got, "movies")
	}
	if got := NewScheduler(db, nil, &fakeSeriesImporter{}, nil, 0).GetMode(); got != "series" {
		t.Fatalf("GetMode() = %q, want %q", got, "series")
	}
	if got := NewScheduler(db, &fakeMovieImporter{}, &fakeSeriesImporter{}, nil, 0).GetMode(); got != "dual" {
		t.Fatalf("GetMode() = %q, want %q", got, "dual")
	}
}

func TestSchedulerTriggerScan_CoordinatesImporters(t *testing.T) {
	logger := zerolog.Nop()
	config.GlobalLogger = &logger

	db := setupServicesTestDBWithMigrations(t)

	movie := &fakeMovieImporter{
		count: 2,
		importResult: &models.ScanResult{
			FilesProcessed: 2,
			MoviesAdded:    1,
			Errors:         []string{"movie warning"},
		},
	}
	series := &fakeSeriesImporter{
		count: 3,
		importResult: &models.ScanResult{
			FilesProcessed: 3,
			EpisodesAdded:  2,
			Errors:         []string{"series warning"},
		},
	}

	s := NewScheduler(db, movie, series, nil, 0)
	result, err := s.TriggerScan()
	if err != nil {
		t.Fatalf("TriggerScan() error = %v", err)
	}

	if result.FilesFound != 5 || result.FilesProcessed != 5 {
		t.Fatalf("unexpected file counters: %+v", result)
	}
	if result.MoviesAdded != 1 || result.EpisodesAdded != 2 {
		t.Fatalf("unexpected added counters: %+v", result)
	}
	if len(result.Errors) != 2 {
		t.Fatalf("expected merged errors, got %+v", result.Errors)
	}

	if movie.lastCtx == nil || !movie.lastCtx.SuppressStartComplete || movie.lastCtx.TotalOverride != 5 || movie.lastCtx.Offset != 0 {
		t.Fatalf("unexpected movie progress context: %+v", movie.lastCtx)
	}
	if series.lastCtx == nil || !series.lastCtx.SuppressStartComplete || series.lastCtx.TotalOverride != 5 || series.lastCtx.Offset != 2 {
		t.Fatalf("unexpected series progress context: %+v", series.lastCtx)
	}
}

func TestSchedulerStopCurrentScan_StopsBothImporters(t *testing.T) {
	movie := &fakeMovieImporter{}
	series := &fakeSeriesImporter{}
	s := NewScheduler(nil, movie, series, nil, 0)

	s.StopCurrentScan()

	if !movie.stopped || !series.stopped {
		t.Fatalf("expected both importers to receive stop signal")
	}
}

func TestSchedulerGetScanStatus_PrecedenceAndErrors(t *testing.T) {
	running := &fakeMovieImporter{status: &models.ScanStatus{Status: "running", FilesFound: 7}}
	series := &fakeSeriesImporter{status: &models.ScanStatus{Status: "idle"}}
	s := NewScheduler(nil, running, series, nil, 0)

	status, err := s.GetScanStatus()
	if err != nil {
		t.Fatalf("GetScanStatus() error = %v", err)
	}
	if status.Status != "running" || status.FilesFound != 7 {
		t.Fatalf("unexpected status: %+v", status)
	}

	s = NewScheduler(nil, &fakeMovieImporter{statusErr: errors.New("boom")}, nil, nil, 0)
	if _, err := s.GetScanStatus(); err == nil {
		t.Fatalf("expected status error")
	}
}

func TestSchedulerSingleAndSubsetTriggers(t *testing.T) {
	movie := &fakeMovieImporter{importResult: &models.ScanResult{FilesProcessed: 1}}
	series := &fakeSeriesImporter{importResult: &models.ScanResult{FilesProcessed: 2}}
	s := NewScheduler(nil, movie, series, nil, 0)

	if _, err := s.TriggerMoviesScan(); err != nil {
		t.Fatalf("TriggerMoviesScan() error = %v", err)
	}
	if _, err := s.TriggerSeriesScan(); err != nil {
		t.Fatalf("TriggerSeriesScan() error = %v", err)
	}

	if _, err := s.TriggerSingleMovieScan(42); err != nil {
		t.Fatalf("TriggerSingleMovieScan() error = %v", err)
	}
	if movie.lastMovieID != 42 {
		t.Fatalf("expected movie id 42, got %d", movie.lastMovieID)
	}

	if _, err := s.TriggerSingleSeriesScan(77); err != nil {
		t.Fatalf("TriggerSingleSeriesScan() error = %v", err)
	}
	if series.lastSeriesID != 77 {
		t.Fatalf("expected series id 77, got %d", series.lastSeriesID)
	}
}

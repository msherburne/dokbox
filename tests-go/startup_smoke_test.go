package testsgo

import (
	"github.com/msherburne/dokbox/internal/app"
	"strings"
	"testing"
)

func TestNewModelBootstrapsIntoReadyShell(t *testing.T) {
	model := app.NewModel(app.Dependencies{})
	if model == nil {
		t.Fatal("expected model")
	}

	initialView := model.View()
	if !strings.Contains(initialView, "Bootstrapping shell") {
		t.Fatalf("expected bootstrap view, got %q", initialView)
	}

	initCmd := model.Init()
	if initCmd == nil {
		t.Fatal("expected startup init command")
	}

	startupMsg := initCmd()
	if startupMsg == nil {
		t.Fatal("expected startup init command to emit a message")
	}

	nextModel, nextCmd := model.Update(startupMsg)
	if nextCmd != nil {
		t.Fatal("expected bootstrap update to finish without follow-up command")
	}

	readyView := nextModel.View()
	if !strings.Contains(readyView, "Starting rewrite shell") {
		t.Fatalf("expected ready shell view, got %q", readyView)
	}

	if strings.Contains(readyView, "Bootstrapping shell") {
		t.Fatalf("expected bootstrap copy to clear after startup, got %q", readyView)
	}
}

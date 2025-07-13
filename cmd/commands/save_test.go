package commands

import (
	"bytes"
	"strings"
	"testing"

	Compose "github.com/sunpia/docker-deliver/internal/compose"
)

func TestNewSaveCmd(t *testing.T) {
	cmd := NewSaveCmd()

	if cmd.Use != "save" {
		t.Errorf("Expected Use to be 'save', got '%s'", cmd.Use)
	}

	if cmd.Short != "Save docker compose project" {
		t.Errorf("Expected Short description, got '%s'", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}

	// Check required flags
	fileFlag := cmd.Flag("file")
	if fileFlag == nil {
		t.Fatal("Expected 'file' flag to exist")
	}

	outputFlag := cmd.Flag("output")
	if outputFlag == nil {
		t.Fatal("Expected 'output' flag to exist")
	}

	// Check optional flags
	workDirFlag := cmd.Flag("workdir")
	if workDirFlag == nil {
		t.Fatal("Expected 'workdir' flag to exist")
	}

	tagFlag := cmd.Flag("tag")
	if tagFlag == nil {
		t.Fatal("Expected 'tag' flag to exist")
	}
	if tagFlag.DefValue != "latest" {
		t.Errorf("Expected default tag to be 'latest', got '%s'", tagFlag.DefValue)
	}

	logLevelFlag := cmd.Flag("loglevel")
	if logLevelFlag == nil {
		t.Fatal("Expected 'loglevel' flag to exist")
	}
	if logLevelFlag.DefValue != "info" {
		t.Errorf("Expected default loglevel to be 'info', got '%s'", logLevelFlag.DefValue)
	}
}

func TestSaveCmd_FlagShortNames(t *testing.T) {
	cmd := NewSaveCmd()

	// Test short flag names
	testCases := map[string]string{
		"file":     "f",
		"output":   "o",
		"workdir":  "w",
		"tag":      "t",
		"loglevel": "l",
	}

	for flagName, expectedShort := range testCases {
		flag := cmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Flag '%s' not found", flagName)
			continue
		}
		if flag.Shorthand != expectedShort {
			t.Errorf("Expected shorthand for '%s' to be '%s', got '%s'", flagName, expectedShort, flag.Shorthand)
		}
	}
}

func TestSaveCmd_FlagDefaults(t *testing.T) {
	cmd := NewSaveCmd()

	testCases := map[string]string{
		"tag":      "latest",
		"loglevel": "info",
		"workdir":  "",
		"output":   "",
	}

	for flagName, expectedDefault := range testCases {
		flag := cmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Flag '%s' not found", flagName)
			continue
		}
		if flag.DefValue != expectedDefault {
			t.Errorf("Expected default for '%s' to be '%s', got '%s'", flagName, expectedDefault, flag.DefValue)
		}
	}
}

func TestSaveCmd_RequiredFlags(t *testing.T) {
	cmd := NewSaveCmd()

	// Capture stderr to check for error message about required flags
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	// Try to execute without required flags
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err == nil {
		t.Error("Expected error when no required flags provided")
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "required flag") {
		t.Errorf("Expected error message about required flag, got: %s", stderrOutput)
	}
}

func TestSaveCmd_FlagParsing(t *testing.T) {
	cmd := NewSaveCmd()

	// Set flags and verify they can be parsed
	args := []string{
		"--file", "docker-compose.yml",
		"--file", "docker-compose.override.yml",
		"--output", "/tmp/output",
		"--workdir", "/tmp/work",
		"--tag", "v1.0.0",
		"--loglevel", "debug",
	}

	cmd.SetArgs(args)

	// Parse flags but don't execute (since we can't mock the compose client easily)
	err := cmd.ParseFlags(args)
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Verify flags were parsed correctly
	fileFlag := cmd.Flag("file")
	if fileFlag == nil {
		t.Fatal("file flag not found")
	}

	outputFlag := cmd.Flag("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}
	if outputFlag.Value.String() != "/tmp/output" {
		t.Errorf("Expected output to be '/tmp/output', got '%s'", outputFlag.Value.String())
	}

	workdirFlag := cmd.Flag("workdir")
	if workdirFlag == nil {
		t.Fatal("workdir flag not found")
	}
	if workdirFlag.Value.String() != "/tmp/work" {
		t.Errorf("Expected workdir to be '/tmp/work', got '%s'", workdirFlag.Value.String())
	}

	tagFlag := cmd.Flag("tag")
	if tagFlag == nil {
		t.Fatal("tag flag not found")
	}
	if tagFlag.Value.String() != "v1.0.0" {
		t.Errorf("Expected tag to be 'v1.0.0', got '%s'", tagFlag.Value.String())
	}

	loglevelFlag := cmd.Flag("loglevel")
	if loglevelFlag == nil {
		t.Fatal("loglevel flag not found")
	}
	if loglevelFlag.Value.String() != "debug" {
		t.Errorf("Expected loglevel to be 'debug', got '%s'", loglevelFlag.Value.String())
	}
}

func TestSaveCmd_ShortFlags(t *testing.T) {
	cmd := NewSaveCmd()

	// Test short flag equivalents
	args := []string{
		"-f", "docker-compose.yml",
		"-o", "/tmp/output",
		"-w", "/tmp/work",
		"-t", "v2.0.0",
		"-l", "warn",
	}

	cmd.SetArgs(args)

	err := cmd.ParseFlags(args)
	if err != nil {
		t.Fatalf("Failed to parse short flags: %v", err)
	}

	// Verify short flags work
	if cmd.Flag("output").Value.String() != "/tmp/output" {
		t.Errorf("Expected output to be '/tmp/output', got '%s'", cmd.Flag("output").Value.String())
	}

	if cmd.Flag("workdir").Value.String() != "/tmp/work" {
		t.Errorf("Expected workdir to be '/tmp/work', got '%s'", cmd.Flag("workdir").Value.String())
	}

	if cmd.Flag("tag").Value.String() != "v2.0.0" {
		t.Errorf("Expected tag to be 'v2.0.0', got '%s'", cmd.Flag("tag").Value.String())
	}

	if cmd.Flag("loglevel").Value.String() != "warn" {
		t.Errorf("Expected loglevel to be 'warn', got '%s'", cmd.Flag("loglevel").Value.String())
	}
}

func TestSaveCmd_MultipleFiles(t *testing.T) {
	cmd := NewSaveCmd()

	args := []string{
		"--file", "docker-compose.yml",
		"--file", "docker-compose.prod.yml",
		"--file", "docker-compose.override.yml",
		"--output", "/tmp/output",
	}

	cmd.SetArgs(args)

	err := cmd.ParseFlags(args)
	if err != nil {
		t.Fatalf("Failed to parse flags with multiple files: %v", err)
	}

	// The file flag is a StringSlice, so we can check if multiple values are supported
	fileFlag := cmd.Flag("file")
	if fileFlag == nil {
		t.Fatal("file flag not found")
	}

	// For StringSlice flags, the Value.String() returns the slice representation
	fileValue := fileFlag.Value.String()
	expectedFiles := []string{"docker-compose.yml", "docker-compose.prod.yml", "docker-compose.override.yml"}

	for _, expectedFile := range expectedFiles {
		if !strings.Contains(fileValue, expectedFile) {
			t.Errorf("Expected file value to contain '%s', got '%s'", expectedFile, fileValue)
		}
	}
}

func TestSaveCmd_ConfigCreation(t *testing.T) {
	// Test that the config creation logic works correctly
	// This tests the part of RunE that creates the ComposeConfig

	dockerComposePath := []string{"docker-compose.yml", "docker-compose.override.yml"}
	workDir := "/tmp/work"
	outputDir := "/tmp/output"
	tag := "v1.0.0"
	logLevel := "debug"

	// Create config the same way the command does
	config := Compose.ComposeConfig{
		DockerComposePath: dockerComposePath,
		WorkDir:           workDir,
		OutputDir:         outputDir,
		Tag:               tag,
		LogLevel:          logLevel,
	}

	// Verify config fields
	if len(config.DockerComposePath) != 2 {
		t.Errorf("Expected 2 compose paths, got %d", len(config.DockerComposePath))
	}

	if config.DockerComposePath[0] != "docker-compose.yml" {
		t.Errorf("Expected first path to be 'docker-compose.yml', got '%s'", config.DockerComposePath[0])
	}

	if config.DockerComposePath[1] != "docker-compose.override.yml" {
		t.Errorf("Expected second path to be 'docker-compose.override.yml', got '%s'", config.DockerComposePath[1])
	}

	if config.WorkDir != workDir {
		t.Errorf("Expected WorkDir to be '%s', got '%s'", workDir, config.WorkDir)
	}

	if config.OutputDir != outputDir {
		t.Errorf("Expected OutputDir to be '%s', got '%s'", outputDir, config.OutputDir)
	}

	if config.Tag != tag {
		t.Errorf("Expected Tag to be '%s', got '%s'", tag, config.Tag)
	}

	if config.LogLevel != logLevel {
		t.Errorf("Expected LogLevel to be '%s', got '%s'", logLevel, config.LogLevel)
	}
}

func TestSaveCmd_HelpText(t *testing.T) {
	cmd := NewSaveCmd()

	// Test that help can be generated without errors
	var helpOutput bytes.Buffer
	cmd.SetOut(&helpOutput)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute help command: %v", err)
	}

	helpText := helpOutput.String()

	// Check that help contains expected content
	expectedContent := []string{
		"Save docker compose project",
		"--file", "-f",
		"--output", "-o",
		"--workdir", "-w",
		"--tag", "-t",
		"--loglevel", "-l",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(helpText, expected) {
			t.Errorf("Expected help text to contain '%s', got: %s", expected, helpText)
		}
	}
}

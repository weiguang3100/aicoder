package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillsManagement(t *testing.T) {
	// Setup temp home
	tmpHome, err := os.MkdirTemp("", "skills-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	os.Setenv("HOME", tmpHome)
	if os.Getenv("USERPROFILE") != "" {
		defer os.Setenv("USERPROFILE", os.Getenv("USERPROFILE"))
		os.Setenv("USERPROFILE", tmpHome)
	}

	app := &App{testHomeDir: tmpHome}

	// 1. Test ListSkills (Empty)
	skills := app.ListSkills()
	if len(skills) != 0 {
		t.Errorf("Expected 0 skills, got %d", len(skills))
	}

	// 2. Test AddSkill (Address)
	err = app.AddSkill("TestSkill1", "Description 1", "address", "@test/skill")
	if err != nil {
		t.Errorf("AddSkill failed: %v", err)
	}

	skills = app.ListSkills()
	if len(skills) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "TestSkill1" || skills[0].Value != "@test/skill" {
		t.Errorf("Skill data mismatch: %+v", skills[0])
	}

	// 3. Test AddSkill (Zip) - requires a dummy zip file
	zipPath := filepath.Join(tmpHome, "test.zip")
	os.WriteFile(zipPath, []byte("dummy zip content"), 0644)

	err = app.AddSkill("TestSkill2", "Description 2", "zip", zipPath)
	if err != nil {
		t.Errorf("AddSkill (zip) failed: %v", err)
	}

	skills = app.ListSkills()
	if len(skills) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(skills))
	}

	// Verify zip was copied
	skillsDir := app.GetSkillsDir()
	copiedZip := filepath.Join(skillsDir, "test.zip")
	if _, err := os.Stat(copiedZip); os.IsNotExist(err) {
		t.Errorf("Zip file was not copied to %s", copiedZip)
	}

	// 4. Test DeleteSkill
	err = app.DeleteSkill("TestSkill2")
	if err != nil {
		t.Errorf("DeleteSkill failed: %v", err)
	}

	skills = app.ListSkills()
	if len(skills) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "TestSkill1" {
		t.Errorf("Wrong skill remained: %s", skills[0].Name)
	}

	// Verify zip was deleted
	if _, err := os.Stat(copiedZip); !os.IsNotExist(err) {
		t.Errorf("Zip file was not deleted")
	}
}

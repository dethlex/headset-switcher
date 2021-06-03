package icons

import (
	_ "embed"
	"fmt"
	"os"
	"path"
)

var (
	//go:embed disabled.png
	disabled     []byte
	IconDisabled string

	//go:embed listen.png
	listen     []byte
	IconListen string

	//go:embed speak.png
	speak     []byte
	IconSpeak string

	// directory for icons
	iconDir string
)

func saveFile(name string, icon []byte) error {
	file, err := os.Create(name)
	defer file.Close()

	if err != nil {
		return fmt.Errorf("Can't create file: %w", err)
	}

	if _, err = file.Write(icon); err != nil {
		return fmt.Errorf("Can't write file: %w", err)
	}
	return nil
}

func CreateIcons() error {

	// Get user cache dir
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("Can't locate user cache folder: %w", err)
	}
	iconDir = path.Join(cacheDir, "hp-switch")
	if err := os.MkdirAll(iconDir, 0766); err != nil {
		return fmt.Errorf("Can't create icon folder path: %w", err)
	}

	// put all binary data to files
	IconListen = path.Join(iconDir, "listen.png")
	if err := saveFile(IconListen, listen); err != nil {
		return fmt.Errorf("Can't write icon file: %w", err)
	}

	IconSpeak = path.Join(iconDir, "speak.png")
	if err := saveFile(IconSpeak, speak); err != nil {
		return fmt.Errorf("Can't write icon file: %w", err)
	}

	IconDisabled = path.Join(iconDir, "disabled.png")
	if err := saveFile(IconDisabled, disabled); err != nil {
		return fmt.Errorf("Can't write icon file: %w", err)
	}

	return nil
}

func DeleteIcons() error {
	if err := os.RemoveAll(iconDir); err != nil {
		return fmt.Errorf("Can't remove icon folder: %w", err)
	}
	return nil
}

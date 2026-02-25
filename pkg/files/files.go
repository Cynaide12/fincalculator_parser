package lib_files

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

func ValidateFileExtensions(allowedExtensins []string, fileExt string) bool {
	for _, ext := range allowedExtensins {
		if ext == fileExt {
			return true
		}
	}

	return false
}

func CreateDigitalPath(baseDir string, userSelectedPath string, docName string) (string, error) {
	safePath := strings.ReplaceAll(userSelectedPath, "/", string(os.PathSeparator))
	safeDocName := SanitizeFileName(docName)

	fullPath := filepath.Join(baseDir, safePath, safeDocName)

	//  Создаём папки (если их нет)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", fmt.Errorf("ошибка создания папок: %w", err)
	}

	return fullPath, nil
}

func SanitizeFileName(name string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			return r
		}
		return '_'
	}, name)
}

func NormalizePath(path string) string {
	return strings.ReplaceAll(filepath.Clean(path), "\\", "/")
}

func SaveScanFile(src *multipart.FileHeader, dstPath string, fileName string) error {
	// Открываем исходный файл
	srcFile, err := src.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Создаём файл в папке назначени
	dstFile, err := os.Create(filepath.Join(dstPath, fileName))
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Копируем содержимое
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}

func DeleteFiles(filePaths []string) error {
	for _, file := range filePaths {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	return nil
}

func MoveFile(oldPath, newPath string) error {
	// Создаем директорию для нового пути
	if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		return err
	}

	// Переносим файл
	if err := os.Rename(oldPath, newPath); err != nil {
		// Если перенос между разделами файловой системы, используем копирование
		if err := CopyFile(oldPath, newPath); err != nil {
			return err
		}
		return os.Remove(oldPath)
	}

	return nil
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func UniqueFileName(fileName string) string {
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)
	id := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d%s", name, id, ext)
}

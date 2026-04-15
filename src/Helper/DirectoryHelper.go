package Helper

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func GetFolderNames(directory string) ([]string, error) {
	folderNames := []string{}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != directory && filepath.Dir(path) == directory {
			folderNames = append(folderNames, info.Name())
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return folderNames, nil
}
func GetFilenamesFromDirectory(directory string) ([]string, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}
func CreateFolderIfNotExists(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check folder existence: %w", err)
	}

	return nil
}
func MergeFiles(file1Path, file2Path, mergedFilePath string) error {
	file1Content, err := ioutil.ReadFile(file1Path)
	if err != nil {
		return fmt.Errorf("unable to read file 1: %v", err)
	}
	file2Content, err := ioutil.ReadFile(file2Path)
	if err != nil {
		return fmt.Errorf("unable to read file 2: %v", err)
	}

	mergedContent := []byte(strings.TrimSpace(string(file1Content)) + "\n" + strings.TrimSpace(string(file2Content)) + "\n")

	err = os.Remove(file1Path)
	if err != nil {
		return fmt.Errorf("unable to delete file 1: %v", err)
	}

	err = os.Remove(file2Path)
	if err != nil {
		return fmt.Errorf("unable to delete file 2: %v", err)
	}

	mergedDir := filepath.Dir(mergedFilePath)
	err = os.MkdirAll(mergedDir, 0755)
	if err != nil {
		return fmt.Errorf("unable to create merged file directory: %v", err)
	}

	err = ioutil.WriteFile(mergedFilePath, mergedContent, 0644)
	if err != nil {
		return fmt.Errorf("unable to write merged file: %v", err)
	}
	return nil
}

func ExtractZip(zipPath, destPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(destPath, file.Name)

		if file.FileInfo().IsDir() {
			err := os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
			continue
		}

		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		writer, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer writer.Close()

		reader, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip file entry: %v", err)
		}
		defer reader.Close()

		_, err = io.Copy(writer, reader)
		if err != nil {
			return fmt.Errorf("failed to extract file: %v", err)
		}
	}

	return nil
}

func CreateZip(sourceFile, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	sourceFileInfo, err := os.Stat(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	zipEntry, err := zipWriter.Create(sourceFileInfo.Name())
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %v", err)
	}

	source, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer source.Close()

	_, err = io.Copy(zipEntry, source)
	if err != nil {
		return fmt.Errorf("failed to add file to zip: %v", err)
	}

	return nil
}

func GetNameFromFileName(fileName string) string {
	name := filepath.Base(fileName)
	extension := filepath.Ext(name)
	nameWithoutExtension := name[:len(name)-len(extension)]
	return nameWithoutExtension
}

package Api

import (
	"bytes"
	"context"
	"fmt"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io"
	"log"
	"mime"
	"os"
	"regexp"
	"data-fetcher-api/src/Helper"
	"github.com/getsentry/sentry-go"
)

const DRIVE_LINK = "https://drive.google.com/drive/folders/"
const FILE_LINK = "https://drive.google.com/file/d/"

func CreateClient() (*drive.Service, error) {
	ctx := context.Background()
	config, _ := Helper.GetParameter()
	filePath := config.Parameters.ConfigPath
	client, err := drive.NewService(ctx, option.WithCredentialsFile(filePath), option.WithScopes(drive.DriveScope))
	if err != nil {
		sentry.CaptureException(err)
	}

	return client, nil
}

func CreateFolder(client *drive.Service, parentID string, folderName string) (string, error) {
	folder := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}

	createdFolder, err := client.Files.Create(folder).Do()
	if err != nil {
		return "", fmt.Errorf("unable to create folder: %v", err)
	}
	return DRIVE_LINK + createdFolder.Id, nil
}

func CheckFolderExists(client *drive.Service, parentID string, folderName string) (string, bool, error) {
	query := fmt.Sprintf("name = '%s' and '%s' in parents and mimeType = 'application/vnd.google-apps.folder' and trashed = false", folderName, parentID)
	response, err := client.Files.List().Q(query).Do()
	if err != nil {
		return "", false, fmt.Errorf("unable to check folder existence: %v", err)
	}

	if len(response.Files) > 0 {
		folder := response.Files[0]
		folderWebLink := DRIVE_LINK + folder.Id
		return folderWebLink, true, nil
	}

	return "", false, nil
}

func CheckFileExists(client *drive.Service, parentID string, fileName string) (string, bool, error) {
	query := fmt.Sprintf("name = '%s' and '%s' in parents and mimeType != 'application/vnd.google-apps.folder' and trashed = false", fileName, parentID)
	response, err := client.Files.List().Q(query).Do()
	if err != nil {
		return "", false, fmt.Errorf("unable to check file existence: %v", err)
	}

	if len(response.Files) > 0 {
		file := response.Files[0]
		fileWebLink := FILE_LINK + file.Id
		return fileWebLink, true, nil
	}

	return "", false, nil
}
func StoreDataInFile(client *drive.Service, parentID string, fileName string, data []byte) error {
	file := &drive.File{
		Name:    fileName,
		Parents: []string{parentID},
	}

	createdFile, err := client.Files.Create(file).Media(bytes.NewReader(data)).Do()
	if err != nil {
		return fmt.Errorf("unable to store data in file: %v", err)
	}

	log.Printf("Data stored in file '%s' successfully. File ID: %s\n", fileName, createdFile.Id)
	return nil
}

func DownloadFile(client *drive.Service, fileID string, destinationPath string) error {
	resp, err := client.Files.Get(fileID).Download()
	if err != nil {
		fmt.Println("write failed")
		fmt.Println(err)
		return fmt.Errorf("unable to download file: %v", err)
	}
	defer resp.Body.Close()

	file, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("unable to create file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("unable to write file: %v", err)
	}
	return nil
}

func DeleteFile(client *drive.Service, fileID string) error {
	err := client.Files.Delete(fileID).Do()
	if err != nil {
		return fmt.Errorf("unable to delete file: %v", err)
	}
	return nil
}

func UploadFile(client *drive.Service, parentID string, filePath string, fileName string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %v", err)
	}
	defer file.Close()

	mimeType := mime.TypeByExtension(fileName)

	fileMetadata := &drive.File{
		Name:     fileName,
		Parents:  []string{parentID},
		MimeType: mimeType,
	}

	createdFile, err := client.Files.Create(fileMetadata).Media(file).Do()
	if err != nil {
		return "", fmt.Errorf("unable to upload file: %v", err)
	}

	fileURL := FILE_LINK + createdFile.Id
	log.Printf("File '%s' uploaded successfully. File ID: %s\n", fileName, createdFile.Id)
	return fileURL, nil
}

func GetFolderIDFromURL(url string) (string, error) {
	re := regexp.MustCompile(`[a-zA-Z0-9_-]{25,}`)
	matches := re.FindStringSubmatch(url)
	if len(matches) == 0 {
		return "", fmt.Errorf("unable to extract folder ID from URL")
	}
	return matches[0], nil
}

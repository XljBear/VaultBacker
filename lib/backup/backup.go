package backup

import (
	"VaultBacker/lib/file"
	"VaultBacker/lib/mail"
	"VaultBacker/lib/zip"
	"VaultBacker/models"
	"context"
	"database/sql"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
	"os"
	"time"
)

var backupKeyStore = ""

func Backup(config models.Config) (backupKey string, err error) {
	var todayBackupPath string
	defer func() {
		if err != nil {
			_ = cleanUpBackupFolder(todayBackupPath)
		}
	}()
	// Create a new backup directory
	fmt.Print("Preparing backup folder...")
	todayBackupPath, err = prepareBackupFolder()
	if err != nil {
		return
	}
	backupKey = backupKeyStore
	fmt.Println("Done")

	// Copy the vault files to the backup directory
	fmt.Print("Copying vault warden data...")
	err = copyVaultWardenData(config, todayBackupPath)
	if err != nil {
		return
	}
	fmt.Println("Done")

	// Remove sqlite database in the backup directory
	fmt.Print("Removing sql temp file...")
	err = removeSqlTempFiles(todayBackupPath)
	if err != nil {
		return
	}
	fmt.Println("Done")

	// Process sqlite database backup
	fmt.Print("Doing sqlite backup...")
	err = doSqlBackup(config, todayBackupPath)
	if err != nil {
		return
	}
	fmt.Println("Done")

	// Zip the backup directory
	fmt.Print("Zipping backup data...")
	err = zipBackupData(todayBackupPath)
	if err != nil {
		return
	}
	fmt.Println("Done")

	// Check is need zip for the user
	if config.BackupConfig.ForUser.Enabled {
		// Zip the backup directory for the user
		fmt.Print("Zipping backup data for user...")
		err = zipBackupDataForUser(config, todayBackupPath)
		if err != nil {
			return
		}
		fmt.Println("Done")
	}

	// Cleanup the backup directory
	fmt.Print("Cleaning up backup folder...")
	err = cleanUpBackupFolder(todayBackupPath)
	fmt.Println("Done")
	return
}
func getPwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}
func prepareBackupFolder() (todayBackupPath string, err error) {
	backupPath := getPwd() + "/backup/"
	backupKeyStore = time.Now().Format("20060102")
	todayBackupPath = backupPath + backupKeyStore
	//check backup folder is exists then delete it
	if _, fErr := os.Stat(todayBackupPath); fErr == nil {
		_ = os.RemoveAll(todayBackupPath)
	}
	//create backup folder
	err = os.MkdirAll(todayBackupPath, 0777)
	return
}
func copyVaultWardenData(config models.Config, todayBackupPath string) (err error) {
	err = file.Copy(config.BackupConfig.VaultWardenDataPath+"/", todayBackupPath)
	return
}
func removeSqlTempFiles(todayBackupPath string) (err error) {
	needRemoveFile := []string{"db.sqlite3", "db.sqlite3-shm", "db.sqlite3-wal"}
	for _, file := range needRemoveFile {
		err = os.Remove(todayBackupPath + "/" + file)
		if err != nil {
			return
		}
	}
	return
}
func doSqlBackup(config models.Config, todayBackupPath string) (err error) {
	originDbPath := config.BackupConfig.VaultWardenDataPath + "/db.sqlite3"
	backupDbPath := todayBackupPath + "/db.sqlite3"
	db, err := sql.Open("sqlite3", originDbPath)
	if err != nil {
		return
	}
	backupSql := "VACUUM INTO '" + backupDbPath + "';"
	_, err = db.Exec(backupSql)
	return
}
func zipBackupData(todayBackupPath string) (err error) {
	err = zip.Zip(todayBackupPath, todayBackupPath+".zip")
	return
}
func zipBackupDataForUser(config models.Config, todayBackupPath string) (err error) {
	backupNotForUserFiles := config.BackupConfig.ForUser.NotNeedFiles
	for _, file := range backupNotForUserFiles {
		err = os.Remove(todayBackupPath + "/" + file)
		if err != nil {
			return
		}
	}
	backupNotForUseFolders := config.BackupConfig.ForUser.NotNeedFolders
	for _, folder := range backupNotForUseFolders {
		err = os.RemoveAll(todayBackupPath + "/" + folder)
		if err != nil {
			return
		}
	}
	err = zip.Zip(todayBackupPath, todayBackupPath+"_user.zip")
	return
}
func cleanUpBackupFolder(todayBackupPath string) (err error) {
	err = os.RemoveAll(todayBackupPath)
	return
}

func Upload(config models.Config, backupKey string) (err error) {
	backupFile := getPwd() + "/backup/" + backupKey + ".zip"
	BucketURL := config.BackupConfig.Cos.BucketURL
	SecretID := config.BackupConfig.Cos.SecretID
	SecretKey := config.BackupConfig.Cos.SecretKey
	StorePath := config.BackupConfig.Cos.StorePath
	bucketUrl, _ := url.Parse(BucketURL)
	bucket := &cos.BaseURL{BucketURL: bucketUrl}
	cosClient := cos.NewClient(bucket, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  SecretID,
			SecretKey: SecretKey,
		},
	})
	_, err = cosClient.Object.PutFromFile(context.Background(), StorePath+"/"+backupKey+".zip", backupFile, nil)
	return
}

func SendUserMail(config models.Config, backupKey string) (err error) {
	if !config.BackupConfig.ForUser.Enabled {
		return
	}

	todayTimeString := time.Now().Format("2006-01-02")
	backupFilePath := getPwd() + "/backup/" + backupKey + "_user.zip"
	emailSubject := "尊敬的VoxelMatrix高级会员，这是来自Vault的备份文件。"
	emailBody := "您好，<br><br>全量Vault数据库备份文件已经生成。<br><br>"
	emailBody += "备份日期：" + todayTimeString + "</a><br><br>"
	emailBody += "Vault管理员邮箱：<a href='mailto:vault@voxelmatrix.com'>vault@voxelmatrix.com</a><br><br>"
	emailBody += "请注意！附件中的内容涉及高度机密，其中包含您在Vault保存的所有账户及密码（非明文存储，无法解密），仅作为紧急恢复时使用，请妥善保存。"
	if _, fErr := os.Stat(backupFilePath); fErr != nil {
		return
	}

	userEmailList := config.BackupConfig.ForUser.UserEmailList

	mailClient := &mail.SendMail{
		Host:     config.BackupConfig.Smtp.Host,
		Port:     config.BackupConfig.Smtp.Port,
		User:     config.BackupConfig.Smtp.User,
		Password: config.BackupConfig.Smtp.Password,
	}

	for _, email := range userEmailList {
		fmt.Print("Sending backup file to " + email + "...")
		message := mail.Message{
			From:        config.BackupConfig.Smtp.User,
			FromName:    config.BackupConfig.Smtp.SenderName,
			To:          []string{email},
			Cc:          []string{},
			Bcc:         []string{},
			Subject:     emailSubject,
			Body:        emailBody,
			ContentType: "text/html;charset=utf-8",
			Attachment: mail.Attachment{
				FilePath:    backupFilePath,
				Name:        "Top-Secret.zip",
				ContentType: "application/zip",
				WithFile:    true,
			},
		}
		err = mailClient.DoSend(message)
		if err != nil {
			return
		}
		fmt.Println("Done")
	}
	_ = os.Remove(backupFilePath)
	return
}

package main

import (
	"VaultBacker/lib/backup"
	"VaultBacker/lib/web"
	"VaultBacker/models"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
	"time"
)

func main() {

	programPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}

	configData, err := ioutil.ReadFile(programPath + "/config.json")
	if err != nil {
		fmt.Println("Error reading config file")
		return
	}

	config := models.Config{}

	err = json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Println("Error parsing config file")
		return
	}

	if config.WebServer.Enabled {
		fmt.Printf("Starting web server on port %d...\n", config.WebServer.Port)
		go func() {
			web.StartWebServer(config)
		}()
	}

	record := models.Record{}
	//check record.json file exists
	if _, err = os.Stat(programPath + "/record.json"); os.IsNotExist(err) {
		fmt.Println("record.json file not found, creating...")
		saveRecord(record)
	}

	recordData, err := ioutil.ReadFile(programPath + "/record.json")
	if err != nil {
		fmt.Println("Error reading record file")
		return
	}

	err = json.Unmarshal(recordData, &record)
	if err != nil {
		fmt.Println("Error parsing record file")
		return
	}

	for {
		now := time.Now()
		if !config.BackupConfig.Enabled {
			time.Sleep(time.Second * 5)
			continue
		}
		if record.CosRecord.LastBackupTime == nil ||
			int(now.Sub(*record.CosRecord.LastBackupTime).Hours()/24) >= config.BackupConfig.BackupInterval {
			backupKey, err := doBackup(config)
			if err != nil {
				fmt.Println(err)
				return
			}
			record.CosRecord.LastBackupTime = &now
			backupFile := models.BackupFiles{}
			backupFile.BackupTime = &now
			backupFile.Key = backupKey
			record.BackupFiles = append(record.BackupFiles, backupFile)

			saveRecord(record)

			if config.BackupConfig.ForUser.Enabled {
				if record.UserRecord.LastSendTime == nil ||
					int(now.Sub(*record.UserRecord.LastSendTime).Hours()/24) >= config.BackupConfig.ForUser.SendInterval {
					err = sendUserEmail(config, backupKey)
					if err != nil {
						fmt.Println(err)
						return
					}
					record.UserRecord.LastSendTime = &now
					saveRecord(record)
				} else {
					backupFilePath := programPath + "/backup/" + backupKey + "_user.zip"
					_ = os.Remove(backupFilePath)
				}
			}
		}
		dirtyFlag := false
		for i, backupFile := range record.BackupFiles {
			if int(now.Sub(*backupFile.BackupTime).Hours()/24) > config.BackupConfig.BackupRetention {
				backupFilePath := programPath + "/backup/" + backupFile.Key + ".zip"
				fmt.Println("Delete old backup file " + backupFile.Key + ".zip")
				_ = os.Remove(backupFilePath)
				record.BackupFiles = append(record.BackupFiles[:i], record.BackupFiles[i+1:]...)
				dirtyFlag = true
			}
		}
		if dirtyFlag {
			saveRecord(record)
		}
		time.Sleep(time.Second * 5)
	}

}
func saveRecord(record models.Record) {
	programPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	tempData, _ := json.Marshal(record)
	err = ioutil.WriteFile(programPath+"/record.json", tempData, 0644)
	if err != nil {
		fmt.Println("Error writing record.json file")
		return
	}
}
func doBackup(config models.Config) (backupKey string, err error) {

	backupKey, err = backup.Backup(config)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Print("Uploading backup to Cos...")
	err = backup.Upload(config, backupKey)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Done")
	return
}
func sendUserEmail(config models.Config, backupKey string) (err error) {
	err = backup.SendUserMail(config, backupKey)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	return
}

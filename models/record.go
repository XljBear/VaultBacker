package models

import "time"

type Record struct {
	CosRecord struct {
		LastBackupTime *time.Time `json:"LastBackupTime"`
	} `json:"CosRecord"`
	UserRecord struct {
		LastSendTime *time.Time `json:"LastSendTime"`
	} `json:"UserRecord"`
	BackupFiles []BackupFiles `json:"BackupFiles"`
}
type BackupFiles struct {
	Key        string     `json:"Key"`
	BackupTime *time.Time `json:"BackupTime"`
}

package models

type Config struct {
	WebServer struct {
		Enabled bool `json:"Enabled"`
		Port    int  `json:"port"`
	} `json:"WebServer"`
	BackupConfig struct {
		Enabled             bool   `json:"Enabled"`
		VaultWardenDataPath string `json:"VaultWardenDataPath"`
		BackupInterval      int    `json:"BackupInterval"`
		BackupRetention     int    `json:"BackupRetention"`
		ForUser             struct {
			Enabled        bool     `json:"Enabled"`
			NotNeedFiles   []string `json:"NotNeedFiles"`
			NotNeedFolders []string `json:"NotNeedFolders"`
			UserEmailList  []string `json:"UserEmailList"`
			SendInterval   int      `json:"SendInterval"`
		} `json:"ForUser"`
		Cos struct {
			SecretID  string `json:"SecretID"`
			SecretKey string `json:"SecretKey"`
			BucketURL string `json:"BucketUrl"`
			StorePath string `json:"StorePath"`
		} `json:"Cos"`
		Smtp struct {
			Host       string `json:"Host"`
			Port       int    `json:"Port"`
			User       string `json:"User"`
			Password   string `json:"Password"`
			SenderName string `json:"SenderName"`
		} `json:"Smtp"`
	} `json:"BackupConfig"`
}

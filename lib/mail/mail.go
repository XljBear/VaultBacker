package mail

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type Mail interface {
	DoAuth()
	DoSend(message Message) error
}

type SendMail struct {
	User     string
	Password string
	Host     string
	Port     int
	Auth     smtp.Auth
}

type Attachment struct {
	FilePath    string
	Name        string
	ContentType string
	WithFile    bool
}

type Message struct {
	From        string
	FromName    string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	ContentType string
	Attachment  Attachment
}

func (mail *SendMail) DoAuth() {
	mail.Auth = smtp.PlainAuth("", mail.User, mail.Password, mail.Host)
}

func (mail SendMail) DoSend(message Message) error {
	mail.DoAuth()
	buffer := bytes.NewBuffer(nil)
	boundary := "VoxelMatrix Technologies."
	Header := make(map[string]string)
	Header["From"] = message.FromName + "<" + message.From + ">"
	Header["To"] = strings.Join(message.To, ";")
	Header["Cc"] = strings.Join(message.Cc, ";")
	Header["Bcc"] = strings.Join(message.Bcc, ";")
	Header["Subject"] = message.Subject
	Header["Content-Type"] = "multipart/mixed;boundary=" + boundary
	Header["Mime-Version"] = "1.0"
	Header["Date"] = time.Now().String()
	mail.writeHeader(buffer, Header)

	body := "\r\n--" + boundary + "\r\n"
	body += "Content-Type:" + message.ContentType + "\r\n"
	body += "\r\n" + message.Body + "\r\n"
	buffer.WriteString(body)

	if message.Attachment.WithFile {
		Attachment := "\r\n--" + boundary + "\r\n"
		Attachment += "Content-Transfer-Encoding:base64\r\n"
		Attachment += "Content-Disposition:Attachment\r\n"
		Attachment += "Content-Type:" + message.Attachment.ContentType + ";name=\"" + message.Attachment.Name + "\"\r\n"
		buffer.WriteString(Attachment)
		defer func() {
			if err := recover(); err != nil {
				log.Fatalln(err)
			}
		}()
		mail.writeFile(buffer, message.Attachment.FilePath)
	}

	buffer.WriteString("\r\n--" + boundary + "--")
	err := smtp.SendMail(mail.Host+":"+strconv.Itoa(mail.Port), mail.Auth, message.From, message.To, buffer.Bytes())
	return err
}

func (mail SendMail) writeHeader(buffer *bytes.Buffer, Header map[string]string) string {
	header := ""
	for key, value := range Header {
		header += key + ":" + value + "\r\n"
	}
	header += "\r\n"
	buffer.WriteString(header)
	return header
}

// read and write the file to buffer
func (mail SendMail) writeFile(buffer *bytes.Buffer, fileName string) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err.Error())
	}
	payload := make([]byte, base64.StdEncoding.EncodedLen(len(file)))
	base64.StdEncoding.Encode(payload, file)
	buffer.WriteString("\r\n")
	for index, line := 0, len(payload); index < line; index++ {
		buffer.WriteByte(payload[index])
		if (index+1)%76 == 0 {
			buffer.WriteString("\r\n")
		}
	}
}

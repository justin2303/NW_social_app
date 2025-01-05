package db_funcs

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func SftpFileDownload() {
	// SFTP server details
	sftpServer := "199.127.62.163:8822" // srvr
	sftpUser := "FreddyFazbearXXX"     // user
	sftpPass := "asssniper"            // pass

	// Get Current Date
	dateString := GetLastSat()
	remoteFile := fmt.Sprintf("/199.127.62.163_7240/Logs/server_log_%s.txt", dateString)

	fmt.Println("Using current date:", dateString)
	fmt.Println("Remote file path:", remoteFile)

	// Change local file object with current date
	localFile := fmt.Sprintf("./logs/server_log_%s.txt", dateString)
	fmt.Println("Local file will be saved as:", localFile)

	// Create logs directory if it doesn't exist
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		fmt.Println("Failed to create logs directory:", err)
		return
	}

	// Connection phase
	// Create SSH config
	config := &ssh.ClientConfig{
		User: sftpUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sftpPass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // WARNING: This is probably unsafe lol
	}

	// Connect to the SFTP site
	conn, err := ssh.Dial("tcp", sftpServer, config)
	if err != nil {
		fmt.Println("Failure when connecting to SFTP site:", err)
		return
	}
	defer conn.Close()

	// Create a new SFTP client
	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		fmt.Println("Failed when trying to create SFTP client:", err)
		return
	}
	defer sftpClient.Close()

	// Open the remote file
	remoteFileHandle, err := sftpClient.Open(remoteFile)
	if err != nil {
		fmt.Println("Failed when trying to open the remote file:", err)
		return
	}
	defer remoteFileHandle.Close()
	fmt.Println("Remote file SUCCESS!!!")

	// Copy over file phase

	// Local file creation
	outFile, err := os.Create(localFile)
	if err != nil {
		fmt.Println("Failure whem trying to create local file: ", err)
		return
	}
	defer outFile.Close()
	fmt.Println("Local file successfully created")

	// Copy the contents of the remote file to the local file
	_, err = io.Copy(outFile, remoteFileHandle)
	if err != nil {
		fmt.Println("Failure when trying to copy the contents of the file:", err)
		return
	}

	fmt.Println("TOTAL SUCCESS!!!!! 61E WINS AGAIN!!!:", localFile)
}

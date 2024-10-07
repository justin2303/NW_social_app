
package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func SFTP_Automation_Script() {
	// SFTP server details use this is where u can ur details wen trying to conect to the sftp site
	sftpServer := "199.127.62.40:8822"                               // srvr
	sftpUser := "FreddyFazbearXXX"                                   // user
	sftpPass := "asssniper"                                          // password
	remoteFile := "/199.127.62.40_7240/Logs/server_log_10_05_24.txt" // remote file path - dw u dont gotta change the pathway it shud still pull current date

	// this is supposed to pull the current date
	currentDate := time.Now()
	fmt.Println("Using current date:", currentDate.Format("01-02-06"))

/*
	*******NOT USED, INCASE YOU WOULD LIKE TO CHANGE AND USE INPUT DATE INSTEAD***********
	// Ask for date input
	//var inputDate string
	//fmt.Print("Enter the date (MM-DD-YY): ")
	//fmt.Scan(&inputDate)

	// Parse the date
	//date, err := time.Parse("01-02-06", inputDate)
	//if err != nil {
	//	fmt.Println("Error parsing date:", err)
	//	return
	//}
	******************************************************************************************
*/


	// format the local file name based on the current date 
	localFile := fmt.Sprintf("server_log_%02d_%02d_%02d.txt", currentDate.Month(), currentDate.Day(), currentDate.Year()%100)
	fmt.Println("Local file will be saved as:", localFile)

/// connection phase ///

	// create SSH config
	config := &ssh.ClientConfig{
		User: sftpUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sftpPass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // this is probably super unsafe but i cudnt figure out how to do shit with the hsot key lol
	}

	// connect to the SFTP site 
	conn, err := ssh.Dial("tcp", sftpServer, config)
	if err != nil {
		fmt.Println("failure when connecting to sftp site: ", err)
		return
	}
	defer conn.Close()



	// create a new SFTP client
	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		fmt.Println("failed wen trying to create sftp site: ", err)
		return
	}
	defer sftpClient.Close()

	// open the remote file
	remoteFileHandle, err := sftpClient.Open(remoteFile)
	if err != nil {
		fmt.Println("failed when trying to connect to the pathway: ", err)
		return
	}
	defer remoteFileHandle.Close()
	fmt.Println("remote file SUCCESS!!!: ")

///copy over file phase //// 


	// create local file
	outFile, err := os.Create(localFile)
	if err != nil {
		fmt.Println("failure when trying to create local file:", err)
		return
	}
	defer outFile.Close()
	fmt.Println("Local file successfully created")

	// copy the contents of the remote file to the local file
	_, err = io.Copy(outFile, remoteFileHandle)
	if err != nil {
		fmt.Println("Failure when trying to copy the contents of the file:", err)
		return
	}

	fmt.Println("TOTAL SUCCESS!!!!! 61E WINS AGAIN!!!:", localFile)
}

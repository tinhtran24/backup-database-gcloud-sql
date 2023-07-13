package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var projectID = flag.String("projectID", "", "projectID")
var instanceID = flag.String("instanceID", "", "instanceID")
var databaseName = flag.String("databaseName", "", "databaseName")
var backupBucketName = flag.String("backupBucketName", "", "backupBucketName")

// make main func to run the program
func main() {
	//parse the flags
	flag.Parse()
	//call the function
	err := BackupAndMaskData()
	//handle the error
	if err != nil {
		log.Fatal(err)
	}
}

func BackupAndMaskData() error {
	// gcloud set projectID
	projectCmd := fmt.Sprintf("gcloud config set project %s", *projectID)
	if err := runCommand(projectCmd); err != nil {
		return err
	}
	// get file mask.mysql convert to string
	file, err := ioutil.ReadFile("mask.mysql")
	if err != nil {
		return err
	}
	mask := string(file)

	// Generate a unique backup name
	backupName := fmt.Sprintf("%s_backup_%s", *databaseName, uuid.New().String())
	// Create a backup
	backupCmd := fmt.Sprintf("gcloud sql backups create %s --instance=%s --database=%s --async --quiet", backupName, *instanceID, *databaseName)
	if err := runCommand(backupCmd); err != nil {
		return err
	}
	// Export the backup to a Cloud Storage bucket
	exportPath := fmt.Sprintf("gs://%s/%s", *backupBucketName, backupName)
	exportCmd := fmt.Sprintf("gcloud sql export sql %s %s --instance=%s --database=%s --quiet", backupName, exportPath, *instanceID, *databaseName)
	if err := runCommand(exportCmd); err != nil {
		return err
	}
	// Mask sensitive data
	maskCmd := fmt.Sprintf("gcloud sql import sql %s %s --instance=%s --database=%s --quiet --query=\"%s\"", backupName, exportPath, *instanceID, *databaseName, mask)
	if err := runCommand(maskCmd); err != nil {
		return err
	}
	return nil
}

func runCommand(cmd string) error {
	log.Println("Running command:", cmd)
	// Execute the command
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}
	return nil
}

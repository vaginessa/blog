package main

import (
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"log"
	_ "path"
	_ "path/filepath"
	_ "sort"
	"strings"
	"time"
)

var backupFreq = 12 * time.Hour
var bucketDelim = "/"

type BackupConfig struct {
	AwsAccess string
	AwsSecret string
	Bucket    string
	S3Dir     string
	LocalDir  string
}

// removes "/" if exists and adds delim if missing
func sanitizeDirForList(dir, delim string) string {
	if strings.HasPrefix(dir, "/") {
		dir = dir[1:]
	}
	if !strings.HasSuffix(dir, delim) {
		dir = dir + delim
	}
	return dir
}

func listBackupFiles(config *BackupConfig, max int) (*s3.ListResp, error) {
	auth := aws.Auth{config.AwsAccess, config.AwsSecret}
	b := s3.New(auth, aws.USEast).Bucket(config.Bucket)
	dir := sanitizeDirForList(config.S3Dir, bucketDelim)
	return b.List(dir, bucketDelim, "", max)
}

// tests if s3 credentials are valid and aborts if aren't
func ensureValidConfig(config *BackupConfig) {
	if !PathExists(config.LocalDir) {
		log.Fatalf("Invalid s3 backup: directory to backup '%s' doesn't exist\n", config.LocalDir)
	}

	if !strings.HasSuffix(config.S3Dir, bucketDelim) {
		config.S3Dir += bucketDelim
	}
	_, err := listBackupFiles(config, 10)
	if err != nil {
		log.Fatalf("Invalid s3 backup: bucket.List failed %s\n", err.Error())
	}
}

func BackupLoop(config *BackupConfig) {
	ensureValidConfig(config)
	for {
		//doBackup(config)
		time.Sleep(backupFreq)
	}
}

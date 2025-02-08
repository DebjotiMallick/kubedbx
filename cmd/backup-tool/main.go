package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"database-backup/internal/backup"     // Import your backup package
	"database-backup/internal/config"     // Import your config package
	"database-backup/internal/kubernetes" // Import your kubernetes package
	// Import your notification package
	// Import your storage package
	// ... other imports if needed
)

func main() {
	// 1. Load Configuration
	configFilePath := os.Getenv("DATABASE_CONFIG_FILE")
	if configFilePath == "" {
		configFilePath = "/etc/config/databases.yaml" // Default path
	}

	dbConfig, err := config.LoadConfig(configFilePath) // Use config package
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// 2. Kubernetes Client Initialization
	k8sClient, err := kubernetes.NewClient() // Use kubernetes package
	if err != nil {
		log.Fatalf("Error initializing Kubernetes client: %v", err)
	}

	// 3. Main Backup Loop
	for _, db := range dbConfig.Databases {
		fmt.Printf("Starting backup for: %s (%s)\n", db.Name, db.Type)

		var backupErr error // To store any backup errors

		switch db.Type {
		case "mongodb":
			backupErr = backup.BackupMongoDB(k8sClient, db) // Call MongoDB backup function
		case "mysql":
			backupErr = backup.BackupMySQL(k8sClient, db) // Call MySQL backup function
		// case "postgresql":
		// 	backupErr = backup.BackupPostgreSQL(k8sClient, db) // Call PostgreSQL backup function
		// case "minio":
		// 	backupErr = backup.BackupMinio(db) // Call Minio backup function
		// case "milvus":
		// 	backupErr = backup.BackupMilvus(db) // Call Milvus backup function
		default:
			log.Printf("Unsupported database type: %s", db.Type)
			continue // Skip to the next database
		}

		if backupErr != nil {
			log.Printf("Backup for %s (%s) failed: %v", db.Name, db.Type, backupErr)
			// Send failure notification (optional)
			// notification.SendSlackNotification(db, fmt.Sprintf("Backup for %s failed: %v", db.Name, backupErr))
			// notification.SendEmailNotification(db, fmt.Sprintf("Backup for %s failed: %v", db.Name, backupErr))

		} else {
			fmt.Printf("Backup for %s (%s) completed successfully.\n", db.Name, db.Type)
			// Send success notification (optional)
			// notification.SendSlackNotification(db, fmt.Sprintf("Backup for %s completed", db.Name))
			// notification.SendEmailNotification(db, fmt.Sprintf("Backup for %s completed", db.Name))

			// Upload to Cloud Storage (if backup was successful)
			// err = storage.UploadToCOS(db, db.Type)
			// if err != nil {
			// 	log.Printf("Upload to COS for %s failed: %v", db.Name, err)
			// } else {
			// 	fmt.Printf("Upload to COS for %s completed successfully.\n", db.Name)
			// }
		}
	}

	// 4. Graceful Shutdown (Optional but Recommended)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan // Block until a signal is received
	fmt.Println("Shutting down gracefully...")
}

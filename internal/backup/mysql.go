package backup

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"

	"database-backup/internal/config"
	k8s "database-backup/internal/kubernetes" // Import your kubernetes package

	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver
	"k8s.io/client-go/kubernetes"
)

// BackupMySQL performs a MySQL backup.
func BackupMySQL(k8sClient *kubernetes.Clientset, db config.Database) error {
	secret, err := k8s.GetSecret(k8sClient, db.Namespace, db.SecretName)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	password, ok := secret.Data["MYSQL_PASSWORD"]
	if !ok {
		return fmt.Errorf("MYSQL_PASSWORD not found in secret")
	}

	// Connect to the MySQL database (for testing connection)
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", db.Name, string(password), db.Host, db.Database)
	dbConn, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer dbConn.Close()

	err = dbConn.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}

	log.Printf("Successfully connected to MySQL database %s", db.Name)

	cmd := exec.Command("mysqldump", "-u"+db.Name, "-p"+string(password), "-h"+db.Host, db.Database)
	outfile, err := os.Create("/tmp/" + db.Name + ".sql")
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	cmd.Stderr = os.Stderr // Redirect stderr for debugging

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("MySQL backup failed: %w", err)
	}

	return nil
}

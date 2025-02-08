package backup

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"database-backup/internal/config"
	k8s "database-backup/internal/kubernetes"

	"k8s.io/client-go/kubernetes"
)

// BackupMongoDB performs a MongoDB backup
func BackupMongoDB(k8sClient *kubernetes.Clientset, db config.Database) error {

	cmd := exec.Command("mongodump", "--uri="+db.URI, "--archive=/tmp/"+db.Name+".archive")
	cmd.Stderr = os.Stderr

	if db.TLSSecret != "" {
		secret, err := k8s.GetSecret(k8sClient, db.Namespace, db.TLSSecret)
		if err != nil {
			return fmt.Errorf("failed to get secret: %w", err)
		}

		caCert, ok := secret.Data["ca.crt"]
		if !ok {
			return fmt.Errorf("ca.crt not found in secret %s", db.TLSSecret)
		}

		clientCert, ok := secret.Data["client.crt"]
		if !ok {
			return fmt.Errorf("client.crt not found in secret %s", db.TLSSecret)
		}

		clientKey, ok := secret.Data["client.key"]
		if !ok {
			return fmt.Errorf("client.key not found in secret %s", db.TLSSecret)
		}

		// Write certificates to temporary files
		caCertFile, err := ioutil.TempFile("", "ca.crt")
		if err != nil {
			return fmt.Errorf("error creating temporary ca.crt file: %w", err)
		}
		defer caCertFile.Close()
		_, err = caCertFile.Write(caCert)
		if err != nil {
			return fmt.Errorf("error writing ca.crt to temporary file: %w", err)
		}
		caCertPath := caCertFile.Name()

		clientCertFile, err := ioutil.TempFile("", "client.crt")
		if err != nil {
			return fmt.Errorf("error creating temporary client.crt file: %w", err)
		}
		defer clientCertFile.Close()
		_, err = clientCertFile.Write(clientCert)
		if err != nil {
			return fmt.Errorf("error writing client.crt to temporary file: %w", err)
		}

		clientCertPath := clientCertFile.Name()

		clientKeyFile, err := ioutil.TempFile("", "client.key")
		if err != nil {
			return fmt.Errorf("error creating temporary client.key file: %w", err)
		}
		defer clientKeyFile.Close()
		_, err = clientKeyFile.Write(clientKey)
		if err != nil {
			return fmt.Errorf("error writing client.key to temporary file: %w", err)
		}

		clientKeyPath := clientKeyFile.Name()

		cmd.Env = append(os.Environ(),
			"MONGODB_TLS_CA_FILE="+caCertPath,
			"MONGODB_TLS_CERTIFICATE_KEY_FILE="+clientCertPath,
			"MONGODB_TLS_KEY_FILE="+clientKeyPath)
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("MongoDB backup failed: %w", err)
	}

	return nil
}

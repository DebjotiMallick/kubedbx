#!/bin/bash
echo "MySQL Backup Script Starts......"
#this part will iterate for as many MySQL instances in Cluster - start

input="/dbbackup-scripts/mysql-db-list.txt"
while read -r line
do
	echo "--------------------------------------------------------"
	IFS=":"
	read -ra str <<< "$line"
	
	dbname="--all-databases"
	if [ ${str[2]} != "A" ]; then
	dbname=${str[2]}
	fi
	FILENAME="${str[0]}-${str[1]}-dump-$(date +"%d-%m-%Y_%s")"
	echo $FILENAME
	echo "MyQSL Backup Generates File /tmp/$FILENAME.sql.gz"

	service=${str[0]}
	namespace=${str[1]}
	export password=$(oc get secret -n $namespace $service -o jsonpath="{.data.mysql-root-password}" | base64 -d)
	
	dt=$(date '+%d/%m/%Y %H:%M:%S');
	echo "Dump Starts AT : $dt"

	mysqldump --single-transaction --quick -h $service.$namespace -u root -p$password $dbname | gzip > /tmp/$FILENAME.sql.gz
	if [ "$?" -eq 0 ]; then
		echo "mysqldump command Successful"    
		dt=$(date '+%d/%m/%Y %H:%M:%S');
		echo "MySQL Dump Ends AT : $dt"

		echo "check file size"
		FILESIZE=$(du -bs /tmp/$FILENAME.sql.gz | awk '{print $1}')
		SIZELIMIT=1073741824
		FILESIZE_MB=$(du -bs /tmp/$FILENAME.sql.gz | awk '{print $1/2^20}')
		echo "File Size : $FILESIZE_MB MB"

		if [[ $FILESIZE -gt $SIZELIMIT ]]; then 
			echo "Splits Large Backup Dump File"
			split -b 1024M /tmp/$FILENAME.sql.gz "/tmp/$FILENAME.sql.part"

			echo "------ List Splitted Files ------"
			ls -ltr /tmp/$FILENAME.sql.part*
			echo "------ List Splitted Files ------"
			
			for file in /tmp/$FILENAME.sql.part*; do
				uploadfile=$(basename "$file")
				echo "$uploadfile"
				echo "MySQL Backup Dump Split /tmp/$uploadfile is being Uploaded into cos-4-db-backup COS and roks-dev-mysqldbbackup bucket"
				nodejs /dbbackup-scripts/cos-connect.js /tmp/$uploadfile $uploadfile roks-dev-mysqldbbackup
			done
		else 
			echo "MySQL Backup Dump /tmp/$FILENAME.sql.gz is being Uploaded into cos-4-db-backup COS and roks-dev-mysqldbbackup bucket"
			nodejs /dbbackup-scripts/cos-connect.js /tmp/$FILENAME.sql.gz $FILENAME.sql.gz roks-dev-mysqldbbackup
		fi
	else
		echo "mysqldump encountered an Error"
		rm -f /tmp/$FILENAME
	fi
	
	echo "--------------------------------------------------------"
done < "$input"

echo "Uploading database backup logs into COS"
nodejs /dbbackup-scripts/cos-connect.js /tmp/mysql_backup_$(date +%F).log mysql_backup_$(date +%F).log roks-dev-mysqldbbackup
#this part will iterate for as many MySQL instances in Cluster - end

echo "MySQL Backup Script Ends......"
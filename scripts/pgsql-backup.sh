#!/bin/bash
echo "PGSQL Backup Script Starts......"
#this part will iterate for as many pgsql instances in Cluster - start

input="/dbbackup-scripts/pgsql-db-list.txt"
while read -r line
do
	echo "--------------------------------------------------------"
	IFS=":"
	read -ra str <<< "$line"
	
	service=${str[0]}
	namespace=${str[1]}

	FILENAME="$service-$namespace-dump-$(date +"%d-%m-%Y_%s")"
	echo $FILENAME
	echo "PGSQL Backup Generates File /tmp/$FILENAME.gz"

	dt=$(date '+%d/%m/%Y %H:%M:%S');
	echo "Dump Starts AT : $dt"
	export password=$(oc get secret -n $namespace $service -o jsonpath="{.data.postgres-password}" | base64 -d)
	PGPASSWORD=$password pg_dumpall -h $service.$namespace -U postgres | gzip > /tmp/$FILENAME.gz
	if [ "$?" -eq 0 ]; then
		echo "pgsqldump command Successful"
		dt=$(date '+%d/%m/%Y %H:%M:%S');
		echo "pgsql Dump Ends AT : $dt"

		echo "check file size"
		FILESIZE=$(du -bs /tmp/$FILENAME.gz | awk '{print $1}')
		SIZELIMIT=1073741824
		FILESIZE_MB=$(du -bs /tmp/$FILENAME.gz | awk '{print $1/2^20}')
		echo "File Size : $FILESIZE_MB MB"

		if [[ $FILESIZE -gt $SIZELIMIT ]]; then 
			echo "Splits Large Backup Dump File"
			split -b 1024M /tmp/$FILENAME.gz "/tmp/$FILENAME.part"

			echo "------ List Splitted Files ------"
			ls -ltr /tmp/$FILENAME.part*
			echo "------ List Splitted Files ------"
			
			for file in /tmp/$FILENAME.part*; do
				uploadfile=$(basename "$file")
				echo "$uploadfile"
				echo "pgsql Backup Dump Split /tmp/$uploadfile is being Uploaded into cos-4-db-backup COS and roks-dev-pgsqlbackup bucket"
				nodejs /dbbackup-scripts/cos-connect.js /tmp/$uploadfile $uploadfile roks-dev-pgsqlbackup
			done
		else 
			echo "pgsql Backup Dump /tmp/$FILENAME.gz is being Uploaded into cos-4-db-backup COS and roks-dev-pgsqlbackup bucket"
			nodejs /dbbackup-scripts/cos-connect.js /tmp/$FILENAME.gz $FILENAME.gz roks-dev-pgsqlbackup
		fi
	else
		echo "pgsqldump encountered an Error"
		rm -f /tmp/$FILENAME
	fi
	echo "--------------------------------------------------------"
done < "$input"

echo "Uploading database backup logs into COS"
nodejs /dbbackup-scripts/cos-connect.js /tmp/pgsql_backup_$(date +%F).log pgsql_backup_$(date +%F).log roks-dev-pgsqlbackup
#this part will iterate for as many pgsql instances in Cluster - end

echo "pgsql Backup Script Ends......"
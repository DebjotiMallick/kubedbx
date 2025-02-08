#!/bin/bash
echo "MinIO Backup Script Starts......"
#this part will iterate for as many minio instances in Cluster - start

input="/dbbackup-scripts/minio-db-list.txt"
while read -r line
do
	echo "--------------------------------------------------------"
	IFS=":"
	read -ra str <<< "$line"
	
	service=${str[0]}
	namespace=${str[1]}

	FILENAME="$service-$namespace-dump-$(date +"%d-%m-%Y_%s")"
	echo $FILENAME

	dt=$(date '+%d/%m/%Y %H:%M:%S');
	echo "Dump Starts AT : $dt"
	# script goes here
	export rootuser=$(oc get secret $service -n $namespace -o=jsonpath="{.data.rootUser}" | base64 -d)
	export rootpass=$(oc get secret $service -n $namespace -o=jsonpath="{.data.rootPassword}" | base64 -d)
	mc alias set --insecure my_minio https://$service.$namespace:9000 $rootuser $rootpass
	mc cp -r --insecure my_minio/ /tmp/$FILENAME/
	tar -czvf /tmp/$FILENAME.gz /tmp/$FILENAME/
	if [ "$?" -eq 0 ]; then
		echo "MinIO command Successful"
		dt=$(date '+%d/%m/%Y %H:%M:%S');
		echo "MinIO Dump Ends AT : $dt"

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
				echo "MinIO Backup Dump Split /tmp/$uploadfile is being Uploaded into cos-4-db-backup COS and talktocorpus-dev-miniobackup bucket"
				nodejs /dbbackup-scripts/cos-connect.js /tmp/$uploadfile $uploadfile talktocorpus-dev-miniobackup
			done
		else 
			echo "MinIO Backup Dump /tmp/$FILENAME.gz is being Uploaded into cos-4-db-backup COS and talktocorpus-dev-miniobackup bucket"
			nodejs /dbbackup-scripts/cos-connect.js /tmp/$FILENAME.gz $FILENAME.gz talktocorpus-dev-miniobackup
		fi
	else
		echo "MinIO dump encountered an Error"
		rm -f /tmp/$FILENAME
	fi
	echo "--------------------------------------------------------"
done < "$input"

echo "Uploading database backup logs into COS"
nodejs /dbbackup-scripts/cos-connect.js /tmp/minio_backup_$(date +%F).log minio_backup_$(date +%F).log talktocorpus-dev-miniobackup
#this part will iterate for as many pgsql instances in Cluster - end

echo "MinIO Backup Script Ends......"
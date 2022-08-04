Content-Type: multipart/mixed; boundary="//"
MIME-Version: 1.0
Content-Type: text/cloud-config; charset="us-ascii"
MIME-Version: 1.0
Content-Transfer-Encoding: 7bit
Content-Disposition: attachment; filename="cloud-config.txt"
cloud_final_modules:
- [scripts-user, always]
Content-Type: text/x-shellscript; charset="us-ascii"
MIME-Version: 1.0
Content-Transfer-Encoding: 7bit
Content-Disposition: attachment; filename="userdata.txt"
#!/bin/bash
yum remove docker docker-engine docker.io containerd runc
yum update -y
yum install docker -y
systemctl enable docker.service
systemctl start docker.service
docker stop fragbot
docker rm fragbot
docker pull ishaanrao/fragbots:latest
docker run ishaanrao/fragbots:latest --name fragbot -e ACCESS_TOKEN -e AUTHKEY -e BACKEND_URI -e BOT_ID
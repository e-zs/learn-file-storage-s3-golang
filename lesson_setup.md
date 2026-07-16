# Install gcc
```bash
sudo apt install gcc
```
Ensure the environment variable CGO_ENABLED is set to 1:
```bash
go env CGO_ENABLED

# If the command above prints 0, run this:
go env -w CGO_ENABLED=1
```

# Download samples
./samplesdownload.sh

# Install SQLite
```bash
sudo apt update
sudo apt install sqlite3
```
Connect / Exit
```bash
sqlite3 tubely.db
.exit
```

# AWS CLI
CH3-L2
https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html

To install the AWS CLI, run the following commands.
```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
```

# AWS S3 Go SDK
https://github.com/aws/aws-sdk-go-v2
```bash
go get github.com/aws/aws-sdk-go-v2/service/s3 github.com/aws/aws-sdk-go-v2/config
```
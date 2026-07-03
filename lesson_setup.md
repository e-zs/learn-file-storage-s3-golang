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

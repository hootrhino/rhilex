#!/bin/bash
RESET='\033[0m'
RED='\033[31m'
BLUE='\033[34m'
YELLOW='\033[33m'

# 打印红色文本
echo_red() {
    echo -e "${RED}$1${RESET}"
}

# 打印蓝色文本
echo_blue() {
    echo -e "${BLUE}$1${RESET}"
}

# 打印黄色文本
echo_yellow() {
    echo -e "${YELLOW}$1${RESET}"
}
install(){
    local source_dir="$PWD"
    local service_file="/etc/systemd/system/rhilex.service"
    local executable="/usr/local/rhilex"
    local WORKING_DIRECTORY="/usr/local/"
    local config_file="/usr/local/rhilex.ini"
    local db_file="/usr/local/rhilex.db"
cat > "$service_file" << EOL
[Unit]
Description=rhilex Daemon
After=network.target

[Service]
Environment="ARCHSUPPORT=EEKITH3"
WorkingDirectory=$WORKING_DIRECTORY
ExecStart=$executable run -config=$config_file -db=$db_file
ConditionPathExists=!/var/run/rhilex-upgrade.lock
Restart=always
User=root
Group=root
StartLimitInterval=0
RestartSec=5
[Install]
WantedBy=multi-user.target
EOL
    chmod +x $source_dir/rhilex
    echo "[.] Copy $source_dir/rhilex to $WORKING_DIRECTORY."
    cp "$source_dir/rhilex" "$executable"
    echo "[.] Copy $source_dir/rhilex.ini to $WORKING_DIRECTORY."
    cp "$source_dir/rhilex.ini" "$config_file"
    echo "[.] Copy $source_dir/license.key to /usr/local/license.key."
    cp "$source_dir/license.key" "/usr/local/license.key"
    echo "[.] Copy $source_dir/license.lic to /usr/local/license.lic."
    cp "$source_dir/license.lic" "/usr/local/license.lic"
    systemctl daemon-reload
    systemctl enable rhilex
    systemctl start rhilex
    if [ $? -eq 0 ]; then
        echo "[√] rhilex service has been created and extracted."
    else
        echo "[x] Failed to create the rhilex service or extract files."
    fi
    exit 0
}

start(){
    systemctl daemon-reload
    systemctl start rhilex
    echo "[√] rhilex started as a daemon."
    exit 0
}
status(){
    systemctl status rhilex
}
restart(){
    systemctl stop rhilex
    start
}

stop(){
    systemctl stop rhilex
    echo "[√] Service rhilex has been stopped."
}
remove_files() {
    if [ -e "$1" ]; then
        if [[ $1 == *"/upload"* ]]; then
            rm -rf "$1"
        else
            rm "$1"
        fi
        echo "[!] $1 files removed."
    else
        echo "[*] $1 files not found. No need to remove."
    fi
}

uninstall(){
    systemctl stop rhilex
    systemctl disable rhilex
    remove_files /etc/systemd/system/rhilex.service
    remove_files $WORKING_DIRECTORY/rhilex
    remove_files $WORKING_DIRECTORY/rhilex.ini
    remove_files $WORKING_DIRECTORY/rhilex.db
    remove_files $WORKING_DIRECTORY/upload/
    remove_files $WORKING_DIRECTORY/license.key
    remove_files $WORKING_DIRECTORY/license.lic
    rm -f "$WORKING_DIRECTORY/*.txt"
    rm -f "$WORKING_DIRECTORY/*.txt.gz"
    systemctl daemon-reload
    systemctl reset-failed
    echo "[√] rhilex has been uninstalled."
}
#
#
#
main(){
    case "$1" in
        "install" | "start" | "restart" | "stop" | "uninstall" | "create_user" | "status")
            $1
        ;;
        *)
            echo "[x] Invalid command: $1"
            echo "[?] Usage: $0 <install|start|restart|stop|uninstall|status>"
            exit 1
        ;;
    esac
    exit 0
}
#===========================================
# main
#===========================================
if [ "$(id -u)" != "0" ]; then
    echo "[x] This script must be run as root"
    exit 1
else
    main $1
fi
#!/bin/bash

SERVICE_NAME="rhilex"
WORKING_DIRECTORY="/usr/local"
EXECUTABLE_PATH="$WORKING_DIRECTORY/$SERVICE_NAME"
CONFIG_PATH="$WORKING_DIRECTORY/$SERVICE_NAME.ini"

SERVICE_FILE="/etc/systemd/system/rhilex.service"

STOP_SIGNAL="/var/run/rhilex-stop.sinal"
UPGRADE_SIGNAL="/var/run/rhilex-upgrade.lock"

SOURCE_DIR="$PWD"

install(){
cat > "$SERVICE_FILE" << EOL
[Unit]
Description=rhilex Daemon
After=network.target

[Service]
Environment="ARCHSUPPORT=EEKITH3"
WorkingDirectory=$WORKING_DIRECTORY
ExecStart=$EXECUTABLE_PATH run
ConditionPathExists=!/var/run/rhilex-upgrade.lock
Restart=always
User=root
Group=root
StartLimitInterval=0
RestartSec=5
[Install]
WantedBy=multi-user.target
EOL
    chmod +x $SOURCE_DIR/rhilex
    echo "[.] Copy $SOURCE_DIR/rhilex to $WORKING_DIRECTORY."
    cp "$SOURCE_DIR/rhilex" "$EXECUTABLE_PATH"
    echo "[.] Copy $SOURCE_DIR/rhilex.ini to $WORKING_DIRECTORY."
    cp "$SOURCE_DIR/rhilex.ini" "$config_file"
    echo "[.] Copy $SOURCE_DIR/license.key to /usr/local/license.key."
    cp "$SOURCE_DIR/license.key" "/usr/local/license.key"
    echo "[.] Copy $SOURCE_DIR/license.lic to /usr/local/license.lic."
    cp "$SOURCE_DIR/license.lic" "/usr/local/license.lic"
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
    echo "[√] RHILEX started as a daemon."
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
    remove_files "$SERVICE_FILE"
    remove_files "$WORKING_DIRECTORY/rhilex"
    remove_files "$WORKING_DIRECTORY/rhilex.ini"
    remove_files "$WORKING_DIRECTORY/rhilex.db"
    remove_files "$WORKING_DIRECTORY/license.lic"
    remove_files "$WORKING_DIRECTORY/license.key"
    remove_files "$WORKING_DIRECTORY/rhilex_internal_datacenter.db"
    remove_files "$WORKING_DIRECTORY/upload/"
    remove_files "$WORKING_DIRECTORY/rhilexlog.txt"
    remove_files "$WORKING_DIRECTORY/rhilex-daemon-log.txt"
    remove_files "$WORKING_DIRECTORY/rhilex-recover-log.txt"
    remove_files "$WORKING_DIRECTORY/rhilex-upgrade-log.txt"
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
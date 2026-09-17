# #!/bin/bash
set -eu

# Resolve paths (.env, etc.) relative to this script, not the caller's cwd —
# `ssh root@host "bash /root/setup/01.sh"` starts in /root, not /root/setup.
cd "$(dirname "$0")"

# Install dos2unix to format env file
apt --yes install dos2unix
dos2unix .env

source .env

# ==================================================================================== #
# VARIABLES
# ==================================================================================== #

# Set the timezone for the server. A full list of available timezones can be found by 
# running timedatectl list-timezones.
TIMEZONE=Europe/Paris

# Force all output to be presented in en_US for the duration of this script. This avoids  
# any "setting locale failed" errors while this script is running, before we have 
# installed support for all locales. Do not change this setting!
export LC_ALL=en_US.UTF-8 

# ==================================================================================== #
# SCRIPT LOGIC
# ==================================================================================== #

# Enable the "universe" repository.
add-apt-repository --yes universe

# Update all software packages.
apt update

# Set the system timezone and install all locales.
timedatectl set-timezone ${TIMEZONE}
apt --yes install locales-all

# Add the new user (and give them sudo privileges). Skipped if a re-run finds it already there.
id -u "${USER}" >/dev/null 2>&1 || useradd --create-home --shell "/bin/bash" --groups sudo "${USER}"

# This user is SSH-key only: no password at all (rather than one nobody knows),
# and passwordless sudo — a forced/interactive password flow would block every
# non-interactive deploy step (including this script's own later use of the
# account), since there is no password to authenticate with.
passwd --delete "${USER}"
echo "${USER} ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/${USER}"
chmod 440 "/etc/sudoers.d/${USER}"

# Copy the SSH keys from the root user to the new user.
rsync --archive --chown=${USER}:${USER} /root/.ssh /home/${USER}

# Configure the firewall to allow SSH, HTTP and HTTPS traffic.
ufw allow 22
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# Install fail2ban.
apt --yes install fail2ban

# Install the migrate CLI tool.
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
mv migrate.linux-amd64 /usr/local/bin/migrate

# Install Caddy (see https://caddyserver.com/docs/install#debian-ubuntu-raspbian).
apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --batch --yes --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
apt update
apt --yes install caddy

# Install Docker
## Add Docker's official GPG key:
apt install -y ca-certificates curl
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc
## Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null
apt update
## Install
apt install -y  docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
docker volume create halendar-db
docker volume create halendar-api
docker volume create halendar-ollama
docker volume create config_files
## Add user to Docker group (takes effect on their next login; "newgrp" here
## would just hang since this script runs non-interactively as root)
sudo usermod -aG docker ${USER}

# Upgrade all packages. Using the --force-confnew flag means that configuration 
# files will be replaced if newer ones are available.
apt --yes -o Dpkg::Options::="--force-confnew" upgrade

echo "Script complete! Rebooting..."
reboot
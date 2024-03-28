#!/bin/bash

# ========================
# VARIABLES
# ========================

# Set timezone
TIMEZONE=Europe/Berlin

# Set username
USERNAME=greenlight

# Prompt to enter password for PostgreSQL user
# (rather than hard-coding a password in this script)
read -p "Enter password for greenlight DB user: " DB_PASSWORD

# Force all output to en_US for the duration of this script.
# This avoids any "setting local failed" errors.
export LC_ALL=en_US.UTF-8

# ========================
# SCRIPT LOGIC
# ========================

# Enable the "universe" repository
add-apt-repository --yes universe

# Update all software packages
apt update

# Set timezne and install locales
timedatectl set-timezone ${TIMEZONE}
apt --yes install locales-all

# Add new user and give sudo privileges
useradd --create-home --shell "/bin/bash" --groups sudo "${USERNAME}"

# Force password to be set for the new user the first time they log in
passwd --delete "${USERNAME}"
chage --lastday 0 "${USERNAME}"

# Copy the SSH keys from root to the new user
rsync --archive --chown=${USERNAME}:${USERNAME} /root/.ssh /home/${USERNAME}

# Configure firewall to allow SSH, HTTP, and HTTPS traffic
ufw allow 22
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# Install fail2ban
apt --yes install fail2ban

# Install migrate cli
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
mv migrate.linux-amd64 /usr/local/bin/migrate

# Install PostgreSQL
apt --yes install postgresql

# Set up the greenlight DB and create a user account with the password entered earlier.
sudo -i -u postgres psql -c "CREATE DATABASE greenlight"
sudo -i -u postgres psql -d greenlight -c "CREATE EXTENSION IF NOT EXISTS citext"
sudo -i -u postgres psql -d greenlight -c "CREATE ROLE greenlight WITH LOGIN PASSWORD '${DB_PASSWORD}'"

# Add a DSN for connecting to the greenlight database to the system-wide environment
# variables in the /etc/environment file.
echo "GREENLIGHT_DB_DSN='postgres://greenlight:${DB_PASSWORD}@localhost/greenlight'" >> /etc/environment

# Install Caddy (see https://caddyserver.com/docs/install#debian-ubuntu-raspbian).
apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-arc curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
apt update
apt --yes install caddy

# Upgrade all packages. Using the --force-confnew flag means that configuration # files will be replaced if newer ones are available.
apt --yes -o Dpkg::Options::="--force-confnew" upgrade
echo "Script complete! Rebooting..."
reboot

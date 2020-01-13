# jq is used to escape string for json
wget https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64

# download the go package
wget https://dl.google.com/go/go1.13.1.linux-amd64.tar.gz
# then uncompress it and use the go program in the bin/ folder

# install git
sudo apt-get -y update
sudo apt-get install git

# cloning cothority and odyssey in the GitHub folder
# building bcadmin, csadmin, and pcadmin
# $...

# Execute the startup script at boot
# in /etc/rc.local
/home/enclave/startup.sh >> /home/enclave/rclogs.log 2>&1


# Only allow ssh connection and no other tcp/udp outbounds/inbounds
# one can use "sudo ufw disable" to temporarly download stuff.
sudo ufw disable
sudo ufw reset
sudo ufw default deny incoming
# needed for the logs and communication to cothority
sudo ufw default allow outgoing
sudo ufw allow 22/tcp
sudo ufw enable

# To add a non-login user, this is the user that we will give access to via ssh
sudo adduser scientist --disabled-password
#  --disabled-password: logins are still possible (for example using SSH RSA
#  keys) but not  using password authentication.

# Create the folder that will contain the datasets
cd /home/scientist/
sudo -u scientist mkdir -p python_project/datasets
# Since the 'enclave' user will have to get write access to this folder, we set
# the group with 7. The 'enclave' user will then be added in the 'scientist'
# group
sudo chmod 770 datasets
# Add the 'enclave' user to the 'scientist' group
sudo usermod -aG scientist enclave # may need to reload to take effect...
# Then copy the datasets
cp /home/enclave/datasets/* /home/scientist/datasets/

# create ssh stuff for the 'scientist' so that we can remotely connect via ssh
# using the provided RSA key.
cd /home/scientist
sudo -u scientist mkdir .ssh
sudo chmod 700 .ssh/
sudo -u scientist touch .ssh/authorized_keys
sudo chmod 600 .ssh/authorized_keys
# then add the public key


# To connect to the jupyter remotely via ssh:
ssh -i id_rsa -L 8888:localhost:8888 scientist@164.128.172.29
# then on the host
cd python_project
jupyter notebook
# Then on the host, one can access localhost:8888?token=...

# install conda
CHECKSUM=46d762284d252e51cd58a8ca6c8adc9da2eadc82c342927b2f66ed011d1d8b53
wget -O /tmp/conda_installer.sh https://repo.anaconda.com/archive/Anaconda3-2019.10-Linux-x86_64.sh
echo "$CHECKSUM /tmp/conda_installer.sh" | sha256sum -c -
> check if its ok
chmod +x /tmp/conda_installer.sh 
sudo -u scientist /tmp/conda_installer.sh -b -p /home/scientist/python_project/anaconda3
sudo vim /home/scientist/.bashrc
> add the following at the TOP: export PATH=/home/scientist/python_project/anaconda3/bin:$PATH
# install geopandas for the scientist
sudo -u scientist /home/scientist/python_project/anaconda3/bin/conda install geopandas


# Instructions to build custom scp
# Copy over the patch to /tmp/pii.patch
# Ensure you have the dependencies installed:
sudo apt-get update
sudo apt install build-essential libssl-dev zlib1g-dev
# Get OpenSSH source code from i.e. https://cdn.openbsd.org/pub/OpenBSD/OpenSSH/portable/
wget https://cdn.openbsd.org/pub/OpenBSD/OpenSSH/portable/openssh-8.1p1.tar.gz
tar zxvf openssh-8.1p1.tar.gz
cd openssh-8.1p1
patch -p1 < /tmp/pii.patch
sh configure
make scp
sudo cp ./scp /usr/bin/

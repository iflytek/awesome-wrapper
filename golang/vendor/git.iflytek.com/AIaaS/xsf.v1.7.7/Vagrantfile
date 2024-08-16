Vagrant.configure('2') do |config|
  config.vm.box = "ubuntu/focal"
  config.vm.network "private_network", type: "dhcp"
  config.vm.box_check_update = false
  config.vm.hostname = "testing"
  config.vm.network "forwarded_port", guest: 1995, host: 1995

  # fix issues with slow dns http://serverfault.com/a/595010
  config.vm.provider :virtualbox do |vb, override|
      vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
      vb.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
      vb.customize ["modifyvm", :id, "--memory", "4096"]
      vb.customize ["modifyvm", :id, "--cpus", 2]
  end

  config.ssh.username = "root"
  config.ssh.password = "9527"

  config.vm.provision "shell", inline: <<-SHELL
    echo 'alias sync="rsync -av --delete /vagrant ~"' >>/etc/profile
  SHELL
end
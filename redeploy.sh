gox -osarch="linux/amd64"
ansible-playbook ./ansible/deploy/all.yml -i ./ansible/hosts/live -vvvv

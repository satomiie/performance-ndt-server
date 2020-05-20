This repository contains an Ansible playbook that install and configure a ndt5/ndt7 server written in Go and iPerf listening on port TCP 5001

Clone Git Repository to your Ansible Controller

Edit Inventory file with your target host (replace X.X.X.X with your destination server)

On host_vars folder you will define variables for ndt-server destination folder (variable path_ndt). You will change that variable in order to change destination folder of the service.

Run playbook using the command : "ansible-playbook playbook-perfserver.yaml -k -K -i inventory"

In order to access on services, please use following links:

ndt5+wss: https://localhost:3010/static/widget.html

ndt7: https://localhost/static/ndt7.html

Replace localhost with the IP of the server to access them externally.

iPerf is available running following command : iperf -c "server IP"

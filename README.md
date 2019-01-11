[![GoDoc](https://godoc.org/github.com/m-lab/ndt-server?status.svg)](https://godoc.org/github.com/m-lab/ndt-server) [![Build Status](https://travis-ci.org/m-lab/ndt-server.svg?branch=master)](https://travis-ci.org/m-lab/ndt-server) [![Coverage Status](https://coveralls.io/repos/github/m-lab/ndt-server/badge.svg?branch=master)](https://coveralls.io/github/m-lab/ndt-server?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/m-lab/ndt-server)](https://goreportcard.com/report/github.com/m-lab/ndt-server)

# ndt-server

To run the server locally, first run `gen_local_test_certs.sh`, and then run the
commands
```bash
docker build . -t ndt-server
```
and
```bash
docker run --net=host -v `pwd`:/certs -it -t ndt-server \
    -cert /certs/cert.pem -key /certs/key.pem
```

Once you have done that, you should have a server running on port 3010 on
localhost with metrics available on port 9090.

Try running a test in your browser (certs will appear invalid to your
browser, but everything is safe because it's running locally):

* https://localhost:3010/static/widget.html
* http://localhost:9090/metrics

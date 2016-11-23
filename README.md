External ip controller [![Build Status](https://travis-ci.org/Mirantis/k8s-externalipcontroller.svg?branch=master)](https://travis-ci.org/Mirantis/k8s-externalipcontroller)
======================

[![asciicast](https://asciinema.org/a/93841.png)](https://asciinema.org/a/93841)

How to deploy?
===============
If you want to run ipcontroller with failure tolerance use examples/claims.
Specify link which will be managed by ipcontroller.
```
kubectl apply -f examples/claims/
```

Simple version without failure tolerance can be deployed using:
```
kubectl apply -f examples/simple/
```


How to run tests?
================

Install dependencies and prepare kube dind cluster
```
make get-deps
```

Build necessary images and run tests
```
make test
```

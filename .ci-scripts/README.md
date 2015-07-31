Scripts for Continuous Integration
==================================

The scripts in this folder are designed to be integrated in Travis-CI.

They map the main hooks provided by Travis-CI, so that forkers can customize their processes.

For example, Golang forks of `go get`able repos can use before_install.sh to avoid `go get` of 
upstream imports, using an environment variable UPSTREAM_OWNER containing the Github name of 
the owner of the upstream repo.


Configuring for dev for novice Go developers
This utilized Jetbrains GoLandD
Dfils notes this uses the dep tools

Download go
download dep
https://golang.github.io/dep/docs/installation.html
download GoLand

Docker: 
you need to be running:
mino server


create a directory for Go development, eg: dev_earthcube/go/src
Open goloand, configure gopath for new projects: file>other settings>Settings for new projects
In Go, set gopath to directory created above:

Now checkout from source VCS>Checkout from Version Control
https://github.com/earthcubearchitecture-project418/gleaner
save to> earthcube.org/Project418/gleaner
dev_earthcube/go/src/earthcube.org/Project418/gleaner

GoLand will ask about dependencies, enable, select the dep app/program/executable

add configuration, add Go Build, select Package
cmd /gleaner/


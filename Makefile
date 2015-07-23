

SHELL = /bin/bash

run:
	go build
	./go-mdism

startserver:
	python app.py&

build:
	ubuntu@ip-172-31-58-182:~/go-mdism$ sudo nano ~/.profile
ubuntu@ip-172-31-58-182:~/go-mdism$ source ~/.profile
ubuntu@ip-172-31-58-182:~/go-mdism$ go get github.com/cheggaaa/pb
ubuntu@ip-172-31-58-182:~/go-mdism$ github.com/codegangsta/cli
-bash: github.com/codegangsta/cli: No such file or directory
ubuntu@ip-172-31-58-182:~/go-mdism$ go get github.com/codegangsta/cli
ubuntu@ip-172-31-58-182:~/go-mdism$ go get github.com/davecheney/profile
ubuntu@ip-172-31-58-182:~/go-mdism$ go get github.com/leesper/go_rng



	sudo apt-get install golang git make
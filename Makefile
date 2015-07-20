

SHELL = /usr/local/bin/fish

run:
	go build
	./go-mdism

startserver:
	python app.py&
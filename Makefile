build:
	go install

fmt:
	gofmt -w *.go
	colcheck *.go

tags:
	find ./ -name '*.go' -print0 | xargs -0 gotags > TAGS

loc:
	find ./ -name '*.go' -print0 | xargs -0 wc -l

push:
	git push origin master
	git push github master


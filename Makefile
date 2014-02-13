build:
	go install

fmt:
	find ./ -name '*.go' -print0 | xargs -0 gofmt -w
	find ./ -name '*.go' -print0 | xargs -0 colcheck

tags:
	find ./ -name '*.go' -print0 | xargs -0 gotags > TAGS

loc:
	find ./ -name '*.go' -print0 | xargs -0 wc -l

push:
	git push origin master
	git push github master


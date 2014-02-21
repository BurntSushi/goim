REMOTE=geils:~/www/burntsushi.net/public_html/stuff/goim/

build:
	go install

er:
	./scripts/goim-write-erd > /tmp/goim.er
	erd -i /tmp/goim.er -o /tmp/goim.pdf
	erd -i /tmp/goim.er -o /tmp/goim.png

	rsync /tmp/goim*{pdf,png} $(REMOTE)

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


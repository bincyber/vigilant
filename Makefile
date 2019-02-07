CONTAINER=vigilant
VERSION=v0.1.0

install:
	go get -v

test:
	go test -cover

build:
	go build -o vigilant

build-container:
	docker build -t $(CONTAINER):$(VERSION) .

run-container:
	docker run -d -p 8000:8000 --name $(CONTAINER) $(CONTAINER):$(VERSION)

stop-container:
	docker rm -f $(CONTAINER)

tag-container:
	docker tag $(CONTAINER):$(VERSION) bincyber/$(CONTAINER):$(VERSION)

upload-container:
	docker push bincyber/$(CONTAINER):$(VERSION)

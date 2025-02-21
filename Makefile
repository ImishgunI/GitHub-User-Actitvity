
all: clean github-activity

github-activity:
	go build -gcflags "all=-N -l" -o $@ main.go

clean:
	rm -f github-activity


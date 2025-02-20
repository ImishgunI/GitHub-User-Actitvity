
all: clean github-activity

github-activity:
	go build -o $@ main.go

clean:
	rm -f github-activity


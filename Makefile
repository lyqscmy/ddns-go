all: build

build: main.go
	go build -o ddns

install:
	cp ddns /usr/local/bin
	cp ddns.json /etc/
	cp ddns.timer /usr/lib/systemd/system/
	cp ddns.service /usr/lib/systemd/system/
	systemctl daemon-reload
	systemctl enable ddns.timer
	systemctl enable ddns.service

status:
	systemctl status ddns.timer
	systemctl status ddns.service

clean:
	rm ddns
	
phony: clean, status
BUILD=dist
DEPLOYMENT=$(BUILD).zip

all: 
	go build --buildmode=plugin -o plugins/print.so plugins/print.go
	bee run
	#go test tests/default_test.go
run: 
	#make test
	#bee run -downdoc=true -gendoc=true

pack:
	bee pack

test:
	cp -pr conf tests/
	go test tests/default_test.go
	rm -rf tests/conf

.PHONY: clean
clean:
	rm -f $(BUILD) $(DEPLOYMENT)
	rm -f lambda.log

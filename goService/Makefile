ifndef APP_PATH
APP_PATH=$(PWD)
endif

gen-doc-docker:
	docker run -it --rm --name apidoc -v $(APP_PATH):/root -w /root dmitrymomot/apidoc -i /root/src/apiPoc –o /root/doc –v

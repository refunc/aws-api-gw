SHELL := $(shell which bash) # ensure bash is used
export BASH_ENV=scripts/common

# populate vars
$(shell source scripts/version; env | grep -E '_VERSION|_IMAGE|REGISTRY_PREFIX' >.env)
include .env

BINS := aws-api-gw

GOOS := $(shell eval $$(go env); echo $${GOOS})
ARCH := $(shell eval $$(go env); echo $${GOARCH})

LD_FLAGS := -X github.com/refunc/aws-api-gw/pkg/version.Version=$(AWS_AIP_GW_VERSION)

clean:
	rm -rf bin/*

ifneq ($(GOOS),linux)
images: clean dockerbuild
	export GOOS=linux; make $@
else
images: $(addsuffix -image, $(BINS))
endif

bins: $(BINS)

bin/$(GOOS):
	mkdir -p $@

$(BINS): % : bin/$(GOOS) bin/$(GOOS)/%
	@log_info "Build: $@"

bin/$(GOOS)/%:
	@echo GOOS=$(GOOS)
	CGO_ENABLED=0 go build \
	-tags netgo -installsuffix netgo \
	-ldflags "-s -w $(LD_FLAGS)" \
	-a \
	-o $@ \
	./cmd/$*/

ifneq ($(GOOS),linux)
%-image:
	export GOOS=linux; make $@
else
%-image: % package/Dockerfile
	@rm package/$* 2>/dev/null || true && cp bin/linux/$* package/
	cd package \
	&& docker build \
	--build-arg https_proxy="$${HTTPS_RPOXY}" \
	--build-arg http_proxy="$${HTTP_RPOXY}" \
	--build-arg BIN_TARGET=$* \
	-t $(TARGET_IMAGE) .
	@log_info "Image: $(TARGET_IMAGE)"
endif

bin/$(GOOS)/aws-api-gw: $(shell find pkg -type f -name '*.go') $(shell find cmd -type f -name '*.go')
aws-api-gw-image: TARGET_IMAGE=$(AWS_AIP_GW_IMAGE)

push: images
	@log_info "start pushing images"; \
	docker push $(AWS_AIP_GW_IMAGE); \
	log_info "tag aws-api-gw to latest"; \
	docker tag $(AWS_AIP_GW_IMAGE) $(REGISTRY_PREFIX)aws-api-gw:latest && \
	docker push $(REGISTRY_PREFIX)aws-api-gw:latest

build-container:
	docker build -t refunc:build -f Dockerfile.build .

dockerbuild: build-container
	@log_info "make bins in docker"
	@docker run --rm -it -v $(shell pwd):/github.com/refunc/aws-api-gw refunc:build make bins

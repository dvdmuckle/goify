.PHONY: go-build
go-build:
	go build -o spc
.PHONY: go-vendor
go-vendor:
	go mod vendor
.PHONY: rpm-build
rpm-build:
	mkdir -p rpm-build
ifdef mock-version
	mock --scm-enable --scm-option method=git --scm-option package=spc --scm-option spec=spc.spec --scm-option branch=dev --scm-option write_tar=True --scm-option git_get='git clone https://github.com/dvdmuckle/spc.git' --enable-network --resultdir rpm-build -r ${mock-version}
else 
	mock --scm-enable --scm-option method=git --scm-option package=spc --scm-option spec=spc.spec --scm-option branch=dev --scm-option write_tar=True --scm-option git_get='git clone https://github.com/dvdmuckle/spc.git' --enable-network --resultdir rpm-build
endif
.PHONY: rpm-build-all-arch
VERSION_ID = $(shell cat /etc/os-release | grep VERSION_ID | cut -d '=' -f2)
rpm-build-all-arch:
	for arch in $(shell ls /etc/mock/ | grep fedora-${VERSION_ID} | cut -d '-' -f3 | cut -d '.' -f1); do \
		mock --scm-enable --scm-option method=git --scm-option package=spc --scm-option spec=spc.spec --scm-option branch=dev --scm-option write_tar=True --scm-option git_get='git clone https://github.com/dvdmuckle/spc.git' --enable-network --resultdir rpm-build-$$arch -r fedora-${VERSION_ID}-$$arch ; \
	done
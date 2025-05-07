FROM  registry.opensuse.org/opensuse/tumbleweed:latest as builder
MAINTAINER "Christian Goll <cgoll@suse.com>"
RUN  zypper ar https://download.opensuse.org/repositories/science:/machinelearning/openSUSE_Tumbleweed/science:machinelearning.repo &&\
   zypper --gpg-auto-import-keys ref &&\
   zypper install -y go git faiss-devel libfaiss findutils
WORKDIR /kowalski
COPY . .
RUN go mod tidy && go mod vendor
RUN go build kowalski.go

ENV KW_DATABASE /suseDoc
ENV KW_URL http://ollama:11434

ENTRYPOINT "/kowalski/kowalski"

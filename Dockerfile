FROM  registry.opensuse.org/opensuse/tumbleweed:latest AS builder
MAINTAINER "Christian Goll <cgoll@suse.com>"
RUN  zypper ar https://download.opensuse.org/repositories/science:/machinelearning/openSUSE_Tumbleweed/science:machinelearning.repo &&\
   zypper --gpg-auto-import-keys ref &&\
   zypper install -y go git faiss-devel libfaiss findutils
WORKDIR /kowalski
COPY . .
RUN go mod tidy && go mod vendor
RUN go build kowalski.go

FROM  registry.opensuse.org/opensuse/tumbleweed:latest AS BUILDER
RUN  zypper ar https://download.opensuse.org/repositories/science:/machinelearning/openSUSE_Tumbleweed/science:machinelearning.repo &&\
   zypper --gpg-auto-import-keys ref &&\
   zypper install -y libfaiss
COPY --from=build /kowalski/kowalski /kowalski/kowalski

ENV KW_DATABASE=/suseDoc
ENV KW_URL=http://localhost:11434

ENTRYPOINT "/kowalski/kowalski"

FROM  registry.opensuse.org/opensuse/tumbleweed:latest as builder
MAINTAINER "Christian Goll <cgoll@suse.com>"
RUN  zypper ar https://download.opensuse.org/repositories/science:/machinelearning/openSUSE_Tumbleweed/science:machinelearning.repo &&\
   zypper --gpg-auto-import-keys ref &&\
   zypper in -y go git faiss-devel libfaiss findutils
WORKDIR /kowalski
COPY . .
RUN go mod tidy && go mod vendor
RUN go build kowalski.go
# clone the documentation
RUN git clone --depth 1 https://github.com/SUSE/doc-sle.git

ENV KW_DATABASE /suseDoc
ENV KW_URL http://host.docker.internal:11434

# Just add one document to the db
RUN find ./doc-sle -name zypper.xml -type f | xargs ./kowalski database add doc-sle

FROM  registry.opensuse.org/opensuse/tumbleweed:latest as runtime
MAINTAINER "Christian Goll <cgoll@suse.com>"
RUN  zypper ar https://download.opensuse.org/repositories/science:/machinelearning/openSUSE_Tumbleweed/science:machinelearning.repo &&\
   zypper --gpg-auto-import-keys ref &&\
   zypper in -y libfaiss

COPY --from=builder /kowalski/kowalski /kowalski/kowalski
COPY --from=builder /suseDoc /suseDoc
WORKDIR /kowalski
ENV KW_DATABASE /suseDoc
ENV KW_URL http://host.docker.internal:11434
CMD ["/kowalski/kowalski", "chat"]

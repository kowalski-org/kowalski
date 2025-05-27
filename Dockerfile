FROM  ghcr.io/mslacken/kowalski-binary:latest
MAINTAINER "Christian Goll <cgoll@suse.com>"
RUN curl -L https://github.com/kowalski-org/kowalski/releases/download/v0.0.1/suseDoc.tar.gz | tar xz
WORKDIR /kowalski

ENV KW_DATABASE=/suseDoc
ENV KW_URL=http://host.docker.internal:11434

ENTRYPOINT ["/kowalski/kowalski"]

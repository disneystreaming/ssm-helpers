FROM debian:10-slim as downloader

RUN apt update && apt install -y \
    ca-certificates \
    curl \
    python \
    python-pip \
    python-setuptools

# Extract the session-manager-plugin
RUN pip install awscli --upgrade --user \
  && curl -L "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb" -o "session-manager-plugin.deb" \
  && dpkg -x session-manager-plugin.deb /tmp

FROM debian:10-slim

RUN apt update && apt install -y \
    ca-certificates \
    jq \
    locales \
    python \
    python-six \
    tmux \
    && sed -i -e 's/# en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen \
    && locale-gen

ENV LANG=en_US.UTF-8 LANGUAGE=en_US.UTF-8 LC_ALL=en_US.UTF-8

COPY --from=downloader /root/.local/lib /root/.local/lib
COPY --from=downloader /root/.local/bin/aws /usr/local/bin/aws
COPY --from=downloader /tmp/usr/local/sessionmanagerplugin/bin/session-manager-plugin /usr/local/sessionmanagerplugin/bin/session-manager-plugin
RUN ln -s /usr/local/sessionmanagerplugin/bin/session-manager-plugin /usr/local/bin/session-manager-plugin

COPY ssm /usr/local/bin/ssm

ENTRYPOINT ["/usr/local/bin/ssm"]

FROM nginx:1.14-alpine

RUN set -xe; \
    sed -Ei "s/dl-cdn\.alpinelinux\.org/mirrors.tuna.tsinghua.edu.cn/g" /etc/apk/repositories;\
    mkdir apk-cache;\
    apk update --cache-dir apk-cache;\
    apk add make git musl-dev go -t build-deps --cache-dir apk-cache;\
    cd /home;\
    git clone https://github.com/Sunmxt/Starlinks.git;\
    cd Starlinks; \
    make; \
    mv ./bin/starlinks /usr/bin/; \
    apk del -t build-deps;\
    rm -rf /home/Starlinks; \
    rm -rf /apk-cache

COPY webui /home/webui/
COPY docker/nginx.conf.tmpl /home/config/
COPY docker/entrypoint.sh /
RUN set -xe; \
    chmod a+x /entrypoint.sh

ENTRYPOINT /entrypoint.sh

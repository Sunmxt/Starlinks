FROM alpine:3.6-tuna

RUN set -xe; \
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


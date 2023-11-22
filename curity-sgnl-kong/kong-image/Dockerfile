FROM kong:3.0.0-alpine

WORKDIR /usr/kong/

# switch to root to install SGNL plugin and rocks in /usr/local
USER root

RUN mkdir -p /usr/local/share/lua/5.1/kong/plugins/sgnlplugin

COPY sgnl/* /usr/local/share/lua/5.1/kong/plugins/sgnlplugin/
RUN chown -R kong:kong /usr/local/share/lua/5.1/kong/plugins/sgnlplugin

WORKDIR /usr/kong

#install dependencies
RUN apk update && \
  apk add gcc libc-dev && \
  luarocks install lua-cjson && \
  luarocks install lua-resty-jwt

#install phantom-token plugin
RUN git config --global url."https://".insteadOf git:// && \
    git config --global advice.detachedHead false && \
    luarocks install kong-phantom-token 2.0.0

# Prepare directory for the volume, so that proper permissions are applied
RUN mkdir /usr/local/kong/declarative && chown kong:kong /usr/local/kong/declarative

# back to kong user
USER kong

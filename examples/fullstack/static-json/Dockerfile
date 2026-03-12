FROM nginx:alpine

RUN mkdir -p /usr/share/nginx/html/api

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY data/ /usr/share/nginx/html/api/

EXPOSE 80

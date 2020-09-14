FROM debian:buster

RUN apt-get update && \
    apt-get install --no-install-recommends -y \
    python3-flask python3-ldap uwsgi uwsgi-plugin-python3 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /code
COPY . /code/

EXPOSE 8080
CMD /usr/bin/uwsgi --http-socket 0.0.0.0:8080 --master --plugin python3 --module ghostream:app --static-map /static=/code/ghostream/static

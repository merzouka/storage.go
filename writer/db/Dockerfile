FROM postgres:16 AS base
ENV POSTGRES_USER docker
ENV POSTGRES_PASSWORD Kev06122003
ENV POSTGRES_DB metadata

EXPOSE 5432

COPY *.sql /docker-entrypoint-initdb.d/

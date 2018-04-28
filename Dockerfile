FROM ubuntu:16.04

MAINTAINER Kuznetsov Alexander (Dmit)

# Update packages
RUN apt-get -y update


# Installing useful packages
RUN apt-get install -y wget curl git python tree build-essential apt-utils

# Installing Golang
RUN wget https://storage.googleapis.com/golang/go1.9.2.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.9.2.linux-amd64.tar.gz && mkdir go && mkdir go/src && mkdir go/bin && mkdir go/pkg

ENV GOPATH $HOME/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Installing `dep` for managing project's dependencies
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# Installing postgresql
ENV PGVER 9.6

# Download Postgres
RUN echo deb http://apt.postgresql.org/pub/repos/apt/ xenial-pgdg main > /etc/apt/sources.list.d/pgdg.list
RUN wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN apt-get -y update
RUN apt-get install -y postgresql-$PGVER



# Setting up postgres
USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker forum_tp &&\
    /etc/init.d/postgresql stop

RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "synchronous_commit = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "shared_buffers = 512MB" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "autovacuum = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "max_connections = 500" >> /etc/postgresql/$PGVER/main/postgresql.conf

# Expose the PostgreSQL port
EXPOSE 5432

# Add VOLUMEs to allow backup of config, logs and databases
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]


#-------------------------------------------------------------------------------------

# Setting up golang environment
USER root

EXPOSE 5000

WORKDIR $GOPATH/src/github.com/Alex-Kuz/tp-database
ADD . $GOPATH/src/github.com/Alex-Kuz/tp-database/


RUN chmod +x ./scripts/*

RUN ./scripts/build.sh


RUN tree -L 3 ./

# Main command
ENTRYPOINT ["./scripts/start_in_docker.sh"]
CMD [""]

#docker build -t "safestopclient" .
#docker run -p 80:8080 -e SSC_ENV=development --name safestopclient --rm safestopclient

FROM golang:1

#RUN go get github.com/tools/godep
RUN go get gopkg.in/gomail.v2 && go get github.com/gorilla/mux && go get github.com/gorilla/sessions && go get github.com/spf13/viper && go get github.com/golang/dep/cmd/dep && go get github.com/lib/pq && go get -u github.com/pressly/goose/cmd/goose && go get github.com/gorilla/sessions && go get golang.org/x/crypto/bcrypt && go get github.com/gorilla/csrf && go get github.com/jmoiron/sqlx && go get github.com/op/go-logging && go get github.com/gorilla/websocket && go get github.com/stripe/stripe-go
#&& apt-get update && echo "postfix postfix/mailname string safestopclient.com" | debconf-set-selections && echo "postfix postfix/main_mailer_type string 'Internet Site'" | debconf-set-selections && apt-get install -y postfix && service postfix reload && postfix start

ENV SSC_DB_HOST=192.168.0.0
ENV SSC_DB_USERNAME=safestopapp
ENV SSC_DB_NAME=safestopclient

WORKDIR /go/src/github.com/schoolwheels/safestopclient
ADD . /go/src/github.com/schoolwheels/safestopclient
RUN go install github.com/schoolwheels/safestopclient
ENTRYPOINT /go/bin/safestopclient
EXPOSE 8443 8080
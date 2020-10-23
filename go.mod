module github.com/cardiacsociety/web-services

go 1.14

// +heroku goVersion go1.14
// +heroku install ./...

require (
	github.com/34South/envr v0.0.0-20170706023707-e57a7716f427
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/8o8/email v0.2.2
	github.com/aws/aws-sdk-go v1.19.11
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-chi/chi v4.0.2+incompatible // indirect
	github.com/go-playground/locales v0.12.1 // indirect
	github.com/go-playground/universal-translator v0.16.0 // indirect
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gorilla/mux v1.7.1
	github.com/graphql-go/graphql v0.7.8
	github.com/graphql-go/handler v0.2.3
	github.com/hashicorp/go-uuid v1.0.1
	github.com/imdario/mergo v0.3.7
	github.com/jung-kurt/gofpdf v1.0.2
	github.com/kr/pty v1.1.4 // indirect
	github.com/leodido/go-urn v1.1.0 // indirect
	github.com/mailgun/mailgun-go/v3 v3.3.3 // indirect
	github.com/mailru/easyjson v0.0.0-20190403194419-1ea4449da983 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.8.1
	github.com/rogpeppe/go-internal v1.3.0 // indirect
	github.com/rs/cors v1.6.0
	github.com/sendgrid/sendgrid-go v3.4.1+incompatible
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/urfave/negroni v1.0.0
	golang.org/x/crypto v0.0.0-20190411191339-88737f569e3a // indirect
	golang.org/x/net v0.0.0-20190415214537-1da14a5a36f2 // indirect
	golang.org/x/sys v0.0.0-20190415145633-3fd5a3612ccd // indirect
	gopkg.in/go-playground/validator.v9 v9.28.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
)

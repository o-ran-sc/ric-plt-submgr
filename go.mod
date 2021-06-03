module gerrit.o-ran-sc.org/r/ric-plt/submgr

go 1.12

replace gerrit.o-ran-sc.org/r/ric-plt/sdlgo => gerrit.o-ran-sc.org/r/ric-plt/sdlgo.git v0.5.2

replace gerrit.o-ran-sc.org/r/ric-plt/xapp-frame => gerrit.o-ran-sc.org/r/ric-plt/xapp-frame.git v0.8.2

replace gerrit.o-ran-sc.org/r/com/golog => gerrit.o-ran-sc.org/r/com/golog.git v0.0.2

replace gerrit.o-ran-sc.org/r/ric-plt/e2ap => ./e2ap/

require (
	gerrit.o-ran-sc.org/r/ric-plt/e2ap v0.0.0-00010101000000-000000000000
	gerrit.o-ran-sc.org/r/ric-plt/sdlgo v0.5.0
	gerrit.o-ran-sc.org/r/ric-plt/xapp-frame v0.0.0-00010101000000-000000000000
	github.com/go-openapi/errors v0.19.3
	github.com/go-openapi/runtime v0.19.4
	github.com/go-openapi/strfmt v0.19.4
	github.com/go-openapi/swag v0.19.7
	github.com/go-openapi/validate v0.19.6
	github.com/gorilla/mux v1.7.1
	github.com/segmentio/ksuid v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.5.1
)

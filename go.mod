go 1.12

module gerrit.o-ran-sc.org/r/ric-plt/submgr

require (
	gerrit.o-ran-sc.org/r/ric-plt/xapp-frame v0.0.23
	github.com/go-openapi/errors v0.19.2
	github.com/go-openapi/runtime v0.19.7
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/swag v0.19.5
	github.com/go-openapi/validate v0.19.3
	github.com/spf13/viper v1.5.0
)

replace gerrit.o-ran-sc.org/r/ric-plt/sdlgo => gerrit.o-ran-sc.org/r/ric-plt/sdlgo.git v0.2.0

replace gerrit.o-ran-sc.org/r/ric-plt/xapp-frame => gerrit.o-ran-sc.org/r/ric-plt/xapp-frame.git v0.0.23

replace gerrit.o-ran-sc.org/r/com/golog => gerrit.o-ran-sc.org/r/com/golog.git v0.0.1

replace gerrit.o-ran-sc.org/r/ric-plt/e2ap => ./e2ap/
